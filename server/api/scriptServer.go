package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"sync"
)

type ScriptServer struct {
	DB             *sql.DB
	NextId         int64
	cancelChannels map[int64]chan struct{}
	Lock           sync.Mutex
}

var _ ServerInterface = (*ScriptServer)(nil)

func NewScriptServer(db *sql.DB) *ScriptServer {
	return &ScriptServer{
		DB:             db,
		NextId:         1,
		cancelChannels: make(map[int64]chan struct{}),
	}
}

func sendScriptServerError(w http.ResponseWriter, code int, message string) {
	scriptError := Error{
		Code:    int32(code),
		Message: message,
	}
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(scriptError)
	if err != nil {
		return
	}
}

func (s *ScriptServer) GetCommands(w http.ResponseWriter, r *http.Request) {
	s.Lock.Lock()
	ping := s.DB.Ping()
	s.Lock.Unlock()
	if ping != nil {
		log.Println("Problems connecting to the database!")
		return
	}

	//s.Lock.Lock()
	//defer s.Lock.Unlock()
	s.Lock.Lock()
	rows, err := s.DB.Query("SELECT * FROM Script.scripts LIMIT 10")
	s.Lock.Unlock()

	if err != nil {
		log.Println("Failed to find commands!", err)
		return
	}

	var foundCommand Command
	var commands []Command
	for rows.Next() {
		err := rows.Scan(&foundCommand.Id, &foundCommand.BodyScript, &foundCommand.ResultRunScript, &foundCommand.Status)
		if err != nil {
			return
		}
		commands = append(commands, foundCommand)
	}

	response, err2 := json.Marshal(commands)
	if err2 != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (s *ScriptServer) CreateCommand(w http.ResponseWriter, r *http.Request) {
	var newScript Command
	if err := json.NewDecoder(r.Body).Decode(&newScript); err != nil {
		sendScriptServerError(w, http.StatusBadRequest, "Неверный формат для newScript!")
		return
	}

	ping := s.DB.Ping()

	if ping != nil {
		log.Println("Problems connecting to the database!")
		return
	}

	s.Lock.Lock()
	defer s.Lock.Unlock()
	err := s.DB.QueryRow("INSERT INTO Script.scripts(body_script, result_run_script, status) VALUES($1,$2,$3) RETURNING id", newScript.BodyScript, newScript.ResultRunScript, newScript.Status).Scan(&newScript.Id)
	if err != nil {
		log.Println("Failed to insert row:", err)
		return
	}
	s.NextId++

	w.WriteHeader(http.StatusCreated)
	resp := "Новая запись создана под ID: " + strconv.FormatInt(newScript.Id, 10)
	json.NewEncoder(w).Encode(resp)

}

func (s *ScriptServer) ShowCommandById(w http.ResponseWriter, r *http.Request, commandId int64) {
	ping := s.DB.Ping()
	if ping != nil {
		log.Println("Problems connecting to the database!")
		return
	}

	s.Lock.Lock()
	defer s.Lock.Unlock()

	var foundCommand Command
	err := s.DB.QueryRow("SELECT * FROM Script.scripts WHERE id = $1", commandId).Scan(&foundCommand.Id, &foundCommand.BodyScript, &foundCommand.ResultRunScript, &foundCommand.Status)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendScriptServerError(w, http.StatusNotFound, fmt.Sprintf("Не найден скрипт по id %d!", commandId))
		} else {
			log.Println("Failed to find command by id:", err)
		}
		return
	}

	response, err2 := json.Marshal(foundCommand)
	if err2 != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (s *ScriptServer) RunCommandById(w http.ResponseWriter, r *http.Request, commandId int64) {
	var wg sync.WaitGroup

	ping := s.DB.Ping()
	if ping != nil {
		sendScriptServerError(w, http.StatusNotFound, fmt.Sprintf("Ошибка соединения с базой данных!"))
		log.Println("Problems connecting to the database!")
		return
	}

	//s.Lock.Lock()
	//defer s.Lock.Unlock()

	var foundCommand Command
	s.Lock.Lock()
	err := s.DB.QueryRow("SELECT * FROM Script.scripts WHERE id = $1", commandId).Scan(&foundCommand.Id, &foundCommand.BodyScript, &foundCommand.ResultRunScript, &foundCommand.Status)
	s.Lock.Unlock()

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendScriptServerError(w, http.StatusNotFound, fmt.Sprintf("Не найдена команда по id %d!", commandId))
		} else {
			sendScriptServerError(w, http.StatusNotFound, fmt.Sprintf("Ошибка при запросе!"))
			log.Println("Failed to find command by id:", err)
		}
		return
	}

	if foundCommand.Status == InProgress || foundCommand.Status == Ended {
		sendScriptServerError(w, http.StatusOK, fmt.Sprintf("Команда под id %d уже запущена или выполнена. Текущий статус: %s!", commandId, foundCommand.Status))
		log.Println("Команда уже запущена или выполнена:", foundCommand.Status)
		return
	}

	s.Lock.Lock()
	_, err = s.DB.Exec("UPDATE Script.scripts SET status = $1 WHERE id = $2", InProgress, commandId)
	s.Lock.Unlock()

	if err != nil {
		sendScriptServerError(w, http.StatusNotFound, fmt.Sprintf("Ошибка при обновлении статуса!"))
		log.Println("Ошибка при обновлении статуса:", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cancelChan := make(chan struct{})
	s.Lock.Lock()
	s.cancelChannels[commandId] = cancelChan
	s.Lock.Unlock()

	log.Printf("Команда запущена %d!", commandId)
	cmd := exec.CommandContext(ctx, "bash", "-c", foundCommand.BodyScript)
	wg.Add(1)
	go func() {

		defer wg.Done()
		var out bytes.Buffer
		cmd.Stdout = &out
		err = cmd.Run()

		if err != nil {
			log.Println("Ошибка при исполнении команды:", err)
			s.Lock.Lock()
			_, err = s.DB.Exec("UPDATE Script.scripts SET status = $1 WHERE id = $2", Crush, commandId)
			s.Lock.Unlock()
			if err != nil {
				log.Println("Ошибка при обновлении статуса:", err)
			}
			sendScriptServerError(w, http.StatusNotFound, fmt.Sprintf("Ошибка при исполнении команды!"))
			return
		}

		s.Lock.Lock()
		_, err = s.DB.Exec("UPDATE Script.scripts SET result_run_script = $1, status = $2 WHERE id = $3", out.String(), InProgress, commandId)
		s.Lock.Unlock()
		if err != nil {
			sendScriptServerError(w, http.StatusNotFound, fmt.Sprintf("Ошибка при обновлении резульатата выполнения команды!"))
			log.Println("Ошибка при обновлении резульатата выполнения команды:", err)
			return
		}

		s.Lock.Lock()
		_, err = s.DB.Exec("UPDATE Script.scripts SET status = $1 WHERE id = $2", Ended, commandId)
		s.Lock.Unlock()
		if err != nil {
			sendScriptServerError(w, http.StatusNotFound, fmt.Sprintf("Ошибка при обновлении статуса!"))
			log.Println("Ошибка при обновлении статуса:", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		resp := fmt.Sprintf("Команда под ID %d, выполнена и результат записан в базу данных!", commandId)
		json.NewEncoder(w).Encode(resp)

	}()

	select {
	case <-cancelChan:
		wg.Done()
		log.Println("Начало отмены процесса!")
		cmd.Process.Kill()
		s.Lock.Lock()
		_, err = s.DB.Exec("UPDATE Script.scripts SET status = $1 WHERE id = $2", Aborted, commandId)
		s.Lock.Unlock()
		if err != nil {
			sendScriptServerError(w, http.StatusNotFound, fmt.Sprintf("Ошибка при обновлении статуса!"))
			log.Println("Ошибка при обновлении статуса:", err)
		}
		sendScriptServerError(w, http.StatusAccepted, fmt.Sprintf("Выполнение команды под ID %d прерван!", commandId))

	default:
		s.Lock.Lock()
		defer s.Lock.Unlock()
		wg.Wait()
		cancel()
		log.Println("Процесс не был отменен. Резульатат вносится в базу данных!")
		//cancelChan, ok := s.cancelChannels[commandId]
		//if !ok {
		//	log.Println("Канал не найден!")
		//	return
		//}
		close(cancelChan)
		delete(s.cancelChannels, commandId)
	}
	/*
		("UPDATE Script.scripts SET body_script = $1, result_run_script = $2, status = $3 WHERE id = $4",
		        newBodyScript, newResultRunScript, newStatus, id)
	*/

}

func (s *ScriptServer) StopCommandById(w http.ResponseWriter, r *http.Request, commandId int64) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	var foundCommand Command
	err := s.DB.QueryRow("SELECT * FROM Script.scripts WHERE id = $1", commandId).Scan(&foundCommand.Id, &foundCommand.BodyScript, &foundCommand.ResultRunScript, &foundCommand.Status)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendScriptServerError(w, http.StatusNotFound, fmt.Sprintf("Не найдена команда по id %d!", commandId))
		} else {
			sendScriptServerError(w, http.StatusNotFound, fmt.Sprintf("Ошибка при запросе!"))
			log.Println("Ошибка при поиске команды по id:", err)
		}
		return
	}

	if foundCommand.Status != InProgress {
		sendScriptServerError(w, http.StatusNotFound, fmt.Sprintf("Команда под id %d не находится в процессе выполнения!", commandId))
		return
	}

	cancelChan, ok := s.cancelChannels[commandId]
	if !ok {
		sendScriptServerError(w, http.StatusNotFound, fmt.Sprintf("Команда под id %d не находится в процессе выполнения!", commandId))
		log.Println("Канал не найден!")
		return
	}

	cancelChan <- struct{}{}

	w.WriteHeader(http.StatusAccepted)
	resp := fmt.Sprintf("Команда под id %d найдена и запрос на прерывание передан!", commandId)
	json.NewEncoder(w).Encode(resp)
}

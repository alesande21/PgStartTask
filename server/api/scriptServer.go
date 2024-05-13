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

//type Application struct {
//	scriptServer ServerInterface
//}

type DatabaseHandler interface {
	QueryRow(query string, args ...interface{}) (*Command, error)
	Query(query string, args ...interface{}) ([]Command, error)
	Exec(query string, args ...interface{}) error
	Ping() error
}

type ScriptServer struct {
	DB             DatabaseHandler
	NextId         int64
	cancelChannels map[int64]chan struct{}
	commandToRun   chan Command
	Lock           sync.Mutex
}

var _ ServerInterface = (*ScriptServer)(nil)

func NewScriptServer(db DatabaseHandler, commandToRun chan Command) *ScriptServer {
	return &ScriptServer{
		DB:             db,
		NextId:         1,
		cancelChannels: make(map[int64]chan struct{}),
		commandToRun:   commandToRun,
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
	commands, err := s.DB.Query("SELECT * FROM Script.scripts LIMIT 10")
	s.Lock.Unlock()

	if err != nil {
		return
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

	s.Lock.Lock()
	defer s.Lock.Unlock()
	command, err := s.DB.QueryRow("INSERT INTO Script.scripts(body_script, result_run_script, status) VALUES($1,$2,$3) RETURNING id, body_script, result_run_script, status", newScript.BodyScript, newScript.ResultRunScript, newScript.Status)

	if err != nil {
		log.Println("Failed to insert row:", err)
		return
	}
	s.NextId++

	w.WriteHeader(http.StatusCreated)
	resp := "Новая запись создана под ID: " + strconv.FormatInt(command.Id, 10)
	json.NewEncoder(w).Encode(resp)

}

func (s *ScriptServer) ShowCommandById(w http.ResponseWriter, r *http.Request, commandId int64) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	foundCommand, err := s.DB.QueryRow("SELECT * FROM Script.scripts WHERE id = $1", commandId)

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
	//var wg sync.WaitGroup

	ping := s.DB.Ping()
	if ping != nil {
		sendScriptServerError(w, http.StatusNotFound, fmt.Sprintf("Ошибка соединения с базой данных!"))
		log.Println("Проблема с подключением к базе данных!")
		return
	}

	//var foundCommand Command
	s.Lock.Lock()
	foundCommand, err := s.DB.QueryRow("SELECT * FROM Script.scripts WHERE id = $1", commandId)
	//err := s.DB.QueryRow("SELECT * FROM Script.scripts WHERE id = $1", commandId).Scan(&foundCommand.Id, &foundCommand.BodyScript, &foundCommand.ResultRunScript, &foundCommand.Status)
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
	err = s.DB.Exec("UPDATE Script.scripts SET status = $1 WHERE id = $2", InProgress, commandId)
	//_, err = s.DB.Exec("UPDATE Script.scripts SET status = $1 WHERE id = $2", InProgress, commandId)
	s.Lock.Unlock()

	if err != nil {
		sendScriptServerError(w, http.StatusNotFound, fmt.Sprintf("Ошибка при обновлении статуса!"))
		log.Println("Ошибка при обновлении статуса:", err)
		return
	}

	s.commandToRun <- *foundCommand

	w.WriteHeader(http.StatusAccepted)
	resp := fmt.Sprintf("Команда под ID %d запущена!", commandId)
	json.NewEncoder(w).Encode(resp)

}

func (s *ScriptServer) StopCommandById(w http.ResponseWriter, r *http.Request, commandId int64) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	//var foundCommand Command
	//err := s.DB.QueryRow("SELECT * FROM Script.scripts WHERE id = $1", commandId).Scan(&foundCommand.Id, &foundCommand.BodyScript, &foundCommand.ResultRunScript, &foundCommand.Status)
	foundCommand, err := s.DB.QueryRow("SELECT * FROM Script.scripts WHERE id = $1", commandId)

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
	resp := fmt.Sprintf("Команда c id %d найдена и запрос на прерывание передан!", commandId)
	json.NewEncoder(w).Encode(resp)
}

func (s *ScriptServer) RunCommand(command *Command) {
	ctx, cancel := context.WithCancel(context.Background())

	cancelChan := make(chan struct{})
	s.Lock.Lock()
	s.cancelChannels[command.Id] = cancelChan
	s.Lock.Unlock()

	var statusCancel bool
	statusCancel = false

	cmd := exec.CommandContext(ctx, "bash", "-c", command.BodyScript)
	var out bytes.Buffer
	cmd.Stdout = &out
	log.Printf("Команда c ID %d запущена!", command.Id)

	go func() {
		<-cancelChan
		cancel()
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		s.Lock.Lock()
		err := s.DB.Exec("UPDATE Script.scripts SET status = $1 WHERE id = $2", Aborted, command.Id)
		//_, err := s.DB.Exec("UPDATE Script.scripts SET status = $1 WHERE id = $2", Aborted, command.Id)
		s.Lock.Unlock()
		if err != nil {
			log.Println("Ошибка при обновлении статуса:", err)
		}
		log.Printf("Команда c ID %d отменена!", command.Id)
		statusCancel = true
		close(cancelChan)
		delete(s.cancelChannels, command.Id)
	}()

	//go func() {
	err := cmd.Run()
	if err != nil {
		if !statusCancel {
			log.Println("Ошибка при исполнении команды или команда была прервана:", err)
			s.Lock.Lock()
			err = s.DB.Exec("UPDATE Script.scripts SET status = $1 WHERE id = $2", Crush, command.Id)
			//_, err = s.DB.Exec("UPDATE Script.scripts SET status = $1 WHERE id = $2", Crush, command.Id)
			s.Lock.Unlock()
			if err != nil {
				log.Println("Ошибка при обновлении статуса(Crush) в базе данных:", err)
			}
		}
		cancel()
		return
	}

	s.Lock.Lock()
	err = s.DB.Exec("UPDATE Script.scripts SET result_run_script = $1, status = $2 WHERE id = $3", out.String(), Ended, command.Id)
	//_, err = s.DB.Exec("UPDATE Script.scripts SET result_run_script = $1, status = $2 WHERE id = $3", out.String(), Ended, command.Id)
	s.Lock.Unlock()
	if err != nil {
		//sendScriptServerError(w, http.StatusNotFound, fmt.Sprintf("Ошибка при обновлении резульатата выполнения команды!"))
		log.Printf("Ошибка при внесении в базу данных результата выполнение скрипта под id %d: %s", command.Id, err)
		return
	}
	//}()

	cancel()
	log.Printf("Команды под id %d выполнена. Результат вносится в базу данных!", command.Id)
	//close(cancelChan)
	//delete(s.cancelChannels, command.Id)
}

//func (s *ScriptServer) AwaitingCancel(cancelChan chan struct{}) {
//
//}

func ControlRunningCommand(server *ScriptServer) {
	for {
		select {
		case commandId := <-server.commandToRun:
			go server.RunCommand(&commandId)
		}
	}
}

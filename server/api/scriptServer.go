package api

import (
	"bytes"
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
	DB     *sql.DB
	NextId int64
	Lock   sync.Mutex
}

var _ ServerInterface = (*ScriptServer)(nil)

func NewScriptServer(db *sql.DB) *ScriptServer {
	return &ScriptServer{
		DB:     db,
		NextId: 1,
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
	ping := s.DB.Ping()
	if ping != nil {
		log.Println("Problems connecting to the database!")
		return
	}

	s.Lock.Lock()
	defer s.Lock.Unlock()

	rows, err := s.DB.Query("SELECT * FROM Script.scripts LIMIT 10")

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
	ping := s.DB.Ping()
	if ping != nil {
		sendScriptServerError(w, http.StatusNotFound, fmt.Sprintf("Ошибка соединения с базой данных!"))
		log.Println("Problems connecting to the database!")
		return
	}

	s.Lock.Lock()
	defer s.Lock.Unlock()

	var foundCommand Command
	err := s.DB.QueryRow("SELECT * FROM Script.scripts WHERE id = $1", commandId).Scan(&foundCommand.Id, &foundCommand.BodyScript, &foundCommand.ResultRunScript, &foundCommand.Status)

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

	_, err = s.DB.Exec("UPDATE Script.scripts SET status = $1 WHERE id = $2", InProgress, commandId)
	if err != nil {
		sendScriptServerError(w, http.StatusNotFound, fmt.Sprintf("Ошибка при обновлении статуса!"))
		log.Println("Ошибка при обновлении статуса:", err)
		return
	}

	cmd := exec.Command("bash", "-c", foundCommand.BodyScript)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		log.Println("Ошибка при исполнении команды:", err)
		_, err = s.DB.Exec("UPDATE Script.scripts SET status = $1 WHERE id = $2", Crush, commandId)
		if err != nil {
			log.Println("Ошибка при обновлении статуса:", err)
		}
		sendScriptServerError(w, http.StatusNotFound, fmt.Sprintf("Ошибка при исполнении команды!"))
		return
	}

	_, err = s.DB.Exec("UPDATE Script.scripts SET result_run_script = $1, status = $2 WHERE id = $3", out.String(), InProgress, commandId)
	if err != nil {
		sendScriptServerError(w, http.StatusNotFound, fmt.Sprintf("Ошибка при обновлении резульатата выполнения команды!"))
		log.Println("Ошибка при обновлении резульатата выполнения команды:", err)
		return
	}

	_, err = s.DB.Exec("UPDATE Script.scripts SET status = $1 WHERE id = $2", Ended, commandId)
	if err != nil {
		sendScriptServerError(w, http.StatusNotFound, fmt.Sprintf("Ошибка при обновлении статуса!"))
		log.Println("Ошибка при обновлении статуса:", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	resp := fmt.Sprintf("Команда под ID %d, выполнена и результат записан в базу данных!", commandId)
	json.NewEncoder(w).Encode(resp)

	/*
		("UPDATE Script.scripts SET body_script = $1, result_run_script = $2, status = $3 WHERE id = $4",
		        newBodyScript, newResultRunScript, newStatus, id)
	*/

}

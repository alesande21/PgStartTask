package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	s.Lock.Lock()
	defer s.Lock.Unlock()

	commands := Command{Id: 1, BodyScript: "Kuku", ResultRunScript: "", Status: New}

	s.NextId++

	response, err := json.Marshal(commands)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	//db.ConnectionToDB()

	w.Header().Set("Content-Type", "appication/json")
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
	fmt.Println("TUTUTUTUTU")
	err := s.DB.QueryRow("INSERT INTO Script.scripts(body_script, result_run_script, status) VALUES($1,$2,$3) RETURNING id", newScript.BodyScript, newScript.ResultRunScript, newScript.Status).Scan(&newScript.Id)
	if err != nil {
		log.Println("Failed to insert row:", err)
		return
	}
	s.NextId++

	w.WriteHeader(http.StatusCreated)
	resp := "Новая запис создана под ID: " + strconv.FormatInt(newScript.Id, 10)
	json.NewEncoder(w).Encode(resp)

}

func (s *ScriptServer) ShowCommandById(w http.ResponseWriter, r *http.Request, commandId string) {
	// Здесь реализация метода ShowCommandById
}

func (s *ScriptServer) RunCommandById(w http.ResponseWriter, r *http.Request, commandId string) {
	// Здесь реализация метода RunCommandById
}

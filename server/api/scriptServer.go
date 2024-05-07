package api

import (
	"PgStartTask/db"
	"encoding/json"
	"net/http"
	"sync"
)

type ScriptServer struct {
	Scripts map[int64]Command
	NextId  int64
	Lock    sync.Mutex
}

var _ ServerInterface = (*ScriptServer)(nil)

func NewScriptServer() *ScriptServer {
	return &ScriptServer{
		Scripts: make(map[int64]Command),
		NextId:  1000,
	}
}

func (s *ScriptServer) GetCommands(w http.ResponseWriter, r *http.Request) {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	commands1 := Command{Id: 1, BodyScript: "Kuku", ResultRunScript: "", Status: New}

	commands2 := Command{Id: 1, BodyScript: "Kuku2", ResultRunScript: "", Status: Aborted}

	s.Scripts[s.NextId] = commands1
	s.NextId++
	s.Scripts[s.NextId] = commands2
	s.NextId++

	response, err := json.Marshal(s.Scripts)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

	db.ConnectionToDB()

	w.Header().Set("Content-Type", "appication/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (s *ScriptServer) CreateCommand(w http.ResponseWriter, r *http.Request) {
	// Здесь реализация метода CreateCommand
}

func (s *ScriptServer) ShowCommandById(w http.ResponseWriter, r *http.Request, commandId string) {
	// Здесь реализация метода ShowCommandById
}

func (s *ScriptServer) RunCommandById(w http.ResponseWriter, r *http.Request, commandId string) {
	// Здесь реализация метода RunCommandById
}

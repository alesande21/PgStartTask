package mocks

import (
	"PgStartTask/server/api"
	"errors"
)

type MockDB struct {
	commands map[int64]api.Command
	nextID   int64
}

func (r *MockDB) Query(query string, args ...interface{}) ([]api.Command, error) {

	mockCommands := []api.Command{
		{Id: 1, BodyScript: "script1", ResultRunScript: "result1", Status: "Pending"},
		{Id: 2, BodyScript: "script2", ResultRunScript: "result2", Status: "InProgress"},
	}

	return mockCommands, nil
}

func (r *MockDB) Ping() error {
	return nil
}

func (r *MockDB) QueryRow(query string, args ...interface{}) (*api.Command, error) {

	expectedCommand := &api.Command{
		Id:              1,
		BodyScript:      "test_script",
		ResultRunScript: "test_result",
		Status:          "Pending",
	}

	return expectedCommand, nil
}

func (r *MockDB) Exec(query string, args ...interface{}) error {

	return errors.New("expected error")
}

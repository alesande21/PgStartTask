package mocks

import (
	"PgStartTask/server/api"
	"errors"
	"strings"
)

type MockDB struct {
	commands map[int64]*api.Command
	nextID   int64
}

func NewMockDB() *MockDB {
	return &MockDB{
		commands: make(map[int64]*api.Command),
		nextID:   1,
	}
}

func (r *MockDB) Query(query string, args ...interface{}) (api.Commands, error) {
	var mockCommands api.Commands
	if len(query) == 0 {
		return nil, errors.New("пустой запрос")
	}

	for i := 0; i < len(r.commands); i++ {
		if i >= 10 {
			break
		}

		command := r.commands[int64(i)]
		mockCommands = append(mockCommands, *command)
	}

	return mockCommands, nil
}

func (r *MockDB) Ping() error {
	return nil
}

func (r *MockDB) QueryRow(query string, args ...interface{}) (*api.Command, error) {

	queryWords := strings.Fields(query)
	if len(queryWords) == 0 {
		return nil, errors.New("пустой запрос")
	}

	firstWord := queryWords[0]

	switch strings.ToUpper(firstWord) {
	case "SELECT":
		if len(args) != 1 {
			return nil, errors.New("недостаточно параметров для select")
		}
		id, ok := args[0].(int64)
		if !ok {
			return nil, errors.New("неверный тип аргумента для SELECT")
		}
		command, found := r.commands[id]

		if !found {
			return nil, errors.New("элемент не найден SELECT")
		}

		return command, nil

	case "INSERT":

		if len(args) != 3 {
			return nil, errors.New("недостаточно параметров для insert")
		}
		bodyScript, ok1 := args[0].(string)
		resultRunScript, ok2 := args[1].(string)
		status, ok3 := args[2].(api.CommandStatus)

		if !ok1 || !ok2 || !ok3 {
			return nil, errors.New("неверный тип аргумента для INSERT")
		}

		command := &api.Command{
			Id:              r.nextID,
			BodyScript:      bodyScript,
			ResultRunScript: resultRunScript,
			Status:          status,
		}

		r.commands[r.nextID] = command
		r.nextID++

		return command, nil
	default:
		return nil, errors.New("неизвестная операция")
	}

	return nil, errors.New("неизвестная операция")
}

func (r *MockDB) Exec(query string, args ...interface{}) error {

	if len(args) != 3 && len(args) != 2 {
		return errors.New("недостаточно параметров для insert")
	}

	if len(args) == 2 {
		status, ok1 := args[0].(api.CommandStatus)
		id, ok2 := args[1].(int64)
		if !ok1 || !ok2 {
			return errors.New("неверный тип аргумента для exec")
		}

		cmd, ok := r.commands[id]
		if !ok {
			return errors.New("нет команды под id")
		}
		cmd.Status = status

		return nil
	}

	if len(args) == 3 {
		resScript, ok1 := args[0].(string)
		status, ok2 := args[2].(api.CommandStatus)
		id, ok3 := args[1].(int64)
		if !ok1 || !ok2 || !ok3 {
			return errors.New("неверный тип аргумента для exec")
		}
		cmd, ok := r.commands[id]
		if !ok {
			return errors.New("нет команды под id")
		}
		cmd.Status = status
		cmd.ResultRunScript = resScript

		return nil

	}

	return errors.New("unexpected error")
}

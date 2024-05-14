package tests

import (
	"PgStartTask/server/api"
	"PgStartTask/tests/mocks"
	"encoding/json"
	"fmt"
	middleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func doGet(t *testing.T, mux *mux.Router, url string) *httptest.ResponseRecorder {
	response := testutil.NewRequest().Get(url).WithAcceptJson().GoWithHTTPHandler(t, mux)
	return response.Recorder
}

func TestCommandServer(t *testing.T) {
	var err error

	swagger, err := api.GetSwagger()
	require.NoError(t, err)
	swagger.Servers = nil

	mockDB := mocks.NewMockDB()
	commandToRun := make(chan api.Command)
	scriptServer := api.NewScriptServer(mockDB, commandToRun)

	go api.ControlRunningCommand(scriptServer)

	r := mux.NewRouter()

	r.Use(middleware.OapiRequestValidator(swagger))

	api.HandlerFromMux(scriptServer, r)

	t.Run("Добавить команду1", func(t *testing.T) {
		newCommand := api.Command{
			BodyScript:      "echo 'Hello, World!'",
			Id:              1,
			ResultRunScript: "",
			Status:          "new",
		}

		rr := testutil.NewRequest().Post("/commands").WithJsonBody(newCommand).GoWithHTTPHandler(t, r).Recorder
		assert.Equal(t, http.StatusCreated, rr.Code)
		fmt.Println(rr)
		str := fmt.Sprintf("\"Новая запись создана под ID: %d\"\n", newCommand.Id)
		assert.Equal(t, rr.Body.String(), str)
	})

	t.Run("Добавить команду2", func(t *testing.T) {
		newCommand := api.Command{
			BodyScript:      "echo 'Hello, World!'",
			Id:              2,
			ResultRunScript: "",
			Status:          "new",
		}

		rr := testutil.NewRequest().Post("/commands").WithJsonBody(newCommand).GoWithHTTPHandler(t, r).Recorder
		assert.Equal(t, http.StatusCreated, rr.Code)

		str := fmt.Sprintf("\"Новая запись создана под ID: %d\"\n", newCommand.Id)
		assert.Equal(t, rr.Body.String(), str)
	})

	t.Run("Добавить команду3", func(t *testing.T) {
		newCommand := api.Command{
			BodyScript:      "echo 'Hello, World!'",
			Id:              3,
			ResultRunScript: "",
			Status:          "new",
		}

		rr := testutil.NewRequest().Post("/commands").WithJsonBody(newCommand).GoWithHTTPHandler(t, r).Recorder
		assert.Equal(t, http.StatusCreated, rr.Code)

		str := fmt.Sprintf("\"Новая запись создана под ID: %d\"\n", newCommand.Id)
		assert.Equal(t, rr.Body.String(), str)
	})

	t.Run("Найти команду1", func(t *testing.T) {
		command := api.Command{
			BodyScript:      "echo 'Hello, World!'",
			Id:              1,
			ResultRunScript: "",
			Status:          "new",
		}

		rr := testutil.NewRequest().Get("/commands/1").WithJsonBody(command).GoWithHTTPHandler(t, r).Recorder
		assert.Equal(t, http.StatusOK, rr.Code)

		var resultCommand api.Command
		err = json.NewDecoder(rr.Body).Decode(&resultCommand)

		assert.Equal(t, command.Id, resultCommand.Id)
		assert.Equal(t, command.BodyScript, resultCommand.BodyScript)
		assert.Equal(t, command.ResultRunScript, resultCommand.ResultRunScript)
		assert.Equal(t, command.Status, resultCommand.Status)

	})

	t.Run("Не найти команду70", func(t *testing.T) {
		command := api.Command{
			BodyScript:      "echo 'Hello, World!'",
			Id:              70,
			ResultRunScript: "",
			Status:          "new",
		}

		rr := testutil.NewRequest().Get("/commands/70").WithJsonBody(command).GoWithHTTPHandler(t, r).Recorder
		assert.Equal(t, http.StatusNotFound, rr.Code)

	})

	t.Run("Запустить команду и остановмит команду1", func(t *testing.T) {
		command := api.Command{
			BodyScript:      "echo 'Hello, World!'",
			Id:              1,
			ResultRunScript: "",
			Status:          "aborted",
		}

		rr := testutil.NewRequest().Post("/commands/1/run").WithJsonBody(command).GoWithHTTPHandler(t, r).Recorder
		assert.Equal(t, http.StatusAccepted, rr.Code)

		fmt.Println(rr)
		time.Sleep(5)

	})

}

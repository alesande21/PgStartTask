// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen/v2 version v2.0.0 DO NOT EDIT.
package api

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
	"github.com/oapi-codegen/runtime"
)

// Defines values for CommandStatus.
const (
	Aborted    CommandStatus = "aborted"
	Crush      CommandStatus = "crush"
	Ended      CommandStatus = "ended"
	InProgress CommandStatus = "in_progress"
	New        CommandStatus = "new"
)

// Command defines model for Command.
type Command struct {
	BodyScript      string        `json:"body_script"`
	Id              int64         `json:"id"`
	ResultRunScript string        `json:"result_run_script"`
	Status          CommandStatus `json:"status"`
}

// CommandStatus defines model for Command.Status.
type CommandStatus string

// Commands defines model for Commands.
type Commands = []Command

// Error defines model for Error.
type Error struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

// StatusMessage defines model for StatusMessage.
type StatusMessage struct {
	Message string `json:"message"`
}

// CreateCommandJSONRequestBody defines body for CreateCommand for application/json ContentType.
type CreateCommandJSONRequestBody = Command

// ServerInterface represents all server handlers.
type ServerInterface interface {

	// (GET /commands)
	GetCommands(w http.ResponseWriter, r *http.Request)
	// Создание и добавление команды
	// (POST /commands)
	CreateCommand(w http.ResponseWriter, r *http.Request)
	// Получение команды по идентификатору
	// (GET /commands/{command_id})
	ShowCommandById(w http.ResponseWriter, r *http.Request, commandId int64)
	// Запуск команды по идентификатору
	// (POST /commands/{command_id}/run)
	RunCommandById(w http.ResponseWriter, r *http.Request, commandId int64)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.Handler) http.Handler

// GetCommands operation middleware
func (siw *ServerInterfaceWrapper) GetCommands(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetCommands(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// CreateCommand operation middleware
func (siw *ServerInterfaceWrapper) CreateCommand(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.CreateCommand(w, r)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// ShowCommandById operation middleware
func (siw *ServerInterfaceWrapper) ShowCommandById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "command_id" -------------
	var commandId int64

	err = runtime.BindStyledParameter("simple", false, "command_id", mux.Vars(r)["command_id"], &commandId)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "command_id", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.ShowCommandById(w, r, commandId)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// RunCommandById operation middleware
func (siw *ServerInterfaceWrapper) RunCommandById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "command_id" -------------
	var commandId int64

	err = runtime.BindStyledParameter("simple", false, "command_id", mux.Vars(r)["command_id"], &commandId)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "command_id", Err: err})
		return
	}

	handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.RunCommandById(w, r, commandId)
	}))

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

type UnescapedCookieParamError struct {
	ParamName string
	Err       error
}

func (e *UnescapedCookieParamError) Error() string {
	return fmt.Sprintf("error unescaping cookie parameter '%s'", e.ParamName)
}

func (e *UnescapedCookieParamError) Unwrap() error {
	return e.Err
}

type UnmarshalingParamError struct {
	ParamName string
	Err       error
}

func (e *UnmarshalingParamError) Error() string {
	return fmt.Sprintf("Error unmarshaling parameter %s as JSON: %s", e.ParamName, e.Err.Error())
}

func (e *UnmarshalingParamError) Unwrap() error {
	return e.Err
}

type RequiredParamError struct {
	ParamName string
}

func (e *RequiredParamError) Error() string {
	return fmt.Sprintf("Query argument %s is required, but not found", e.ParamName)
}

type RequiredHeaderError struct {
	ParamName string
	Err       error
}

func (e *RequiredHeaderError) Error() string {
	return fmt.Sprintf("Header parameter %s is required, but not found", e.ParamName)
}

func (e *RequiredHeaderError) Unwrap() error {
	return e.Err
}

type InvalidParamFormatError struct {
	ParamName string
	Err       error
}

func (e *InvalidParamFormatError) Error() string {
	return fmt.Sprintf("Invalid format for parameter %s: %s", e.ParamName, e.Err.Error())
}

func (e *InvalidParamFormatError) Unwrap() error {
	return e.Err
}

type TooManyValuesForParamError struct {
	ParamName string
	Count     int
}

func (e *TooManyValuesForParamError) Error() string {
	return fmt.Sprintf("Expected one value for %s, got %d", e.ParamName, e.Count)
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, GorillaServerOptions{})
}

type GorillaServerOptions struct {
	BaseURL          string
	BaseRouter       *mux.Router
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r *mux.Router) http.Handler {
	return HandlerWithOptions(si, GorillaServerOptions{
		BaseRouter: r,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, r *mux.Router, baseURL string) http.Handler {
	return HandlerWithOptions(si, GorillaServerOptions{
		BaseURL:    baseURL,
		BaseRouter: r,
	})
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options GorillaServerOptions) http.Handler {
	r := options.BaseRouter

	if r == nil {
		r = mux.NewRouter()
	}
	if options.ErrorHandlerFunc == nil {
		options.ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandlerFunc:   options.ErrorHandlerFunc,
	}

	r.HandleFunc(options.BaseURL+"/commands", wrapper.GetCommands).Methods("GET")

	r.HandleFunc(options.BaseURL+"/commands", wrapper.CreateCommand).Methods("POST")

	r.HandleFunc(options.BaseURL+"/commands/{command_id}", wrapper.ShowCommandById).Methods("GET")

	r.HandleFunc(options.BaseURL+"/commands/{command_id}/run", wrapper.RunCommandById).Methods("POST")

	return r
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/+xVzW4jRRB+lVHBcdb2bhCHOWaF0B64sNxQZHVmynavPN1Dd08Wy7K0GyMRKQ+AxAGE",
	"OHCdRbHWcbDzCtVvhLpnbMf25I9E4bKXqJ2p/uqrr76qHkIs00wKFEZDNAQd9zBl/vhSpikTiTtmSmao",
	"DEf/4VAmg7aOFc+M+2kGGUIE2iguujAKgfs7HalSZiACLsyXX0C4jOPCYBeVC1So875pq1zcBKcNM7lP",
	"jCJPIfoeBL6FELhoZ0p2FWoNIaBIMIEQYpXrHoTADqUymMBBuI3oE/+Qc4WJA+Pu1tWS6nitWKzh5OEb",
	"jI0jWAnlKXKDqT98rrADEXzWXOvbrMRtLpUdrcCYUmzgfn+llFS7kscywW1R917Uipqi1qyLNVJuFe4x",
	"1/F1lb32RX+zRtxkdedU1+dwkVx0pMfgpu++7TPde1YKH2hUR6gghCNUmksBETxvtBotx05mKFjGIYK9",
	"RquxByFkzPQ8M6f6qidd9MZyzJnhUrxKIIKv0az65jueSaHLsl60WqXmwqDwV1mW9XnsLzffaMdiOSn3",
	"7bnebbqTIMGy3rJC+pUW9A8VNKczexrQnAo6pzOa0Nyegg/vsLxv7sXyJnKl6+qY/GZPaEofaEZFQJe0",
	"COgjFXRp39HCvrfjgGZrqiVAJnWN3C8VMoNL45fuQG32ZTJ4tCpWY7VpP6NyHO20+Pmjpd0cklvaSUVg",
	"x/Y9XdLEntDcCXpGC/pABf1NF67DVDxph3OBP2YYG0wCrGJC0HmaMjVw3P+gBX10vGlOU5oENN1l7P8/",
	"u+pZt5lYV5drpjL+gUNeDWZzWJ3aPBldO6Wve/Jt1db9wavEj7hiKRpUDny4Vcx3PQx4EshOYHoYVAkC",
	"IwOFRnE8Qv9sQOQXBYQgWOoGcc0Eto0TXtH31hdtdPDATXJnd2+W/f+uBLqwY/vzNUaolsa02l7HNLU/",
	"0dTdtce0sO/seNtwvz8c8X7ma6rc61K/uL7NxScD3nHP/blebPaUzpdvhdt3s//gi6c09V8u7TP3xzOz",
	"J3ThjnOaBPaYCsffPXrTjZoewe2/PEShep+PRv8GAAD//3iGYwfVCwAA",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}

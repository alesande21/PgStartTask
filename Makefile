install:
	go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@latest
	go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@v2.0.0

genClient:

genServer:
	oapi-codegen --config=configs/oapi/configServer.yaml api/scriptAPI.yml

runDefault:
	go run ./server/server.go

runCustom:
	go run ./server/server.go -ip 127.0.0.1 -port 8080

buildDefault:
	go build ./server/server.go

buildCustom:
	go build -o build/server ./server

startServer:
	./build/server

test:
	go test ./tests


#	//go:generate oapi-codegen --config=config.yaml https://petstore3.swagger.io/api/v3/openapi.json

install:
	go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@latest
	go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@v2.0.0


genClient:


genServer:
	oapi-codegen --config=configs/oapi/configServer.yaml api/scriptAPI.yml

#	//go:generate oapi-codegen --config=config.yaml https://petstore3.swagger.io/api/v3/openapi.json

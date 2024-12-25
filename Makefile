TAG := 22d49d5b87a4d5fc60a194dab02dd1d71529687f

schema-update:
	curl https://raw.githubusercontent.com/tdlib/td/${TAG}/td/generate/scheme/td_api.tl 2>/dev/null > ./data/td_api.tl

generate-json:
	go run ./cmd/generateJson/main.go \
		-version "${TAG}" \
		-output "./data/td_api.json"

generate-code:
	go run ./cmd/generateCode/main.go \
		-version "${TAG}" \
		-outputDir "./client" \
		-package client \
		-functionFile function_generated.go \
		-typeFile type_generated.go \
		-unmarshalerFile unmarshaler_generated.go
	go fmt ./...

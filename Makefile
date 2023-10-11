.DEFAULT_GOAL = all

REV := 986f1ab469b9bbff2b95850cc4485e16798a26b7

all: schema-update generate-json generate-code format
.PHONY: all

schema-update:
	curl -sSf https://raw.githubusercontent.com/tdlib/td/${REV}/td/generate/scheme/td_api.tl > ./data/td_api.tl
.PHONY: schema-update

generate-json:
	go run ./cmd/generate-json.go \
		-version ${REV} \
		-output ./data/td_api.json
.PHONY: generate-json

generate-code:
	go run ./cmd/generate-code.go \
		-version ${REV} \
		-outputDir ./client \
		-package client \
		-functionFile function.go \
		-typeFile type.go \
		-unmarshalerFile unmarshaler.go
.PHONY: generate-code

format:
	goimports -local $$(go list -m) -w ./..
.PHONY: format

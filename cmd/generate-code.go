package main

import (
	"bufio"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/zelenin/go-tdlib/codegen"
	"github.com/zelenin/go-tdlib/tlparser"
)

type config struct {
	version             string
	outputDirPath       string
	packageName         string
	functionFileName    string
	typeFileName        string
	unmarshalerFileName string
}

func main() {
	var config config

	flag.StringVar(&config.version, "version", "", "TDLib version")
	flag.StringVar(&config.outputDirPath, "outputDir", "./tdlib", "output directory")
	flag.StringVar(&config.packageName, "package", "tdlib", "package name")
	flag.StringVar(&config.functionFileName, "functionFile", "function.go", "functions filename")
	flag.StringVar(&config.typeFileName, "typeFile", "type.go", "types filename")
	flag.StringVar(&config.unmarshalerFileName, "unmarshalerFile", "unmarshaler.go", "unmarshalers filename")

	flag.Parse()

	resp, err := http.Get("https://raw.githubusercontent.com/tdlib/td/" + config.version + "/td/generate/scheme/td_api.tl")
	if err != nil {
		log.Fatalf("http.Get error: %s", err)
		return
	}
	defer resp.Body.Close()

	schema, err := tlparser.Parse(resp.Body)
	if err != nil {
		log.Fatalf("schema parse error: %s", err)
		return
	}

	err = os.MkdirAll(config.outputDirPath, 0755)
	if err != nil {
		log.Fatalf("error creating %s: %s", config.outputDirPath, err)
	}

	functionFilePath := filepath.Join(config.outputDirPath, config.functionFileName)

	os.Remove(functionFilePath)
	functionFile, err := os.OpenFile(functionFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		log.Fatalf("functionFile open error: %s", err)
	}
	defer functionFile.Close()

	bufio.NewWriter(functionFile).Write(codegen.GenerateFunctions(schema, config.packageName))

	typeFilePath := filepath.Join(config.outputDirPath, config.typeFileName)

	os.Remove(typeFilePath)
	typeFile, err := os.OpenFile(typeFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		log.Fatalf("typeFile open error: %s", err)
	}
	defer typeFile.Close()

	bufio.NewWriter(typeFile).Write(codegen.GenerateTypes(schema, config.packageName))

	unmarshalerFilePath := filepath.Join(config.outputDirPath, config.unmarshalerFileName)

	os.Remove(unmarshalerFilePath)
	unmarshalerFile, err := os.OpenFile(unmarshalerFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		log.Fatalf("unmarshalerFile open error: %s", err)
	}
	defer unmarshalerFile.Close()

	bufio.NewWriter(unmarshalerFile).Write(codegen.GenerateUnmarshalers(schema, config.packageName))
}

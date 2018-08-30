package main

import (
    "bufio"
    "flag"
    "log"
    "os"
    "path/filepath"

    "github.com/zelenin/go-tdlib/tlparser"
    "github.com/zelenin/go-tdlib/codegen"
)

type config struct {
    schemaFilePath      string
    outputDirPath       string
    packageName         string
    functionFileName    string
    typeFileName        string
    unmarshalerFileName string
}

func main() {
    var config config

    flag.StringVar(&config.schemaFilePath, "schema", "./td_api.tl", ".tl schema file")
    flag.StringVar(&config.outputDirPath, "outputDir", "./tdlib", "output directory")
    flag.StringVar(&config.packageName, "package", "tdlib", "package name")
    flag.StringVar(&config.functionFileName, "functionFile", "function.go", "functions filename")
    flag.StringVar(&config.typeFileName, "typeFile", "type.go", "types filename")
    flag.StringVar(&config.unmarshalerFileName, "unmarshalerFile", "unmarshaler.go", "unmarshalers filename")

    flag.Parse()

    schemaFile, err := os.OpenFile(config.schemaFilePath, os.O_RDONLY, os.ModePerm)
    if err != nil {
        log.Fatalf("schemaFile open error: %s", err)
    }
    defer schemaFile.Close()

    schema, err := tlparser.Parse(schemaFile)
    if err != nil {
        log.Fatalf("schema parse error: %s", err)
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

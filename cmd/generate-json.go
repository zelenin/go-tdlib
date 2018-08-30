package main

import (
    "bufio"
    "encoding/json"
    "flag"
    "log"
    "os"
    "path/filepath"
    "strings"
    "github.com/zelenin/go-tdlib/tlparser"
)

func main() {
    var inputFilePath string
    var outputFilePath string

    flag.StringVar(&inputFilePath, "input", "./td_api.tl", "tl schema file")
    flag.StringVar(&outputFilePath, "output", "./td_api.json", "json schema file")

    flag.Parse()

    file, err := os.OpenFile(inputFilePath, os.O_RDONLY, os.ModePerm)
    if err != nil {
        log.Fatalf("open file error: %s", err)
        return
    }
    defer file.Close()

    schema, err := tlparser.Parse(file)
    if err != nil {
        log.Fatalf("schema parse error: %s", err)
        return
    }

    err = os.MkdirAll(filepath.Dir(outputFilePath), os.ModePerm)
    if err != nil {
        log.Fatalf("make dir error: %s", filepath.Dir(outputFilePath))
    }

    file, err = os.OpenFile(outputFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
    if err != nil {
        log.Fatalf("open file error: %s", err)
        return
    }

    data, err := json.MarshalIndent(schema, "", strings.Repeat(" ", 4))
    if err != nil {
        log.Fatalf("json marshal error: %s", err)
        return
    }
    bufio.NewWriter(file).Write(data)
}

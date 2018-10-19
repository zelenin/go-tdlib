package main

import (
    "bufio"
    "encoding/json"
    "flag"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "github.com/zelenin/go-tdlib/tlparser"
)

func main() {
    var version string
    var outputFilePath string

    flag.StringVar(&version, "version", "", "TDLib version")
    flag.StringVar(&outputFilePath, "output", "./td_api.json", "json schema file")

    flag.Parse()

    resp, err := http.Get("https://raw.githubusercontent.com/tdlib/td/" + version + "/td/generate/scheme/td_api.tl")
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

    resp, err = http.Get("https://raw.githubusercontent.com/tdlib/td/" + version + "/td/telegram/Td.cpp")
    if err != nil {
        log.Fatalf("http.Get error: %s", err)
        return
    }
    defer resp.Body.Close()

    err = tlparser.ParseCode(resp.Body, schema)
    if err != nil {
        log.Fatalf("parse code error: %s", err)
        return
    }

    err = os.MkdirAll(filepath.Dir(outputFilePath), os.ModePerm)
    if err != nil {
        log.Fatalf("make dir error: %s", filepath.Dir(outputFilePath))
    }

    file, err := os.OpenFile(outputFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
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

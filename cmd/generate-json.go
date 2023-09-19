package main

import (
	"encoding/json"
	"flag"
	"github.com/zelenin/go-tdlib/tlparser"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
	}
	defer resp.Body.Close()

	schema, err := tlparser.Parse(resp.Body)
	if err != nil {
		log.Fatalf("schema parse error: %s", err)
	}

	resp, err = http.Get("https://raw.githubusercontent.com/tdlib/td/" + version + "/td/telegram/Td.cpp")
	if err != nil {
		log.Fatalf("http.Get error: %s", err)
	}
	defer resp.Body.Close()

	err = tlparser.ParseCode(resp.Body, schema)
	if err != nil {
		log.Fatalf("parse code error: %s", err)
	}

	err = os.MkdirAll(filepath.Dir(outputFilePath), os.ModePerm)
	if err != nil {
		log.Fatalf("make dir error: %s", filepath.Dir(outputFilePath))
	}

	file, err := os.Create(outputFilePath)
	if err != nil {
		log.Fatalf("open file error: %s", err)
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", strings.Repeat(" ", 4))
	err = enc.Encode(schema)
	if err != nil {
		log.Fatalf("enc.Encode error: %s", err)

	}
}

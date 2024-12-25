package main

import (
	"encoding/json"
	"flag"
	"github.com/zelenin/go-tdlib/internal/tlparser"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	var version string
	var outputPath string

	flag.StringVar(&version, "version", "", "TDLib version")
	flag.StringVar(&outputPath, "output", "./td_api.json", "json schema file")
	flag.Parse()

	res, err := http.Get("https://raw.githubusercontent.com/tdlib/td/" + version + "/td/generate/scheme/td_api.tl")
	if err != nil {
		log.Fatalf("http.Get error: %s", err)
	}
	defer res.Body.Close()

	schema, err := tlparser.Parse(res.Body)
	if err != nil {
		log.Fatalf("schema parse error: %s", err)
	}

	res, err = http.Get("https://raw.githubusercontent.com/tdlib/td/" + version + "/td/telegram/Td.cpp")
	if err != nil {
		log.Fatalf("http.Get error: %s", err)
	}
	defer res.Body.Close()

	err = tlparser.ParseCode(res.Body, schema)
	if err != nil {
		log.Fatalf("parse code error: %s", err)
	}

	err = os.MkdirAll(filepath.Dir(outputPath), os.ModePerm)
	if err != nil {
		log.Fatalf("make dir error: %s", filepath.Dir(outputPath))
	}

	f, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("open file error: %s", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", strings.Repeat(" ", 4))
	err = enc.Encode(schema)
	if err != nil {
		log.Fatalf("enc.Encode error: %s", err)
	}
}

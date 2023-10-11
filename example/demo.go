package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	tdlib "github.com/zelenin/go-tdlib/client"
)

func main() {
	ctx := context.Background()

	auth := tdlib.NewAppAuthorizer()
	go tdlib.CliInteractor(auth)

	var (
		apiIdRaw = os.Getenv("API_ID")
		apiHash  = os.Getenv("API_HASH")
	)
	apiId64, err := strconv.ParseInt(apiIdRaw, 10, 32)
	if err != nil {
		log.Fatalf("strconv.Atoi error: %s", err)
	}
	apiId := int32(apiId64)

	auth.TdlibParameters <- &tdlib.SetTdlibParametersRequest{
		UseTestDc:              false,
		DatabaseDirectory:      filepath.Join(".tdlib", "database"),
		FilesDirectory:         filepath.Join(".tdlib", "files"),
		UseFileDatabase:        true,
		UseChatInfoDatabase:    true,
		UseMessageDatabase:     true,
		UseSecretChats:         false,
		ApiId:                  apiId,
		ApiHash:                apiHash,
		SystemLanguageCode:     "en",
		DeviceModel:            "Server",
		SystemVersion:          "1.0.0",
		ApplicationVersion:     "1.0.0",
		EnableStorageOptimizer: true,
		IgnoreFileNames:        false,
	}

	_, err = tdlib.SetLogVerbosityLevel(&tdlib.SetLogVerbosityLevelRequest{
		NewVerbosityLevel: 1,
	})
	if err != nil {
		log.Fatalf("SetLogVerbosityLevel error: %s", err)
	}

	client := tdlib.NewClient()
	if err := tdlib.Authorize(ctx, client, auth); err != nil {
		log.Fatalf("Authorize error: %s", err)
	}

	optionValue, err := tdlib.GetOption(&tdlib.GetOptionRequest{
		Name: "version",
	})
	if err != nil {
		log.Fatalf("GetOption error: %s", err)
	}

	log.Printf("TDLib version: %s", optionValue.(*tdlib.OptionValueString).Value)

	me, err := client.GetMe(ctx)
	if err != nil {
		log.Fatalf("GetMe error: %s", err)
	}

	log.Printf("Me: %s %s [%v]", me.FirstName, me.LastName, me.Usernames)

	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		client.Stop(ctx)
		os.Exit(1)
	}()
}

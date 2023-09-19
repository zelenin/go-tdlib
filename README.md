# go-tdlib

Go wrapper for [TDLib (Telegram Database Library)](https://github.com/tdlib/td) with full support of TDLib v1.8.0

## TDLib installation

Use [TDLib build instructions](https://tdlib.github.io/td/build.html)

## Usage

### Client

[Register an application](https://my.telegram.org/apps) to obtain an api_id and api_hash 

```go
package main

import (
    "log"
    "path/filepath"

    "github.com/zelenin/go-tdlib/client"
)

func main() {
    // client authorizer
    authorizer := client.ClientAuthorizer()
    go client.CliInteractor(authorizer)

    // or bot authorizer
    // botToken := "000000000:gsVCGG5YbikxYHC7bP5vRvmBqJ7Xz6vG6td"
    // authorizer := client.BotAuthorizer(botToken)

    const (
        apiId   = 00000
        apiHash = "8pu9yg32qkuukj83ozaqo5zzjwhkxhnk"
    )

    authorizer.TdlibParameters <- &client.TdlibParameters{
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

    logVerbosity := client.WithLogVerbosity(&client.SetLogVerbosityLevelRequest{
        NewVerbosityLevel: 0,
    })

    tdlibClient, err := client.NewClient(authorizer, logVerbosity)
    if err != nil {
        log.Fatalf("NewClient error: %s", err)
    }

    optionValue, err := tdlibClient.GetOption(&client.GetOptionRequest{
        Name: "version",
    })
    if err != nil {
        log.Fatalf("GetOption error: %s", err)
    }

    log.Printf("TDLib version: %s", optionValue.(*client.OptionValueString).Value)

    me, err := tdlibClient.GetMe()
    if err != nil {
        log.Fatalf("GetMe error: %s", err)
    }

    log.Printf("Me: %s %s [%s]", me.FirstName, me.LastName, me.Username)
}

```

### Receive updates

```go
tdlibClient, err := client.NewClient(authorizer)
if err != nil {
    log.Fatalf("NewClient error: %s", err)
}

listener := tdlibClient.GetListener()
defer listener.Close()
 
for update := range listener.Updates {
    if update.GetClass() == client.ClassUpdate {
        log.Printf("%#v", update)
    }
}
```

### Proxy support

```go
proxy := client.WithProxy(&client.AddProxyRequest{
    Server: "1.1.1.1",
    Port:   1080,
    Enable: true,
    Type: &client.ProxyTypeSocks5{
        Username: "username",
        Password: "password",
    },
})

tdlibClient, err := client.NewClient(authorizer, proxy)

```

## Notes

* WIP. Library API can be changed in the future
* The library supports only the [latest tagged version of tdlib](https://github.com/tdlib/td/tags). Don't use it with an arbitrary commit.
* The package includes a .tl-parser and generated [json-schema](https://github.com/zelenin/go-tdlib/tree/master/data) for creating libraries in other languages

## Author

[Aleksandr Zelenin](https://github.com/zelenin/), e-mail: [aleksandr@zelenin.me](mailto:aleksandr@zelenin.me)

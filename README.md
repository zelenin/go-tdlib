# go-tdlib

Go wrapper for [TDLib (Telegram Database Library)](https://github.com/tdlib/td) with full support of the TDLib.
Current supported version of TDLib corresponds to the commit hash [22d49d5](https://github.com/tdlib/td/commit/22d49d5b87a4d5fc60a194dab02dd1d71529687f), updated on 2024-11-27

## TDLib installation

Use [TDLib build instructions](https://tdlib.github.io/td/build.html) with checkmarked `Install built TDLib to /usr/local instead of placing the files to td/tdlib`. Don't forget to checkout a supported commit (see above).


### Windows

Build with environment variables (use full paths):

```
CGO_ENABLED=1
CGO_CFLAGS=-IC:/path/to/tdlib/build/tdlib/include
CGO_LDFLAGS=-LC:/path/to/tdlib/build/tdlib/bin -ltdjson
```

Example for PowerShell:

```powershell
$env:CGO_ENABLED=1; $env:CGO_CFLAGS="-IC:/td/tdlib/include"; $env:CGO_LDFLAGS="-LC:/td/tdlib/bin -ltdjson"; go build -trimpath -ldflags="-s -w" -o demo.exe .\cmd\demo.go
```
To run, put the .dll from C:/td/tdlib/bin to the directory with the compiled .exe.

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

const (
	apiId   = 00000
	apiHash = "8pu9yg32qkuukj83ozaqo5zzjwhkxhnk"
)

func main() {
	tdlibParameters := &client.SetTdlibParametersRequest{
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
	}
    // client authorizer
    authorizer := client.ClientAuthorizer(tdlibParameters)
    go client.CliInteractor(authorizer)

    // or bot authorizer
    // botToken := "000000000:gsVCGG5YbikxYHC7bP5vRvmBqJ7Xz6vG6td"
    // authorizer := client.BotAuthorizer(tdlibParameters, botToken)

	_, err := client.SetLogVerbosityLevel(&client.SetLogVerbosityLevelRequest{
		NewVerbosityLevel: 1,
	})
	if err != nil {
		log.Fatalf("SetLogVerbosityLevel error: %s", err)
	}
	
    tdlibClient, err := client.NewClient(authorizer)
    if err != nil {
        log.Fatalf("NewClient error: %s", err)
    }

	versionOption, err := client.GetOption(&client.GetOptionRequest{
		Name: "version",
	})
	if err != nil {
		log.Fatalf("GetOption error: %s", err)
	}

	commitOption, err := client.GetOption(&client.GetOptionRequest{
		Name: "commit_hash",
	})
	if err != nil {
		log.Fatalf("GetOption error: %s", err)
	}

	log.Printf("TDLib version: %s (commit: %s)", versionOption.(*client.OptionValueString).Value, commitOption.(*client.OptionValueString).Value)

    me, err := tdlibClient.GetMe()
    if err != nil {
        log.Fatalf("GetMe error: %s", err)
    }

    log.Printf("Me: %s %s", me.FirstName, me.LastName)
}

```

### QR Code login

```go
package main

import (
	"github.com/skip2/go-qrcode"
	"log"
	"path/filepath"

	"github.com/zelenin/go-tdlib/client"
)

const (
	apiId   = 00000
	apiHash = "8pu9yg32qkuukj83ozaqo5zzjwhkxhnk"
)

func main() {
	tdlibParameters := &client.SetTdlibParametersRequest{
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
	}
	// client authorizer
	authorizer := client.QrAuthorizer(tdlibParameters, func(link string) error {
		return qrcode.WriteFile(link, qrcode.Medium, 256, "qr.png")
	})

	tdlibClient, err := client.NewClient(authorizer)
	if err != nil {
		log.Fatalf("NewClient error: %s", err)
	}
}

````

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

## Example

[Example application](https://github.com/zelenin/go-tdlib/tree/master/example)

```
cd example
docker build --network host --progress plain --tag tdlib-test .
docker run --rm -it -e "API_ID=00000" -e "API_HASH=abcdef0123456789" tdlib-test ash
./app
```

## Notes

* WIP. Library API can be changed in the future
* The package includes a .tl-parser and generated [json-schema](https://github.com/zelenin/go-tdlib/tree/master/data) for creating libraries in other languages

## Author

[Aleksandr Zelenin](https://github.com/zelenin/), e-mail: [aleksandr@zelenin.me](mailto:aleksandr@zelenin.me)

# go-tdlib

Go wrapper for [TDLib (Telegram Database Library)](https://github.com/tdlib/td) with full support of the TDLib.
Current supported version of TDLib corresponds to the commit hash [b498497](https://github.com/tdlib/td/commit/b498497bbfd6b80c86f800b3546a0170206317d3), updated on 2025-03-16

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
	"context"
	"github.com/zelenin/go-tdlib/client"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
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

	if commitOption.(*client.OptionValueString).Value != client.TDLIB_VERSION {
		log.Printf("TDLib version supported by the library (%s) is not the same as TDLib version (%s)", client.TDLIB_VERSION, commitOption.(*client.OptionValueString).Value)
	}

	me, err := tdlibClient.GetMe(context.Background())
	if err != nil {
		log.Fatalf("GetMe error: %s", err)
	}

	log.Printf("Me: %s %s", me.FirstName, me.LastName)

	ch := make(chan os.Signal, 2)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	tdlibClient.Close(context.Background())
	os.Exit(1)
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
resHandCallback := func(result client.Type) {
    log.Printf("%#v", result.GetType())
}

tdlibClient, err := client.NewClient(authorizer, client.WithResultHandler(client.NewCallbackResultHandler(resHandCallback)))
if err != nil {
    log.Fatalf("NewClient error: %s", err)
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

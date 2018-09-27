// +build libtdjson
// +build linux windows

package client

/*
#cgo linux LDFLAGS: -ltdjson -lstdc++ -lssl -lcrypto -ldl -lz -lm
#cgo windows CFLAGS: -Ic:/td -Ic:/td/example/csharp/build
#cgo windows LDFLAGS: -Lc:/td/example/csharp/build/Release -ltdjson
*/
import "C"

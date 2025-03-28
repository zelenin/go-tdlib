//go:build libtdjson && linux

package client

/*
#cgo linux CFLAGS: -I/usr/local/include
#cgo linux LDFLAGS: -L/usr/local/lib -ltdjson -lstdc++ -lssl -lcrypto -ldl -lz -lm
*/
import "C"

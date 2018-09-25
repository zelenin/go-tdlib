// +build libtdjson
// +build linux

package client

/*
#cgo linux LDFLAGS: -ltdjson -lstdc++ -lssl -lcrypto -ldl -lz -lm
*/
import "C"

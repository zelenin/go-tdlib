// +build !libtdjson
// +build linux

package client

/*
#cgo linux LDFLAGS: -ltdjson_static -ltdjson_private -ltdclient -ltdcore -ltdactor -ltddb -ltdsqlite -ltdnet -ltdutils -lstdc++ -lssl -lcrypto -ldl -lz -lm
*/
import "C"

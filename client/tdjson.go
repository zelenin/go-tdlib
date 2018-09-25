// +build !linux

package client

/*
#cgo darwin CFLAGS: -I/usr/local/include
#cgo windows CFLAGS: -IC:/src/td -IC:/src/td/build
#cgo darwin LDFLAGS: -L/usr/local/lib -L/usr/local/opt/openssl/lib -ltdjson_static -ltdjson_private -ltdclient -ltdcore -ltdactor -ltddb -ltdsqlite -ltdnet -ltdutils -lstdc++ -lssl -lcrypto -ldl -lz -lm
#cgo windows LDFLAGS: -LC:/src/td/build/Debug -ltdjson
*/
import "C"

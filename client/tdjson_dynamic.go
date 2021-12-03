// +build libtdjson
// +build linux darwin windows

package client

/*
#cgo linux CFLAGS: -I/usr/local/include
#cgo linux LDFLAGS: -L/usr/local/lib -ltdjson -lstdc++ -lssl -lcrypto -ldl -lz -lm
#cgo darwin CFLAGS: -I/usr/local/include
#cgo darwin LDFLAGS: -L/usr/local/lib -ltdjson -lstdc++ -lssl -lcrypto -ldl -lz -lm
#cgo windows CFLAGS: -I${SRCDIR}/../../td -I${SRCDIR}/../../td/build
#cgo windows LDFLAGS: -L${SRCDIR}/../../td/tdlib/lib -ltdjson
*/
import "C"

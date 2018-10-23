package client

import (
	"fmt"
	"math/rand"
)

type ExtraGenerator func() string

func UuidV4Generator() ExtraGenerator {
	return func() string {
		var uuid [16]byte
		rand.Read(uuid[:])

		uuid[6] = (uuid[6] & 0x0f) | 0x40
		uuid[8] = (uuid[8] & 0x3f) | 0x80

		return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", uuid[:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
	}
}

package goshazam

import (
	"crypto/rand"
	"fmt"
)

func randomUUID() string {
	var u [16]byte
	_, _ = rand.Read(u[:])
	u[6] = (u[6] & 0x0f) | 0x40 // RFC 4122 version 4
	u[8] = (u[8] & 0x3f) | 0x80 // RFC 4122 variant 10
	return fmt.Sprintf("%08X-%04X-%04X-%04X-%012X", u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
}

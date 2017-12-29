package rvgc

import (
	"fmt"
)

func Cmd2Hex(cmd []string) []byte {

	if cmd[0] == "add" {
		return []byte{'\x33', '\x05', '\xb5', '\x00'}
	} else {
		return []byte{'\x67', '\x80', '\x00', '\x00'}
	}
}

package commands

import (
	"strconv"
	"time"
)

// Right now, for fun, transform it to SetExAt only for AOF
// Much better is to convert the whole handler to SetExAt
// Also, when the validation is separated from the handler, there won't be errors ignored
func SetExTransform(args [][]byte) [][]byte {
	seconds, _ := strconv.ParseInt(string(args[2]), 10, 64)
	expiresAt := time.Now().Unix() + seconds
	return [][]byte{
		[]byte("SETEXAT"),
		args[1],
		[]byte(strconv.FormatInt(expiresAt, 10)),
		args[3],
	}
}

func ExpireTransform(args [][]byte) [][]byte {
	seconds, _ := strconv.ParseInt(string(args[2]), 10, 64)
	expiresAt := time.Now().Unix() + seconds
	return [][]byte{
		[]byte("EXPIREAT"),
		args[1],
		[]byte(strconv.FormatInt(expiresAt, 10)),
	}
}

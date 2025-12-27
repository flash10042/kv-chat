package protocol

import (
	"fmt"
	"strings"
)

func EncodeCommand(args [][]byte) []byte {
	var builder strings.Builder
	fmt.Fprintf(&builder, "*%d\r\n", len(args))
	for _, arg := range args {
		fmt.Fprintf(&builder, "$%d\r\n", len(arg))
		builder.Write(arg)
		builder.WriteString("\r\n")
	}
	return []byte(builder.String())
}

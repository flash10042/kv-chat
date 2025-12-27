package response

import (
	"fmt"
	"strings"
)

const (
	SimpleStringPrefix = "+"
	ErrorPrefix        = "-ERR "
	BulkStringPrefix   = "$"
	IntegerPrefix      = ":"
	ArrayPrefix        = "*"
)

func FormatResponse(prefix string, message string) string {
	return prefix + message + "\r\n"
}

func FormatBulkString(value []byte) string {
	if value == nil {
		return FormatResponse(BulkStringPrefix, "-1")
	}
	return fmt.Sprintf("%s%d\r\n%s\r\n", BulkStringPrefix, len(value), value)
}

func FormatArray(values [][]byte) string {
	length := len(values)
	var builder strings.Builder
	fmt.Fprintf(&builder, "%s%d\r\n", ArrayPrefix, length)
	for _, v := range values {
		fmt.Fprintf(&builder, "%s", FormatBulkString(v))
	}
	return builder.String()
}

func ErrWrongTypeResponse() string {
	return FormatResponse(ErrorPrefix, "Wrong type")
}

func ErrInternalResponse() string {
	return FormatResponse(ErrorPrefix, "Internal error")
}

func ErrEmptyCommandResponse() string {
	return FormatResponse(ErrorPrefix, "Empty command")
}

func ErrInvalidIntegerResponse() string {
	return FormatResponse(ErrorPrefix, "Invalid integer")
}

func ErrWrongArityResponse() string {
	return FormatResponse(ErrorPrefix, "Wrong number of arguments")
}

func ErrUnknownCommandResponse() string {
	return FormatResponse(ErrorPrefix, "Unknown command")
}

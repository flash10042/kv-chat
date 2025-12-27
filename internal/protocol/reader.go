package protocol

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	arrayPrefix      = '*'
	bulkStringPrefix = '$'
)

func ReadCommand(reader *bufio.Reader) ([][]byte, error) {
	b, err := reader.Peek(1)
	if err != nil {
		return nil, err
	}

	if b[0] == arrayPrefix {
		return readArray(reader)
	}

	// Inline command fallback
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("empty command")
	}

	parts := strings.Fields(line)
	args := make([][]byte, len(parts))
	for i, p := range parts {
		args[i] = []byte(p)
	}

	return args, nil
}

func readArray(reader *bufio.Reader) ([][]byte, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	if b != arrayPrefix {
		return nil, fmt.Errorf("expected array prefix, got %c", b)
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	n, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil {
		return nil, err
	}
	if n < 0 {
		return [][]byte{}, nil
	}
	values := make([][]byte, n)
	for i := range values {
		values[i], err = readBulkString(reader)
		if err != nil {
			return nil, err
		}
	}
	return values, nil
}

func readBulkString(reader *bufio.Reader) ([]byte, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	if b != bulkStringPrefix {
		return nil, fmt.Errorf("expected bulk string prefix, got %c", b)
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	lengthStr := strings.TrimSpace(line)
	n, err := strconv.Atoi(lengthStr)
	if err != nil {
		return nil, err
	}
	if n < 0 {
		return []byte{}, nil
	}

	buf := make([]byte, n)
	if _, err = io.ReadFull(reader, buf); err != nil {
		return nil, err
	}
	cr, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	lf, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	if cr != '\r' || lf != '\n' {
		return nil, fmt.Errorf("expected CRLF, got %c%c", cr, lf)
	}
	return buf, nil
}

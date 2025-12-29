package server

import (
	"bufio"
	"io"
	"log"
	"net"

	"github.com/flash10042/kv-chat/internal/persistence"
	"github.com/flash10042/kv-chat/internal/protocol"
	"github.com/flash10042/kv-chat/internal/store"
)

func HandleConnection(conn net.Conn, storage *store.Storage, aof *persistence.AOF) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		args, err := protocol.ReadCommand(reader)
		if err != nil {
			if err != io.EOF {
				log.Printf("Failed to read command: %v", err)
			}
			return
		}

		response := protocol.DispatchCommand(protocol.DispatchModePublic, args, storage, aof)

		if _, err := writer.WriteString(response); err != nil {
			log.Printf("Failed to write response: %v", err)
			return
		}
		if err := writer.Flush(); err != nil {
			log.Printf("Failed to flush writer: %v", err)
			return
		}
	}
}

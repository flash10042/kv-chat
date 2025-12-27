package protocol

import (
	"log"
	"strings"

	"github.com/flash10042/kv-chat/internal/commands"
	"github.com/flash10042/kv-chat/internal/persistence"
	"github.com/flash10042/kv-chat/internal/response"
	"github.com/flash10042/kv-chat/internal/store"
)

func checkArity(length int, arity int) bool {
	if arity >= 0 && length != arity {
		return false
	}
	if arity < 0 && length < -arity {
		return false
	}
	return true
}

func DispatchCommand(args [][]byte, storage *store.Storage, aof *persistence.AOF) string {
	if len(args) == 0 {
		return response.ErrEmptyCommandResponse()
	}

	name := strings.ToUpper(string(args[0]))
	command, ok := commands.Registry[name]
	if !ok {
		return response.ErrUnknownCommandResponse()
	}
	if !checkArity(len(args), command.Arity) {
		return response.ErrWrongArityResponse()
	}
	if command.Mutates && aof != nil {
		err := aof.Append(EncodeCommand(args))
		if err != nil {
			log.Printf("Failed to append command to AOF: %v", err)
		}
	}
	return command.Handler(args, storage)
}

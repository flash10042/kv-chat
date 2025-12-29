package protocol

import (
	"log"
	"strings"

	"github.com/flash10042/kv-chat/internal/commands"
	"github.com/flash10042/kv-chat/internal/persistence"
	"github.com/flash10042/kv-chat/internal/response"
	"github.com/flash10042/kv-chat/internal/store"
)

type DispatchMode int

const (
	DispatchModePublic DispatchMode = iota
	DispatchModePrivate
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

func DispatchCommand(dispatchMode DispatchMode, args [][]byte, storage *store.Storage, aof *persistence.AOF) string {
	if len(args) == 0 {
		return response.ErrEmptyCommandResponse()
	}

	name := strings.ToUpper(string(args[0]))
	command, ok := commands.Registry[name]
	if !ok {
		return response.ErrUnknownCommandResponse()
	}

	if command.IsPrivate && dispatchMode != DispatchModePrivate {
		return response.ErrUnknownCommandResponse()
	}

	if !checkArity(len(args), command.Arity) {
		return response.ErrWrongArityResponse()
	}

	// Ideally, handler wouldn't return a bool, but we need it since validation is integrated into handler
	response, ok := command.Handler(args, storage)

	if ok && dispatchMode == DispatchModePublic && command.Mutates && aof != nil {
		// Use AOFTransform if available, otherwise use original args
		aofArgs := args
		if command.AOFTransform != nil {
			aofArgs = command.AOFTransform(args)
		}
		err := aof.Append(EncodeCommand(aofArgs))
		if err != nil {
			log.Printf("Failed to append command to AOF: %v", err)
		}
	}

	return response
}

package commands

import (
	"strconv"
	"time"

	"github.com/flash10042/kv-chat/internal/response"
	"github.com/flash10042/kv-chat/internal/store"
)

type Handler func(args [][]byte, storage *store.Storage) (string, bool)

// TODO: Separate validation and handling
type Command struct {
	Name         string
	Arity        int
	Mutates      bool
	Handler      Handler
	IsPrivate    bool
	AOFTransform func(args [][]byte) [][]byte
}

func init() {
	register(Command{
		Name:    "PING",
		Arity:   1,
		Mutates: false,
		Handler: PingHandler,
	})
	register(Command{
		Name:    "SET",
		Arity:   3,
		Mutates: true,
		Handler: SetHandler,
	})
	register(Command{
		Name:    "GET",
		Arity:   2,
		Mutates: false,
		Handler: GetHandler,
	})
	register(Command{
		Name:    "LPUSH",
		Arity:   3,
		Mutates: true,
		Handler: LPushHandler,
	})
	register(Command{
		Name:    "RPUSH",
		Arity:   3,
		Mutates: true,
		Handler: RPushHandler,
	})
	register(Command{
		Name:    "LRANGE",
		Arity:   4,
		Mutates: false,
		Handler: LRangeHandler,
	})
	register(Command{
		Name:         "EXPIRE",
		Arity:        3,
		Mutates:      true,
		Handler:      ExpireHandler,
		AOFTransform: ExpireTransform,
	})
	register(Command{
		Name:    "TTL",
		Arity:   2,
		Mutates: false,
		Handler: TTLHandler,
	})
	register(Command{
		Name:    "DEL",
		Arity:   2,
		Mutates: true,
		Handler: DelHandler,
	})
	register(Command{
		Name:    "EXISTS",
		Arity:   2,
		Mutates: false,
		Handler: ExistsHandler,
	})
	register(Command{
		Name:         "SETEX",
		Arity:        4,
		Mutates:      true,
		Handler:      SetExHandler,
		AOFTransform: SetExTransform,
	})
	register(Command{
		Name:      "EXPIREAT",
		Arity:     3,
		Mutates:   true,
		Handler:   ExpireAtHandler,
		IsPrivate: true,
	})
	register(Command{
		Name:      "SETEXAT",
		Arity:     4,
		Mutates:   true,
		Handler:   SetExAtHandler,
		IsPrivate: true,
	})
}

func PingHandler(args [][]byte, storage *store.Storage) (string, bool) {
	return response.FormatResponse(response.SimpleStringPrefix, "PONG"), true
}

func SetHandler(args [][]byte, storage *store.Storage) (string, bool) {
	key := string(args[1])
	value := args[2]
	storage.Set(key, value)
	return response.FormatResponse(response.SimpleStringPrefix, "OK"), true
}

func GetHandler(args [][]byte, storage *store.Storage) (string, bool) {
	key := string(args[1])
	value, err := storage.Get(key)
	if err != nil {
		if err == store.ErrWrongType {
			return response.ErrWrongTypeResponse(), false
		}
		return response.ErrInternalResponse(), false
	}
	return response.FormatBulkString(value), true
}

func LPushHandler(args [][]byte, storage *store.Storage) (string, bool) {
	key := string(args[1])
	value := args[2]
	length, err := storage.LPush(key, value)
	if err != nil {
		if err == store.ErrWrongType {
			return response.ErrWrongTypeResponse(), false
		}
		return response.ErrInternalResponse(), false
	}
	return response.FormatResponse(response.IntegerPrefix, strconv.Itoa(length)), true
}

func RPushHandler(args [][]byte, storage *store.Storage) (string, bool) {
	key := string(args[1])
	value := args[2]
	length, err := storage.RPush(key, value)
	if err != nil {
		if err == store.ErrWrongType {
			return response.ErrWrongTypeResponse(), false
		}
		return response.ErrInternalResponse(), false
	}
	return response.FormatResponse(response.IntegerPrefix, strconv.Itoa(length)), true
}

func LRangeHandler(args [][]byte, storage *store.Storage) (string, bool) {
	key := string(args[1])
	start, err := strconv.Atoi(string(args[2]))
	if err != nil {
		return response.ErrInvalidIntegerResponse(), false
	}
	end, err := strconv.Atoi(string(args[3]))
	if err != nil {
		return response.ErrInvalidIntegerResponse(), false
	}
	values, err := storage.LRange(key, start, end)
	if err != nil {
		if err == store.ErrWrongType {
			return response.ErrWrongTypeResponse(), false
		}
		return response.ErrInternalResponse(), false
	}
	return response.FormatArray(values), true
}

func ExpireHandler(args [][]byte, storage *store.Storage) (string, bool) {
	key := string(args[1])
	seconds, err := strconv.ParseInt(string(args[2]), 10, 64)
	if err != nil {
		return response.ErrInvalidIntegerResponse(), false
	}
	ok := storage.Expire(key, seconds)
	if ok {
		return response.FormatResponse(response.IntegerPrefix, "1"), true
	}
	return response.FormatResponse(response.IntegerPrefix, "0"), true
}

func TTLHandler(args [][]byte, storage *store.Storage) (string, bool) {
	key := string(args[1])
	ttl := storage.TTL(key)
	return response.FormatResponse(response.IntegerPrefix, strconv.FormatInt(ttl, 10)), true
}

func DelHandler(args [][]byte, storage *store.Storage) (string, bool) {
	key := string(args[1])
	ok := storage.Del(key)
	if ok {
		return response.FormatResponse(response.IntegerPrefix, "1"), true
	}
	return response.FormatResponse(response.IntegerPrefix, "0"), true
}

func ExistsHandler(args [][]byte, storage *store.Storage) (string, bool) {
	key := string(args[1])
	ok := storage.Exists(key)
	if ok {
		return response.FormatResponse(response.IntegerPrefix, "1"), true
	}
	return response.FormatResponse(response.IntegerPrefix, "0"), true
}

func SetExHandler(args [][]byte, storage *store.Storage) (string, bool) {
	key := string(args[1])
	seconds, err := strconv.ParseInt(string(args[2]), 10, 64)
	if err != nil {
		return response.ErrInvalidIntegerResponse(), false
	}
	value := args[3]
	storage.SetEx(key, seconds, value)
	return response.FormatResponse(response.SimpleStringPrefix, "OK"), true
}

func ExpireAtHandler(args [][]byte, storage *store.Storage) (string, bool) {
	key := string(args[1])
	timestamp, err := strconv.ParseInt(string(args[2]), 10, 64)
	if err != nil {
		return response.ErrInvalidIntegerResponse(), false
	}
	expiresAt := time.Unix(timestamp, 0)
	ok := storage.ExpireAt(key, expiresAt)
	if ok {
		return response.FormatResponse(response.IntegerPrefix, "1"), true
	}
	return response.FormatResponse(response.IntegerPrefix, "0"), true
}

func SetExAtHandler(args [][]byte, storage *store.Storage) (string, bool) {
	key := string(args[1])
	timestamp, err := strconv.ParseInt(string(args[2]), 10, 64)
	if err != nil {
		return response.ErrInvalidIntegerResponse(), false
	}
	expiresAt := time.Unix(timestamp, 0)
	value := args[3]
	storage.SetExAt(key, expiresAt, value)
	return response.FormatResponse(response.SimpleStringPrefix, "OK"), true
}

package commands

import (
	"strconv"

	"github.com/flash10042/kv-chat/internal/response"
	"github.com/flash10042/kv-chat/internal/store"
)

type Handler func(args [][]byte, storage *store.Storage) string

type Command struct {
	Name    string
	Arity   int
	Mutates bool
	Handler Handler
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
		Name:    "EXPIRE",
		Arity:   3,
		Mutates: true,
		Handler: ExpireHandler,
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
		Name:    "SETEX",
		Arity:   4,
		Mutates: true,
		Handler: SetExHandler,
	})
}

func PingHandler(args [][]byte, storage *store.Storage) string {
	return response.FormatResponse(response.SimpleStringPrefix, "PONG")
}

func SetHandler(args [][]byte, storage *store.Storage) string {
	key := string(args[1])
	value := args[2]
	storage.Set(key, value)
	return response.FormatResponse(response.SimpleStringPrefix, "OK")
}

func GetHandler(args [][]byte, storage *store.Storage) string {
	key := string(args[1])
	value, err := storage.Get(key)
	if err != nil {
		if err == store.ErrWrongType {
			return response.ErrWrongTypeResponse()
		}
		return response.ErrInternalResponse()
	}
	return response.FormatBulkString(value)
}

func LPushHandler(args [][]byte, storage *store.Storage) string {
	key := string(args[1])
	value := args[2]
	length, err := storage.LPush(key, value)
	if err != nil {
		if err == store.ErrWrongType {
			return response.ErrWrongTypeResponse()
		}
		return response.ErrInternalResponse()
	}
	return response.FormatResponse(response.IntegerPrefix, strconv.Itoa(length))
}

func RPushHandler(args [][]byte, storage *store.Storage) string {
	key := string(args[1])
	value := args[2]
	length, err := storage.RPush(key, value)
	if err != nil {
		if err == store.ErrWrongType {
			return response.ErrWrongTypeResponse()
		}
		return response.ErrInternalResponse()
	}
	return response.FormatResponse(response.IntegerPrefix, strconv.Itoa(length))
}

func LRangeHandler(args [][]byte, storage *store.Storage) string {
	key := string(args[1])
	start, err := strconv.Atoi(string(args[2]))
	if err != nil {
		return response.ErrInvalidIntegerResponse()
	}
	end, err := strconv.Atoi(string(args[3]))
	if err != nil {
		return response.ErrInvalidIntegerResponse()
	}
	values, err := storage.LRange(key, start, end)
	if err != nil {
		if err == store.ErrWrongType {
			return response.ErrWrongTypeResponse()
		}
		return response.ErrInternalResponse()
	}
	return response.FormatArray(values)
}

func ExpireHandler(args [][]byte, storage *store.Storage) string {
	key := string(args[1])
	seconds, err := strconv.ParseInt(string(args[2]), 10, 64)
	if err != nil {
		return response.ErrInvalidIntegerResponse()
	}
	ok := storage.Expire(key, seconds)
	if ok {
		return response.FormatResponse(response.IntegerPrefix, "1")
	}
	return response.FormatResponse(response.IntegerPrefix, "0")
}

func TTLHandler(args [][]byte, storage *store.Storage) string {
	key := string(args[1])
	ttl := storage.TTL(key)
	return response.FormatResponse(response.IntegerPrefix, strconv.FormatInt(ttl, 10))
}

func DelHandler(args [][]byte, storage *store.Storage) string {
	key := string(args[1])
	ok := storage.Del(key)
	if ok {
		return response.FormatResponse(response.IntegerPrefix, "1")
	}
	return response.FormatResponse(response.IntegerPrefix, "0")
}

func ExistsHandler(args [][]byte, storage *store.Storage) string {
	key := string(args[1])
	ok := storage.Exists(key)
	if ok {
		return response.FormatResponse(response.IntegerPrefix, "1")
	}
	return response.FormatResponse(response.IntegerPrefix, "0")
}

func SetExHandler(args [][]byte, storage *store.Storage) string {
	key := string(args[1])
	seconds, err := strconv.ParseInt(string(args[2]), 10, 64)
	if err != nil {
		return response.ErrInvalidIntegerResponse()
	}
	value := args[3]
	storage.SetEx(key, seconds, value)
	return response.FormatResponse(response.SimpleStringPrefix, "OK")
}

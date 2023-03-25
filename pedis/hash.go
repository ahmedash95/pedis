package pedis

import (
	"fmt"
	"sync"
)

var (
	hashMU sync.RWMutex
	hash   = make(map[string]map[string]string)
)

func HSetHandler(conn *Conn, args []Value) bool {
	if len(args) != 3 {
		conn.Writer.WriteError("ERR wrong number of arguments for 'hset' command")
		return true
	}

	key := args[0].String()
	field := args[1].String()
	value := args[2].String()

	hashMU.Lock()
	if _, ok := hash[key]; !ok {
		hash[key] = make(map[string]string)
	}
	hash[key][field] = value
	hashMU.Unlock()

	conn.Writer.WriteInteger(1)
	return true
}

func HGetHandler(conn *Conn, args []Value) bool {
	if len(args) != 2 {
		conn.Writer.WriteError("ERR wrong number of arguments for 'hget' command")
		return true
	}

	key := args[0].String()
	field := args[1].String()

	hashMU.RLock()
	value, ok := hash[key][field]
	hashMU.RUnlock()

	if !ok {
		conn.Writer.WriteNull()
		return true
	}

	conn.Writer.WriteBulkString(value)
	return true
}

func HGetAllHandler(conn *Conn, args []Value) bool {
	if len(args) != 1 {
		conn.Writer.WriteError("ERR wrong number of arguments for 'hgetall' command")
		return true
	}

	key := args[0].String()

	hashMU.RLock()
	records, ok := hash[key]
	hashMU.RUnlock()

	if !ok {
		conn.Writer.WriteNull()
		return true
	}

	var values []Value
	for k, v := range records {
		values = append(values, BulkString(k), BulkString(v))
	}

	err := conn.Writer.WriteArray(Value{typ: Array, array: values})
	if err != nil {
		fmt.Println(err)
	}

	return true
}

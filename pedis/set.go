package pedis

import "sync"

var setMU sync.RWMutex
var set = make(map[string]string)

func SetHandler(conn *Conn, args []Value) bool {
	if len(args) != 2 {
		conn.Writer.WriteError("ERR wrong number of arguments for 'set' command")
		return true
	}

	key := args[0].String()
	value := args[1].String()

	setMU.Lock()
	set[key] = value
	setMU.Unlock()

	conn.Writer.WriteSimpleString("OK")
	return true
}

func GetHandler(conn *Conn, args []Value) bool {
	if len(args) != 1 {
		conn.Writer.WriteError("ERR wrong number of arguments for 'get' command")
		return true
	}

	key := args[0].String()

	setMU.RLock()
	value, ok := set[key]
	setMU.RUnlock()

	if !ok {
		conn.Writer.WriteNull()
		return true
	}

	conn.Writer.WriteBulkString(value)
	return true
}

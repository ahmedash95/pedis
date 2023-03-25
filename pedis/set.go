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

func DelHandler(conn *Conn, args []Value) bool {
	if len(args) != 1 {
		conn.Writer.WriteError("ERR wrong number of arguments for 'del' command")
		return true
	}

	key := args[0].String()
	setMU.Lock()
	if _, ok := set[key]; !ok {
		setMU.Unlock()

		conn.Writer.WriteInteger(0)
		return true
	}
	delete(set, key)
	setMU.Unlock()

	conn.Writer.WriteInteger(1)
	return true
}

func ExistsHandler(conn *Conn, args []Value) bool {
	if len(args) != 1 {
		conn.Writer.WriteError("ERR wrong number of arguments for 'exists' command")
		return true
	}

	key := args[0].String()

	setMU.RLock()
	_, ok := set[key]
	setMU.RUnlock()

	if !ok {
		conn.Writer.WriteInteger(0)
		return true
	}

	conn.Writer.WriteInteger(1)
	return true
}

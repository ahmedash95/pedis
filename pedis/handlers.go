package pedis

var defaultHandlers = map[string]func(conn *Conn, args []Value) bool{
	"PING":    pingHandler,
	"SET":     SetHandler,
	"GET":     GetHandler,
	"HSET":    HSetHandler,
	"HGET":    HGetHandler,
	"HGETALL": HGetAllHandler,
}

func pingHandler(conn *Conn, args []Value) bool {
	resp := "PONG"
	if len(args) > 0 {
		resp = args[0].String()
	}

	conn.Writer.WriteSimpleString(resp)
	return true
}

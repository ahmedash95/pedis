package pedis

type CommandHandler struct {
	Handler func(conn *Conn, args []Value) bool
	persist bool
}

func (h *CommandHandler) should_persist() bool {
	return h.persist
}

func (h *CommandHandler) call(conn *Conn, args []Value) bool {
	return h.Handler(conn, args)
}

var defaultHandlers = map[string]CommandHandler{
	"PING":    CommandHandler{pingHandler, false},
	"SET":     CommandHandler{SetHandler, true},
	"GET":     CommandHandler{GetHandler, false},
	"DEL":     CommandHandler{DelHandler, true},
	"EXISTS":  CommandHandler{ExistsHandler, false},
	"HSET":    CommandHandler{HSetHandler, true},
	"HGET":    CommandHandler{HGetHandler, false},
	"HGETALL": CommandHandler{HGetAllHandler, false},
	"HDEL":    CommandHandler{HDelHandler, true},
	"HLEN":    CommandHandler{HLenHandler, false},
	"HKEYS":   CommandHandler{HKeysHandler, false},
	"HVALS":   CommandHandler{HValsHandler, false},
}

func pingHandler(conn *Conn, args []Value) bool {
	resp := "PONG"
	if len(args) > 0 {
		resp = args[0].String()
	}

	conn.Writer.WriteSimpleString(resp)
	return true
}

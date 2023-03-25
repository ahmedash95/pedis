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
	"PING":    {pingHandler, false},
	"SET":     {SetHandler, true},
	"GET":     {GetHandler, false},
	"DEL":     {DelHandler, true},
	"EXISTS":  {ExistsHandler, false},
	"HSET":    {HSetHandler, true},
	"HGET":    {HGetHandler, false},
	"HGETALL": {HGetAllHandler, false},
	"HDEL":    {HDelHandler, true},
	"HLEN":    {HLenHandler, false},
	"HKEYS":   {HKeysHandler, false},
	"HVALS":   {HValsHandler, false},
}

func pingHandler(conn *Conn, args []Value) bool {
	resp := "PONG"
	if len(args) > 0 {
		resp = args[0].String()
	}

	conn.Writer.WriteSimpleString(resp)
	return true
}

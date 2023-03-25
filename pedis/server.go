package pedis

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
)

type Conn struct {
	base       net.Conn
	Writer     *Writer
	Reader     *Reader
	remoteAddr string
}

func NewConn(conn net.Conn) *Conn {
	return &Conn{
		base:       conn,
		Writer:     NewWriter(conn),
		Reader:     NewReader(conn),
		remoteAddr: conn.RemoteAddr().String(),
	}
}

type Config struct {
	EnableAof bool
	AofFile   string
}

type Server struct {
	mu       sync.RWMutex
	handlers map[string]CommandHandler
	accept   func(conn *Conn) bool
	config   *Config
	Aof      *Aof
}

func NewServer(config *Config) *Server {
	handlers := make(map[string]CommandHandler)

	for command, handler := range defaultHandlers {
		handlers[strings.ToUpper(command)] = handler
	}

	s := &Server{
		handlers: handlers,
		config:   config,
	}

	if config.EnableAof {
		bootstrapAof(s)
	}

	return s
}

func (s *Server) HandleFunc(command string, handler CommandHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[strings.ToUpper(command)] = handler
}

func (s *Server) AcceptFunc(accept func(conn *Conn) bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accept = accept
}

func (s *Server) handleConn(conn net.Conn) error {
	c := NewConn(conn)
	s.mu.Lock()
	accept := s.accept
	s.mu.Unlock()
	if accept != nil && !accept(c) {
		return nil
	}

	for {
		value, err := c.Reader.resp.Read()
		if err != nil {
			return err
		}

		if value.typ == 0 {
			continue
		}

		if !value.IsArray() {
			continue
		}

		request := value.Array()

		command := strings.ToUpper(request[0].String())

		s.mu.RLock()
		handler, ok := s.handlers[command]
		s.mu.RUnlock()

		if !ok {
			if err := c.Writer.WriteError(fmt.Sprintf("ERR unknown command \"%s\"", command)); err != nil {
				return err
			}

			continue
		}

		if !handler.call(c, request[1:]) {
			return nil
		}

		if s.Aof != nil && handler.should_persist() {
			if err := s.Aof.Append(value); err != nil {
				return err
			}
		}
	}
}

func (s *Server) ListenAndServe(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go func() {
			err = s.handleConn(conn)
			defer conn.Close()

			if err != nil {
				if err == io.EOF {
					return
				}

				fmt.Println(err)
			}
		}()
	}
}

func bootstrapAof(s *Server) {
	aof, err := NewAof(s.config.AofFile)
	if err != nil {
		panic(err)
	}

	s.Aof = aof

	aof.ReadValues(func(value Value) bool {
		commands := value.Array()

		command := strings.ToUpper(commands[0].String())

		s.mu.RLock()
		handler, ok := s.handlers[command]
		s.mu.RUnlock()

		if !ok {
			return true
		}

		// create fake connection with fake writer
		conn := &Conn{
			Writer: NewWriter(io.Discard),
		}

		handler.call(conn, commands[1:])

		return true
	})
}

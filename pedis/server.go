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

type Server struct {
	mu       sync.RWMutex
	handlers map[string]func(conn *Conn, args []Value) bool
	accept   func(conn *Conn) bool
}

func NewServer() *Server {
	handlers := make(map[string]func(conn *Conn, args []Value) bool)

	for command, handler := range defaultHandlers {
		handlers[strings.ToUpper(command)] = handler
	}

	return &Server{
		handlers: handlers,
	}
}

func (s *Server) HandleFunc(command string, handler func(conn *Conn, args []Value) bool) {
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

		commands := value.Array()

		command := strings.ToUpper(commands[0].String())

		s.mu.RLock()
		handler, ok := s.handlers[command]
		s.mu.RUnlock()

		if !ok {
			if err := c.Writer.WriteError(fmt.Sprintf("ERR unknown command \"%s\"", command)); err != nil {
				return err
			}

			continue
		}

		if !handler(c, commands[1:]) {
			return nil
		}
	}

	return nil
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

				panic(err)
			}
		}()
	}
}

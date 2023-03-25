package pedis

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
)

// Resp is a Redis Serialization Protocol (RESP) parser.

// //////////////////////////////////////////////////////////
//
//	VALUE
//
// //////////////////////////////////////////////////////////
// Type represents the type of a RESP value.
type Type byte

const (
	Array   = '*'
	Bulk    = '$'
	Error   = '-'
	Integer = ':'
	String  = '+'
)

type Value struct {
	typ     Type
	integer int
	str     string
	array   []Value
	null    bool
}

func (v Value) Type() Type {
	return v.typ
}

func (v Value) IsArray() bool {
	return v.typ == Array
}

func (v Value) IsBulk() bool {
	return v.typ == Bulk
}

func (v Value) IsError() bool {
	return v.typ == Error
}

func (v Value) IsInteger() bool {
	return v.typ == Integer
}

func (v Value) IsString() bool {
	return v.typ == String
}

func (v Value) IsNull() bool {
	return v.null
}

func (v Value) Array() []Value {
	return v.array
}

func (v Value) Bulk() string {
	return v.str
}

func (v Value) Error() string {
	return v.str
}

func (v Value) Integer() int {
	return v.integer
}

func (v Value) String() string {
	return v.str
}

func BulkString(str string) Value {
	return Value{
		typ: Bulk,
		str: str,
	}
}

// //////////////////////////////////////////////////////////
//
//	RESP
//
// //////////////////////////////////////////////////////////
type Resp struct {
	buf *bufio.Reader
}

func NewResp(buf *bufio.Reader) *Resp {
	return &Resp{
		buf: buf,
	}
}

func (r *Resp) Read() (val Value, err error) {
	b, err := r.buf.ReadByte()
	if err != nil {
		return val, err
	}

	switch b {
	case Bulk:
		return r.readBulk()
	case Array:
		return r.readArray()
	case 13: // carriage return (CR)
	case 10: // line feed (LF)
	default:
		return val, fmt.Errorf("unknown type: %v", string(b))
	}

	return val, nil
}

func (r *Resp) readArray() (val Value, err error) {
	// Read the length of the array
	len, _, err := r.readInteger()
	if err != nil {
		return val, err
	}

	if len > 1024*1024 {
		return val, fmt.Errorf("array too long: %d", len)
	}

	// Read the array
	for i := 0; i < len; i++ {
		value, err := r.Read()
		if err != nil {
			return val, err
		}

		val.array = append(val.array, value)
	}

	val.typ = Array
	return val, nil
}

func (r *Resp) readBulk() (val Value, err error) {
	// Read the length of the bulk string
	len, _, err := r.readInteger()
	if err != nil {
		return val, err
	}

	if len > 512*1024*1024 {
		return val, fmt.Errorf("bulk string too long: %d", len)
	}

	// Read the bulk string
	bulk := make([]byte, len)
	_, err = r.buf.Read(bulk)
	if err != nil {
		return val, err
	}

	val.typ = Bulk
	val.str = string(bulk)

	// Read the trailing CRLF
	_, _, _ = r.readLine()

	return val, nil
}

func (r *Resp) readInteger() (x int, n int, err error) {
	line, n, err := r.readLine()
	if err != nil {
		return 0, 0, err
	}
	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, n, err
	}
	return int(i64), n, nil
}

func (r *Resp) readLine() (line []byte, n int, err error) {
	for {
		b, err := r.buf.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		n += 1
		line = append(line, b)
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}
	return line[:len(line)-2], n, nil
}

// //////////////////////////////////////////////////////////
//
//	WRITER
//
// //////////////////////////////////////////////////////////
type Writer struct {
	w io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w}
}

func (w *Writer) WriteSimpleString(s string) error {
	_, err := w.w.Write([]byte(fmt.Sprintf("+%s\r\n", s)))
	return err
}

func (w *Writer) WriteError(s string) error {
	_, err := w.w.Write([]byte(fmt.Sprintf("-%s\r\n", s)))
	return err
}

func (w *Writer) WriteInteger(i int) error {
	_, err := w.w.Write([]byte(fmt.Sprintf(":%d\r\n", i)))
	return err
}

func (w *Writer) WriteBulkString(s string) error {
	_, err := w.w.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(s), s)))
	return err
}

func (w *Writer) WriteNull() error {
	_, err := w.w.Write([]byte("$-1\r\n"))
	return err
}

func (w *Writer) WriteArray(v Value) error {
	_, err := w.w.Write([]byte(fmt.Sprintf("*%d\r\n", len(v.array))))

	if err != nil {
		return err
	}

	for _, v := range v.array {
		switch v.typ {
		case String:
			err = w.WriteSimpleString(v.str)
		case Error:
			err = w.WriteError(v.str)
		case Integer:
			err = w.WriteInteger(v.integer)
		case Bulk:
			err = w.WriteBulkString(v.str)
		}
	}

	return err
}

func (v Value) MarshalResp() ([]byte, error) {
	return marshalResp(v)
}

func marshalResp(v Value) ([]byte, error) {
	var buf bytes.Buffer

	switch v.typ {
	case String:
		buf.WriteString(fmt.Sprintf("+%s\r\n", v.str))
	case Error:
		buf.WriteString(fmt.Sprintf("-%s\r\n", v.str))
	case Integer:
		buf.WriteString(fmt.Sprintf(":%d\r\n", v.integer))
	case Bulk:
		buf.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(v.str), v.str))
	case Array:
		buf.WriteString(fmt.Sprintf("*%d\r\n", len(v.array)))
		for _, val := range v.array {
			b, err := marshalResp(val)
			if err != nil {
				return buf.Bytes(), err
			}
			buf.Write(b)
		}
	default:
		return buf.Bytes(), fmt.Errorf("unknown type: %v", v.typ)
	}

	return buf.Bytes(), nil
}

// //////////////////////////////////////////////////////////
//
//	READER
//
// //////////////////////////////////////////////////////////
type Reader struct {
	r    *bufio.Reader
	resp *Resp
}

func NewReader(r io.Reader) *Reader {
	bufReader := bufio.NewReader(r)
	return &Reader{bufReader, NewResp(bufReader)}
}

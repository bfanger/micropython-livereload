package multiplex

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
)

// Example for sending "ABC" on channel "uart":
// <MSG 3 uart>ABC</MSG>

// Starts with "<MSG "
// Followed by the number of bytes of the message excluding escape characters (3)
// Followed by a space " "
// Followed by the channel (uart)
// Followed by a ">"
// Followed by the message
// Followed by "</MSG>"  linebreak

func Write(w io.Writer, channel string, message []byte) error {
	tag := fmt.Sprintf("<MSG %d %s>", len(message), channel)
	if _, err := w.Write([]byte(tag)); err != nil {
		return err
	}
	// @todo Excape "<MSG " in message
	if _, err := w.Write(message); err != nil {
		return err
	}
	if _, err := w.Write([]byte("</MSG>")); err != nil {
		return err
	}
	return nil
}

type SerializedWriter struct {
	w     io.Writer
	mutex sync.Mutex
}

func NewSerializedWriter(w io.Writer) *SerializedWriter {
	return &SerializedWriter{w: w}
}
func (s *SerializedWriter) Write(b []byte) (int, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.w.Write(b)
}

type Message struct {
	Channel string
	Body    []byte
}

// Read a message from a stream
// Returns the message and the number of bytes read
func Read(r io.Reader) (*Message, int, error) {
	type ReadMode int
	const (
		Open ReadMode = iota
		Length
		Channel
		End
	)
	var buf bytes.Buffer
	start := []byte{'<', 'M', 'S', 'G', ' '}
	end := []byte{'<', '/', 'M', 'S', 'G', '>'}
	charBuffer := make([]byte, 1)
	prev := byte(' ')
	cursor := 0
	pos := 0
	mode := Open
	var length int
	var err error
	var msg *Message
	for {

		if n, err := r.Read(charBuffer); err != nil {
			return nil, pos + n, err
		}
		char := charBuffer[0]
		pos++
		switch mode {
		case Open:
			if cursor == 0 && prev == '\\' {
				break
			}
			if char == start[cursor] {
				cursor++
			}
			if cursor > len(start)-1 {
				mode = Length
			}
		case Length:
			if _, err := buf.Write(charBuffer); err != nil {
				return nil, pos, err
			}
			if char == ' ' {
				length, err = strconv.Atoi(strings.TrimRight(buf.String(), " "))
				if err != nil {
					return nil, pos, err
				}
				mode = Channel
				buf.Reset()
			}
		case Channel:
			if _, err := buf.Write(charBuffer); err != nil {
				return nil, pos, err
			}
			if char == '>' {
				msg = &Message{Channel: strings.TrimRight(buf.String(), ">")}
				// mode BODY
				msg.Body = make([]byte, length)
				if n, err := r.Read(msg.Body); err != nil {
					return nil, pos + n, err
				}
				pos += length
				// @todo Detect excaped "\<MSG "
				mode = End
				buf.Reset()
				cursor = 0
			}
		case End:
			if char != end[cursor] {
				return nil, pos, errors.New("missing closing tag")
			}
			cursor++
			if cursor > len(end)-1 {
				return msg, pos, nil
			}
		}
		prev = char
	}
}

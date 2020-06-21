// Package multiplex Combine multiple writers into a single writer and on the onther end split the single reader into multiple writers.
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

// Messaging format:
//
// For example sending "ABC" on channel "uart" :
// <MSG 3 uart>ABC</MSG>
//
// Starts with "<MSG "
// Followed by the number of bytes of the message including escape characters (3)
// Followed by a space " "
// Followed by the channel (uart)
// Followed by a ">"
// Followed by the message
// Followed by "</MSG>"  linebreak

const escapeChar = byte('\\')

var openTag = []byte{'<', 'M', 'S', 'G', ' '}
var closeTag = []byte{'<', '/', 'M', 'S', 'G', '>'}
var escapedTag = append([]byte{escapeChar}, openTag...)
var minMessageSize = len(openTag) + 4 + len(closeTag)

func validateChannel(name string) error {
	if strings.Contains(name, ">") {
		return errors.New("'>' is not allowed for a channel name")
	}
	return nil
}

func Write(w io.Writer, channel string, message []byte) error {
	if err := validateChannel(channel); err != nil {
		return err
	}
	if _, err := w.Write(openTag); err != nil {
		return err
	}
	escaped := bytes.ReplaceAll(message, openTag, escapedTag)
	if _, err := w.Write([]byte(fmt.Sprintf("%d %s>", len(escaped), channel))); err != nil {
		return err
	}
	if _, err := w.Write(escaped); err != nil {
		return err
	}
	if _, err := w.Write(closeTag); err != nil {
		return err
	}
	return nil
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
			if cursor == 0 && prev == escapeChar {
				break
			}
			if char == openTag[cursor] {
				cursor++
			}
			if cursor > len(openTag)-1 {
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
				// mode BODY
				escaped := make([]byte, length)
				if n, err := r.Read(escaped); err != nil {
					return nil, pos + n, err
				}
				pos += length
				msg = &Message{
					Channel: strings.TrimRight(buf.String(), ">"),
					Body:    bytes.ReplaceAll(escaped, escapedTag, openTag),
				}
				// escaped :=
				mode = End
				buf.Reset()
				cursor = 0
			}
		case End:
			if char != closeTag[cursor] {
				return nil, pos, errors.New("missing closing tag")
			}
			cursor++
			if cursor > len(closeTag)-1 {
				return msg, pos, nil
			}
		}
		prev = char
	}
}

type Muxer struct {
	w     io.Writer
	mutex sync.Mutex
}

func NewMuxer(w io.Writer) *Muxer {
	return &Muxer{w: w}
}
func (m *Muxer) Write(b []byte) (int, error) {
	m.mutex.Lock() // Serialize the writes
	defer m.mutex.Unlock()
	return m.w.Write(b)
}

type Channel struct {
	Name string
	m    *Muxer
}

func (m *Muxer) Channel(name string) io.Writer {
	return &Channel{Name: name, m: m}
}

func (c *Channel) Write(b []byte) (int, error) {
	if err := Write(c.m, c.Name, b); err != nil {
		return 0, err
	}
	return len(b), nil
}

type Scanner struct {
	r   io.Reader
	buf bytes.Buffer
	m   *Message
	err error
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: r}
}

func (s *Scanner) Scan() bool {
	if s.err != nil {
		return false
	}
	char := make([]byte, 1)
	for {
		if _, err := s.r.Read(char); err != nil {
			s.err = err
			return false
		}
		s.buf.Write(char)
		if s.buf.Len() > minMessageSize {
			m, n, err := Read(bytes.NewReader(s.buf.Bytes()))
			if err != nil {
				continue
			}
			s.buf.Next(n)
			s.m = m
			return true
		}
	}
}

func (s *Scanner) Message() *Message {
	return s.m
}
func (s *Scanner) Channel() string {
	if s.m == nil {
		return ""
	}
	return s.m.Channel
}
func (s *Scanner) Text() string {
	if s.m == nil {
		return ""
	}
	return string(s.m.Body)
}
func (s *Scanner) Bytes() []byte {
	if s.m == nil {
		return nil
	}
	return s.m.Body
}
func (s *Scanner) Err() error {
	return s.err
}

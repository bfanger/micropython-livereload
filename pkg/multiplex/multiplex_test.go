package multiplex

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrite(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		channel  string
		message  []byte
		expected []byte
	}{
		{name: "basic abc", channel: "X", message: []byte("abc"), expected: []byte("<MSG 3 X>abc</MSG>")},
		{name: "basic DEFGH", channel: "Y", message: []byte("DEFGH"), expected: []byte("<MSG 5 Y>DEFGH</MSG>")},
		{name: "escaping", channel: "Z", message: []byte("a<MSG b"), expected: []byte("<MSG 8 Z>a\\<MSG b</MSG>")},
	}

	for _, tc := range tests {
		t.Run(tc.name, assertWrite(tc.channel, tc.message, tc.expected))
	}
}
func assertWrite(channel string, message []byte, expected []byte) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)
		var buf bytes.Buffer
		err := Write(&buf, channel, message)
		assert.NoError(err)
		assert.Equal(expected, buf.Bytes())
	}
}
func TestRead(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		encoded  []byte
		expected *Message
		read     int
		err      error
	}{
		{name: "basic", encoded: []byte("<MSG 3 X>abc</MSG>"), expected: &Message{Channel: "X", Body: []byte("abc")}, read: 18},
		{name: "in stream", encoded: []byte("123 56<MSG 5 Y>DEFGH</MSG>"), expected: &Message{Channel: "Y", Body: []byte("DEFGH")}, read: 26},
		{name: "no message", encoded: []byte("just a string"), expected: nil, read: 13, err: io.EOF},
		{name: "escaping", encoded: []byte("<MSG 8 Z>a\\<MSG b</MSG>"), expected: &Message{Channel: "Z", Body: []byte("a<MSG b")}, read: 23, err: nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, assertRead(tc.encoded, tc.expected, tc.read, tc.err))
	}
}
func assertRead(encoded []byte, expectedMsg *Message, read int, expectedErr error) func(*testing.T) {
	return func(t *testing.T) {
		assert := require.New(t)
		msg, n, err := Read(bytes.NewReader(encoded))
		if expectedErr != nil {
			assert.Equal(expectedErr, err)
		} else {
			assert.NoError(err)
			assert.NotNil(expectedMsg)
			assert.Equal(expectedMsg.Channel, msg.Channel)
			assert.Equal(expectedMsg.Body, msg.Body)
		}
		assert.Equal(read, n)
	}
}
func TestMuxer(t *testing.T) {
	assert := require.New(t)
	var buf bytes.Buffer
	m := NewMuxer(&buf)
	c := m.Channel("IO")
	n, err := c.Write([]byte("Test"))
	assert.NoError(err)
	assert.Equal(4, n)
	n, err = c.Write([]byte("123"))
	assert.NoError(err)
	assert.Equal(3, n)
	assert.Equal("<MSG 4 IO>Test</MSG><MSG 3 IO>123</MSG>", buf.String())
}

func TestScanner(t *testing.T) {
	assert := require.New(t)
	r, w := io.Pipe()
	s := NewScanner(r)
	go func() {
		w.Write([]byte("bla bla <MSG 4 "))
		w.Write([]byte("IO>Test</MSG>other data<MSG 3 CHAN2>123</MSG><MSG 2 IO>Ok</M"))
		w.Write([]byte("SG>"))
		w.Close()
	}()
	assert.True(s.Scan())
	assert.NoError(s.Err())
	assert.Equal("IO", s.Channel())
	assert.Equal("Test", s.Text())

	assert.True(s.Scan())
	assert.NoError(s.Err())
	assert.Equal("CHAN2", s.Channel())
	assert.Equal("123", s.Text())

	assert.True(s.Scan())
	assert.NoError(s.Err())
	assert.Equal("IO", s.Channel())
	assert.Equal("Ok", s.Text())

	assert.False(s.Scan())
	assert.Error(io.EOF, s.Err())
}

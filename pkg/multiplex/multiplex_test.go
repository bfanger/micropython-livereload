package multiplex

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
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
	}

	for _, tc := range tests {
		t.Run(tc.name, assertWrite(tc.channel, tc.message, tc.expected))
	}
}
func assertWrite(channel string, message []byte, expected []byte) func(*testing.T) {
	return func(t *testing.T) {
		assert := assert.New(t)
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
	}
	for _, tc := range tests {
		t.Run(tc.name, assertRead(tc.encoded, tc.expected, tc.read, tc.err))
	}
}
func assertRead(encoded []byte, expectedMsg *Message, read int, expectedErr error) func(*testing.T) {
	return func(t *testing.T) {
		assert := assert.New(t)
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

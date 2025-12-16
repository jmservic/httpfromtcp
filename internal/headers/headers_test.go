package headers

import (
	//"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	//"io"
	"testing"
)

func TestHeadersParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("        HoSt:    localhost:42069      \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 40, n)
	assert.False(t, done)

	// Test Valid 2 headers with existing headers
	headers = NewHeaders()
	data = []byte("host: localhost:42069\r\n Content-Length: 348\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.False(t, done)
	n, done, err = headers.Parse(data[n:])
	require.NoError(t, err)
	assert.False(t, done)
	require.Equal(t, "localhost:42069", headers["host"])
	require.Equal(t, "348", headers["content-length"])
	require.Equal(t, 22, n)

	// Test: Invalid Characters
	headers = NewHeaders()
	data = []byte("<host>: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid done
	headers = NewHeaders()
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.True(t, done)
	assert.Equal(t, 2, n)

	// Test: Multiple Values
	headers = NewHeaders()
	headers.Set("Set-Person", "jonathan-loves-cpp")
	offset := 0
	data = []byte("Set-Person: lane-loves-go\r\n Set-Person: prime-loves-zig\r\n Set-Person: tj-loves-ocaml\r\n\r\n")
	n, done, err = headers.Parse(data)
	offset += n
	n, done, err = headers.Parse(data[offset:])
	offset += n
	n, done, err = headers.Parse(data[offset:])
	offset += n
	n, done, err = headers.Parse(data[offset:])
	require.NoError(t, err)
	assert.True(t, done)
	assert.Equal(t, "jonathan-loves-cpp, lane-loves-go, prime-loves-zig, tj-loves-ocaml", headers["set-person"])
}

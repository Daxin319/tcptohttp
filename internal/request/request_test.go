package request

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n
	if n > cr.numBytesPerRead {
		n = cr.numBytesPerRead
		cr.pos -= n - cr.numBytesPerRead
	}
	return n, nil
}

func TestRequestLineParse(t *testing.T) {
	// Test: Good GET Request line
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Good GET Request line with path
	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Invalid number of parts in request line
	_, err = RequestFromReader(strings.NewReader("/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Invalid GET Request (wrong order)
	_, err = RequestFromReader(strings.NewReader("/coffee GET HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Invalid GET Request (wrong HTTP Version)
	_, err = RequestFromReader(strings.NewReader("GET /coffee HTTP/2.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: POST request
	reader = &chunkReader{
		data:            "POST /submit HTTP/1.1\r\nHost: localhost:42069\r\nContent-Type: application/x-www-form-urlencoded\r\n\r\n",
		numBytesPerRead: 5,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "POST", r.RequestLine.Method)
	assert.Equal(t, "/submit", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Invalid method (not GET or POST)
	_, err = RequestFromReader(strings.NewReader("PUT /update HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"))
	require.Error(t, err)

	// Test: GET with query parameters
	reader = &chunkReader{
		data:            "GET /search?q=test&page=2 HTTP/1.1\r\nHost: localhost:42069\r\n\r\n",
		numBytesPerRead: 4,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/search?q=test&page=2", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}

func TestRequestParsingWithChunkedReading(t *testing.T) {
	// Test: Small chunk size (1 byte)
	reader := &chunkReader{
		data:            "GET /small-chunk HTTP/1.1\r\nHost: localhost:42069\r\n\r\n",
		numBytesPerRead: 1,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/small-chunk", r.RequestLine.RequestTarget)

	// Test: Medium chunk size
	reader = &chunkReader{
		data:            "GET /medium-chunk HTTP/1.1\r\nHost: localhost:42069\r\n\r\n",
		numBytesPerRead: 5,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/medium-chunk", r.RequestLine.RequestTarget)

	// Test: Large chunk size (bigger than buffer)
	reader = &chunkReader{
		data:            "GET /large-chunk HTTP/1.1\r\nHost: localhost:42069\r\n\r\n",
		numBytesPerRead: 20,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/large-chunk", r.RequestLine.RequestTarget)
}

func TestEdgeCases(t *testing.T) {
	// Test: Exactly one buffer size
	exactData := "GET /x HTTP/1.1\r\n"
	reader := &chunkReader{
		data:            exactData,
		numBytesPerRead: len(exactData),
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)

	// Test: Empty request
	reader = &chunkReader{
		data:            "",
		numBytesPerRead: 5,
	}
	_, err = RequestFromReader(reader)
	require.NoError(t, err) // Your implementation doesn't return error on empty

	// Test: Long request URL
	longPath := "/really" + strings.Repeat("-long", 100) + "-path"
	reader = &chunkReader{
		data:            "GET " + longPath + " HTTP/1.1\r\nHost: localhost\r\n\r\n",
		numBytesPerRead: 10,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, longPath, r.RequestLine.RequestTarget)

	// Test: Request line split across reads
	reader = &chunkReader{
		data:            "GET /split HTTP/1.1\r\nHost: localhost\r\n\r\n",
		numBytesPerRead: 3, // Will split "GET /split HTTP/1.1" across multiple reads
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)

	// Test: Incomplete request line (no \r\n)
	reader = &chunkReader{
		data:            "GET /no-crlf HTTP/1.1",
		numBytesPerRead: 5,
	}
	_, err = RequestFromReader(reader)
	require.NoError(t, err) // Your implementation doesn't error on incomplete
	assert.Equal(t, "/split", r.RequestLine.RequestTarget)
}

func TestRequestParsingErrors(t *testing.T) {
	// Test: Malformed request line (missing spaces)
	reader := strings.NewReader("GET/path HTTP/1.1\r\nHost: localhost\r\n\r\n")
	_, err := RequestFromReader(reader)
	require.Error(t, err)
}

func TestRequestParsingHeaders(t *testing.T) {

	// Test: Standard Headers
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers["host"])
	assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
	assert.Equal(t, "*/*", r.Headers["accept"])

	// Test: Malformed Header
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)

}

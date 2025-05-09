package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaderParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 25, n)
	assert.True(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 38, n)
	assert.False(t, done)

	// Test: Valid header with extra whitespace
	headers = NewHeaders()
	data = []byte("Host:    localhost:42069    \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, len(data), n)
	assert.True(t, done)

	// Test: Two Headers with existing map entry
	headers = NewHeaders()
	headers["host"] = "original"
	data = []byte("User-Agent: test\r\nAccept: */*\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "original", headers["host"])
	assert.Equal(t, "test", headers["user-agent"])
	assert.Equal(t, "*/*", headers["accept"])
	assert.Equal(t, len(data), n)
	assert.True(t, done)

	// Test: Invalid no colon
	headers = NewHeaders()
	data = []byte("InvalidHeader\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 15, n)
	assert.False(t, done)

	// Test: Invalid empty field-name
	headers = NewHeaders()
	data = []byte(": value\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 9, n)
	assert.False(t, done)

	// Test: Invalid, spaces before colon in field-name
	headers = NewHeaders()
	data = []byte("Host : localhost\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 18, n)
	assert.False(t, done)

	// Test: Field name with space (invalid)
	headers = NewHeaders()
	data = []byte("User Agent: test\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 18, n)
	assert.False(t, done)

	// Test: Field name with tab (invalid)
	headers = NewHeaders()
	data = []byte("User\tAgent: test\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 18, n)
	assert.False(t, done)

	// Test: Field name with invalid symbol (@)
	headers = NewHeaders()
	data = []byte("User@Agent: test\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 18, n)
	assert.False(t, done)

	// Test: Field name with slash (invalid)
	headers = NewHeaders()
	data = []byte("User/Agent: test\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 18, n)
	assert.False(t, done)

	// Test: Field name with equal sign (invalid)
	headers = NewHeaders()
	data = []byte("User=Agent: test\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 18, n)
	assert.False(t, done)

	// Test: Field name with allowed special characters (sanity check)
	headers = NewHeaders()
	data = []byte("X-Test_123!#$%&'*+-.^_`|~: value\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "value", headers["x-test_123!#$%&'*+-.^_`|~"])
	assert.Equal(t, len(data), n)
	assert.True(t, done)

	// Test: Field name with emoji
	headers = NewHeaders()
	data = []byte("🔥-Header: wow\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 18, n)
	assert.False(t, done)

	// Test: Field name with non-ASCII character
	headers = NewHeaders()
	data = []byte("Üser-Agent: test\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 19, n)
	assert.False(t, done)

	// Test: Emoji in field value (should be allowed)
	headers = NewHeaders()
	data = []byte("x-note: approved ✅\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "approved ✅", headers["x-note"])
	assert.Equal(t, len(data), n)
	assert.True(t, done)

	// Test: Tab in field value (allowed by spec)
	headers = NewHeaders()
	data = []byte("x-tabbed: value\twith\ttabs\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "value\twith\ttabs", headers["x-tabbed"])
	assert.Equal(t, len(data), n)
	assert.True(t, done)

	// Test: Field name with emoji in the middle
	headers = NewHeaders()
	data = []byte("x🔥note: nope\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 17, n)
	assert.False(t, done)

	// Test: Multiple headers with same field-name (should concat)
	headers = NewHeaders()
	data = []byte("lang-pref: tj-likes-ocaml;\r\nlang-pref: prime-likes-zig;\r\nlang-pref: lane-likes-go;\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "tj-likes-ocaml;, prime-likes-zig;, lane-likes-go;", headers["lang-pref"])
	assert.Equal(t, len(data), n)
	assert.True(t, done)

	// Test: Mixed single and multi-value headers
	headers = NewHeaders()
	data = []byte("user-agent: curl\r\nlang-pref: tj-likes-ocaml;\r\nlang-pref: prime-likes-zig;\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "curl", headers["user-agent"])
	assert.Equal(t, "tj-likes-ocaml;, prime-likes-zig;", headers["lang-pref"])
	assert.Equal(t, len(data), n)
	assert.True(t, done)

	// Test: Append to existing field in map
	headers = NewHeaders()
	headers["lang-pref"] = "initial"
	data = []byte("lang-pref: second\r\nlang-pref: third\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "initial, second, third", headers["lang-pref"])
}

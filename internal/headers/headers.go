package headers

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

type Headers map[string]string

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	total := 0

	for {
		if len(data) >= 2 && data[0] == '\r' && data[1] == '\n' {
			return total + 2, true, nil
		}

		idx := bytes.Index(data, []byte("\r\n"))
		if idx == -1 {
			return total, false, nil
		}

		line := string(data[:idx])
		data = data[idx+2:]
		total += idx + 2

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		colonIndex := strings.Index(line, ":")
		if colonIndex <= 0 {
			return total, false, fmt.Errorf("invalid header format: %s", line)
		}

		key := line[:colonIndex]
		value := strings.TrimLeft(line[colonIndex+1:], " ")

		regex := regexp.MustCompile(`^[A-Za-z0-9!#$%&'*+\-.\^_` + "`" + `|~]+$`)
		if !regex.MatchString(key) {
			return total, false, fmt.Errorf("invalid characters in header field name: %s", key)
		}

		lowerKey := strings.ToLower(key)
		if prev, exists := h[lowerKey]; exists {
			h[lowerKey] = prev + ", " + value
		} else {
			h[lowerKey] = value
		}

	}
}

func NewHeaders() Headers {

	return make(Headers)

}

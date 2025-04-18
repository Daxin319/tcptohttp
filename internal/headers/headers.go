package headers

import (
	"fmt"
	"regexp"
	"strings"
)

type Headers map[string]string

func (h Headers) Parse(data []byte) (n int, done bool, err error) {

	split := strings.Split(string(data), "\r\n")
	total := len(data) - len("\r\n")

	if split[0] == "\r\n" {
		return total, true, nil
	}

	trimLeft := strings.TrimLeft(split[0], " ")
	trimRight := strings.TrimRight(trimLeft, " ")

	for i := range trimRight {

		if string(trimRight[i]) == " " || string(trimRight[i]) == "" {
			return 0, false, fmt.Errorf("Invalid header format - %s\n", string(data))
		}

		if i == 0 && string(trimRight[i]) == ":" {
			return 0, false, fmt.Errorf("Invalid header format - %s\n", string(data))
		}

		if string(trimRight[i]) == ":" {
			regex := regexp.MustCompile(`^[A-Za-z0-9!#$%&'*+\-.\^_` + "`" + `|~]+$`)
			key := string(trimRight[:i])

			if !regex.MatchString(key) {
				return 0, false, fmt.Errorf("Invalid characters in field name: %s\n", key)
			}

			fmtKey := strings.ToLower(key)

			tempValue := string(trimRight[i+1:])
			splitTempValue := strings.Split(tempValue, "\r\n")
			value := strings.TrimLeft(splitTempValue[0], " ")

			if h[fmtKey] != "" {
				h[fmtKey] += ", " + value
			} else {
				h[fmtKey] = value
			}

			start := len(split[0]) + 2

			_, _, err := h.Parse(data[start:])
			if err != nil {
				return 0, false, fmt.Errorf("Malformed header: %s\n", string(data[start:]))
			}

			break
		}

		if i == len(trimRight)-1 && string(trimRight[i]) != ":" {
			return 0, false, fmt.Errorf("Invalid header format - %s\n", string(data))
		}
	}
	return total, false, nil
}

func NewHeaders() Headers {

	return make(Headers)

}

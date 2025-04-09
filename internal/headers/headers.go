package headers

import (
	"fmt"
	"strings"
)

type Headers map[string]string

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	l := strings.Split(string(data), "\r\n")
	total := len(data) - len("\r\n")
	if l[0] == "\r\n" {
		return total, true, nil
	}
	t := strings.TrimLeft(l[0], " ")
	tf := strings.TrimRight(t, " ")
	for i := range tf {
		if string(tf[i]) == " " || string(tf[i]) == "" {
			return 0, false, fmt.Errorf("Invalid header format - %s\n", string(data))
		}
		if i == 0 && string(tf[i]) == ":" {
			return 0, false, fmt.Errorf("Invalid header format - %s\n", string(data))
		}
		if string(tf[i]) == ":" {
			k := string(tf[:i])
			tv := string(tf[i+1:])
			stv := strings.Split(tv, "\r\n")
			v := strings.TrimLeft(stv[0], " ")
			h[k] = v
			start := len(l[0]) + 2
			_, _, err := h.Parse(data[start:])
			if err != nil {
				return 0, false, fmt.Errorf("Malformed header: %s\n", string(data[start:]))
			}
			break
		}
		if i == len(tf)-1 && string(tf[i]) != ":" {
			return 0, false, fmt.Errorf("Invalid header format - %s\n", string(data))
		}
	}
	return total, false, nil
}

func NewHeaders() Headers {
	return make(Headers)
}

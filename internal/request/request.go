package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

type Status int

const (
	initialized Status = iota
	done
)

const bufferSize = 8

type Request struct {
	RequestLine  RequestLine
	ParserStatus Status
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize, bufferSize)
	readToIndex := 0
	req := &Request{
		ParserStatus: 0,
	}

	for req.ParserStatus == 0 {
		if readToIndex >= len(buf)-1 {
			newBuf := make([]byte, len(buf)*2, cap(buf)*2)
			_ = copy(newBuf, buf)
			buf = newBuf
		}
		r, err := reader.Read(buf)
		if err == io.EOF {
			req.ParserStatus = 1
			break
		}
		readToIndex += r
		n, err := req.parse(buf)
		if err != nil {
			fmt.Printf("Error parsing data: %v\n", err)
			return nil, err
		}
		readToIndex -= n
	}
	var clear []byte
	_ = copy(buf, clear)

	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.ParserStatus {
	case 0:
		r, n, err := parseRequestLine(data)
		if err != nil {
			fmt.Printf("error parsing data: %v\n", err)
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.ParserStatus = 1
		return n, nil
	case 1:
		return 0, errors.New("Error: Attempting to parse data in done state")
	default:
		return 0, errors.New("Error: Unknown State")
	}
}

func parseRequestLine(data []byte) (*Request, int, error) {
	if !strings.Contains(string(data), "\r\n") {
		return nil, 0, nil
	}
	split := strings.Split(string(data), "\r\n")
	rLineStr := split[0]
	lineSplit := strings.Split(rLineStr, " ")
	if len(lineSplit) == 1 {
		return nil, 0, nil
	}
	if len(lineSplit) != 3 {
		fmt.Printf("error reading data: %v\n", rLineStr)
		return nil, len(data), errors.New("invalid request format")
	}
	if lineSplit[2] != "HTTP/1.1" {
		fmt.Printf("invalid HTTP version: %s\n", lineSplit[2])
		return nil, len(data), errors.New("Invalid HTTP Version:")
	}
	if lineSplit[0] != "GET" && lineSplit[0] != "POST" {
		fmt.Printf("invalid HTTP Method: %s\n", lineSplit[0])
		return nil, len(data), errors.New("Invalid HTTP Method")
	}

	rLine := RequestLine{
		HttpVersion:   lineSplit[2][5:],
		RequestTarget: lineSplit[1],
		Method:        lineSplit[0],
	}

	request := &Request{
		RequestLine: rLine,
	}

	return request, len(data), nil

}

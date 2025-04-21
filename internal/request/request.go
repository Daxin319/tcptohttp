package request

import (
	"errors"
	"fmt"
	"io"
	"main/internal/headers"
	"strings"
)

type Status int

const (
	initialized Status = iota
	done
	requestStateParsingHeaders
)

const bufferSize = 8

type Request struct {
	RequestLine  RequestLine
	Headers      headers.Headers
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
		Headers:      headers.NewHeaders(),
		ParserStatus: initialized,
	}

	for req.ParserStatus != done {

		if readToIndex >= len(buf)-1 {
			newBuf := make([]byte, len(buf)*2)
			_ = copy(newBuf, buf[:readToIndex])
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		if n > 0 {
			readToIndex += n
		}

		if err == io.EOF {
			if req.ParserStatus != done {
				return nil, fmt.Errorf("Unexpected EOF: Headers not terminated properly")
			}
			break
		}

		parsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			fmt.Printf("Error parsing data: %v\n", err)
			return nil, err
		}
		if parsed > 0 {
			copy(buf, buf[parsed:readToIndex])
		}
		readToIndex -= parsed

	}
	var clear []byte
	_ = copy(buf, clear)

	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.ParserStatus != done {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, fmt.Errorf("error parsing request: %v\n", err)
		}
		if n == 0 {
			return totalBytesParsed, nil
		}
		totalBytesParsed += n
	}
	return totalBytesParsed, nil
}

func parseRequestLine(data []byte) (*Request, int, error) {
	if !strings.Contains(string(data), "\r\n") {
		return nil, 0, nil
	}

	split := strings.SplitN(string(data), "\r\n", 2)
	if len(split) < 2 {
		return nil, 0, nil
	}

	rLineStr := split[0]
	lineSplit := strings.Split(rLineStr, " ")
	var request *Request

	switch true {

	case len(lineSplit) == 1:
		return nil, 0, nil

	case len(lineSplit) != 3:
		fmt.Printf("error reading data: %v\n", rLineStr)
		return nil, len(data), errors.New("invalid request format")

	case lineSplit[2] != "HTTP/1.1":
		fmt.Printf("invalid HTTP version: %s\n", lineSplit[2])
		return nil, len(data), errors.New("Invalid HTTP Version:")

	case lineSplit[0] != "GET" && lineSplit[0] != "POST":
		fmt.Printf("invalid HTTP Method: %s\n", lineSplit[0])
		return nil, len(data), errors.New("Invalid HTTP Method")

	default:
		rLine := RequestLine{
			HttpVersion:   lineSplit[2][5:],
			RequestTarget: lineSplit[1],
			Method:        lineSplit[0],
		}

		request = &Request{
			RequestLine: rLine,
		}

	}

	return request, len(split[0]) + len("\r\n"), nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.ParserStatus {

	case initialized:
		if strings.Contains(string(data), "\r\n") {
			newReq, n, err := parseRequestLine(data)
			if err != nil {
				fmt.Printf("error parsing data: %v\n", err)
				return 0, err
			}
			if n == 0 {
				return 0, nil
			}
			if newReq != nil {
				r.RequestLine = newReq.RequestLine
				r.ParserStatus = requestStateParsingHeaders
			}
			return n, nil
		}
		return 0, nil

	case done:
		return 0, errors.New("Error: Attempting to parse data in done state")

	case requestStateParsingHeaders:
		if strings.Contains(string(data), "\r\n") {
			h := r.Headers
			if r.Headers == nil {
				return 0, errors.New("r.Headers is nil")
			}
			n, complete, err := h.Parse(data)
			if err != nil {
				return 0, fmt.Errorf("error parsing header field-lines: %v\n", err)
			}
			if n == 0 {
				return 0, nil
			}
			if complete {
				r.ParserStatus = done
				r.Headers = h
			}
			return n, nil
		}
		return 0, nil

	default:
		return 0, errors.New("Error: Unknown State")

	}
}

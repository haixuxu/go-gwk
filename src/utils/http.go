package utils

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
)

type GwkHttpRequest struct {
	Method    string
	Path      string
	Version   string
	Headers   map[string]string
	RawBuffer []byte
}

func ParseHttpHeader(conn net.Conn) (req *GwkHttpRequest, err error) {
	// 创建缓冲读取器以从连接中读取数据
	reader := io.Reader(conn)
	// 读取第一行，解析请求方法和路径
	requestLine, err := ReadOneLine(reader)
	if err != nil {
		return nil, errors.New("unnormal http request!0x1")
	}

	// 解析请求行
	parts := strings.Split(requestLine, " ")
	if len(parts) < 3 {
		return nil, errors.New("unnormal http request!0x2")
	}
	method := parts[0]
	path := parts[1]
	version := parts[2]

	cache := []byte(fmt.Sprintf("%s %s %s\r\n", method, path, parts[2]))
	// 解析请求头部
	headers := make(map[string]string)
	for {
		line, err := ReadOneLine(reader)
		if err != nil || line == "" {
			break
		}
		cache = append(cache, []byte(line+"\r\n")...)
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			headerName := strings.TrimSpace(parts[0])
			headerValue := strings.TrimSpace(parts[1])
			headers[headerName] = headerValue
		}
	}
	req = &GwkHttpRequest{
		Method:    method,
		Path:      path,
		Version:   version,
		Headers:   headers,
		RawBuffer: cache,
	}
	return req, nil
}

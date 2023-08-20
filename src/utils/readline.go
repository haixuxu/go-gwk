package utils

import "io"

func ReadOneLine(reader io.Reader) (string, error) {
	buf := make([]byte, 0, 256)
	for {
		b := make([]byte, 1)
		_, err := reader.Read(b)
		if err != nil {
			return "", err
		}
		if b[0] == '\r' {
			continue
		}
		if b[0] == '\n' {
			break
		}
		buf = append(buf, b[0])
	}
	return string(buf), nil
}

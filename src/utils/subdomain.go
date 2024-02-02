package utils

import (
	"math/rand"
	"time"
)

const CHARS = "0123456789abcdefghighlmnopqrstwvuxyz"

func GenSubdomain() string {
	rand.Seed(time.Now().UnixNano())

	str := string(CHARS[10+rand.Intn(26)])
	for i := 0; i < 15; i++ {
		char := string(CHARS[rand.Intn(36)])
		str += char
	}

	return str
}

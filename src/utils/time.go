package utils

import (
	"fmt"
	"time"
)

//func GetNowTimestrapInt() uint64 {
//	timest := time.Now().UnixNano() / 1e6
//	return uint64(timest)
//	//tstr := fmt.Sprintf("%v", timest)
//}

func GetNowInt64Bytes() []byte {
	timest := time.Now().UnixNano() / 1e6
	tstr := fmt.Sprintf("%v", timest)
	// tstr := strconv.Itoa(int(timest))
	return []byte(tstr)
}

func GetNowInt64String() string {
	timest := time.Now().UnixNano() / 1e6
	tstr := fmt.Sprintf("%v", timest)
	return tstr
}

package protocol

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"time"
)

const (
	AUTH_REQ = 0x0 // start auth
	AUTH_RES = 0x1 // auth response

	TUNNEL_REQ = 0xa6 // start stub
	TUNNEL_RES = 0xa9 // response stub

	PING_FRAME = 0x6
	PONG_FRAME = 0x9

	STREAM_INIT = 0xf0
	STREAM_EST  = 0xf1
	STREAM_DATA = 0xf2
	STREAM_FIN  = 0xf3
	STREAM_RST  = 0xf4
)

/**
 * // required: type, token
 * @param {*} AUTH_REQ frame
 * |<--type[1]-->|--status(1)--|<------auth token------>|
 * |----- 1 -----|------0------|----------s2------------|
 *
 * @param {*} AUTH_RES frame
 * |<--type[1]-->|--status(1)--|<--------message------->|
 * |----- 1 -----|-----1/2-----|----------s2------------|
 *
 * @param {*} PING frame
 * |<--type[1]-->|----stime--|
 * |----- 1 -----|-----8-----|
 * @param {*} PONG frame
 * |<--type[1]-->|----stime--|----atime----|
 * |----- 1 -----|-----8-----|-----8-------|
 *
 * @param {*} TUNNEL_REQ frame
 * |<--type[1]-->|----pro----|----- port/subdomain-----|
 * |----- 1 -----|----- 1----|--------name:port--------|
 * |----- 1 -----|----- 1----|--------name:domain------|
 *
 * @param {*} TUNNEL_RES frame
 * |<--type[1]-->|----status----|------message-------|
 * |----- 1 -----|----- 1-------|--------------------|
 *
 * @param {*} STREAM_INIT frame
 * |<--type[1]-->|----stream id----|
 * |----- 1 -----|------- 16-------|
 * @param {*} STREAM_EST frame
 * |<--type[1]-->|----stream id----|
 * |----- 1 -----|------- 16-------|
 *
 * @param {*} STREAM_DATA frame
 * |<--type[1]-->|----stream id----|-------data--------|
 * |----- 1 -----|------- 16-------|-------------------|
 *
 * @param {*} STREAM_RST frame
 * |<--type[1]-->|----stream id----|
 * |----- 1 -----|------- 16-------|------message------|
 *
 * @param {*} STREAM_FIN frame
 * |<--type[1]-->|----stream id----|
 * |----- 1 -----|------- 16-------|
 * @returns
 */

type Frame struct {
	Type     uint8
	Stime    uint64
	Atime    uint64
	StreamID string
	Data     []byte
}

func GetNowTimestrapInt() uint64 {
	timest := time.Now().UnixNano() / 1e6
	return uint64(timest)
	//tstr := fmt.Sprintf("%v", timest)
}

func GetTimestrapBytes(timestamp uint64) []byte {
	byteArray := make([]byte, 8)
	binary.BigEndian.PutUint64(byteArray, timestamp) // 将时间戳存储为8个字节的字节数组
	return byteArray
}

func Encode(frame *Frame) []byte {
	if frame.Type == PING_FRAME {
		prefix := []byte{frame.Type}
		stime := GetTimestrapBytes(frame.Stime)
		return append(prefix, stime...)

	} else if frame.Type == PONG_FRAME {
		prefix := []byte{frame.Type}
		stime := GetTimestrapBytes(frame.Stime)
		atime := GetTimestrapBytes(frame.Atime)
		return append(append(prefix, stime...), atime...)
	} else {
		prefix := []byte{frame.Type}
		cidbuf, _ := hex.DecodeString(frame.StreamID)
		buf := append(prefix, cidbuf...)
		if frame.Data == nil {
			return buf
		} else {
			return append(buf, frame.Data...)
		}
	}
}

func Decode(data []byte) (frame *Frame, err error) {
	typeVal := data[0]

	if typeVal == PING_FRAME {
		stByte := data[1:9]
		st := binary.BigEndian.Uint64(stByte)
		return &Frame{Type: typeVal, Stime: st}, nil
	} else if typeVal == PONG_FRAME {
		stime := binary.BigEndian.Uint64(data[1:9])
		atime := binary.BigEndian.Uint64(data[9:17])
		return &Frame{Type: typeVal, Stime: stime, Atime: atime}, nil
	} else if typeVal >= 0xf0 && typeVal <= 0xf4 {
		streamID := hex.EncodeToString(data[1:17])
		dataBuf := data[17:]
		return &Frame{Type: typeVal, StreamID: streamID, Data: dataBuf}, nil
	} else {
		return nil, errors.New("invalid packet!")
	}
}

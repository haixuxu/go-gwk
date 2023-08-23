package protocol

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	AUTH_REQ = 0x0 // start auth
	AUTH_RES = 0x1 // auth response

	TUNNEL_REQ = 0xa6 // start tunnel
	TUNNEL_RES = 0xa9 // response tunnel

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
 * |<--type[1]-->|--status(1)--|<------auth token(32)------>|
 * |----- 1 -----|------0------|--------------s2------------|
 *
 * @param {*} AUTH_RES frame
 * |<--type[1]-->|--status(1)--|<------auth token(32)------>|
 * |----- 1 -----|-----1/2-----|--------------s2------------|
 *

 *
 * @param {*} TUNNEL_REQ frame
 * |<--type[1]-->|----pro----|----port/subdomain----|
 * |----- 1 -----|----- 1----|--------name:port--------|
 * |----- 1 -----|----- 1----|--------name:domain------|
 *
 * @param {*} TUNNEL_RES frame
 * |<--type[1]-->|----status----|------message-------|
 * |----- 1 -----|----- 1-------|--------------------|
 * @param {*} PING frame
 * |<--type[1]-->|----stime---|
 * |----- 1 -----|-----13-----|
 * @param {*} PONG frame
 * |<--type[1]-->|----stime---|-----atime-----|
 * |----- 1 -----|---- 13-----|-----13--------|
 *
 * @param {*} STREAM_INIT frame
 * |<--type[1]-->|----stream id----|
 * |-----1 -----|------- 16-------|
 * @param {*} STREAM_EST frame
 * |<--type[1]-->|----stream id----|
 * |-----1 -----|------- 16-------|
 *
 * @param {*} STREAM_DATA frame
 * |<--type[1]-->|----stream id----|-------data--------|
 * |-----1 -----|------- 16-------|-------------------|
 *
 * @param {*} STREAM_RST frame
 * |<--type[1]-->|----stream id----|
 * |-----1 -----|------- 16-------|
 *
 * @param {*} STREAM_FIN frame
 * |<--type[1]-->|----stream id----|
 * |-----1 -----|------- 16-------|
 * @returns
 */

type Frame struct {
	Type   uint8
	Token  string
	Status uint8
	Stime  string
	Atime  string
	// tunnel req/res
	TunType   uint8
	Port      uint16
	Subdomain string
	Name      string

	Message string

	StreamID string
	Data     []byte
}

func Encode(frame *Frame) []byte {
	if frame.Type == AUTH_REQ || frame.Type == AUTH_RES {
		prefix := []byte{frame.Type, frame.Status}
		token := []byte(frame.Token)
		return append(prefix, token...)
	} else if frame.Type == TUNNEL_REQ {

		var message string
		if frame.TunType == 0x1 {
			message = fmt.Sprintf("%s:%d", frame.Name, frame.Port)
		} else {
			message = fmt.Sprintf("%s:%s", frame.Name, frame.Subdomain)
		}

		prefix := []byte{frame.Type, frame.TunType}

		return append(prefix, []byte(message)...)
	} else if frame.Type == TUNNEL_RES {
		prefix := []byte{frame.Type, frame.Status}

		messageBuf := []byte(frame.Message)
		return append(prefix, messageBuf...)
	} else if frame.Type == PING_FRAME {
		prefix := []byte{frame.Type}
		stime := []byte(frame.Stime)
		return append(prefix, stime...)

	} else if frame.Type == PONG_FRAME {
		prefix := []byte{frame.Type}
		stime := []byte(frame.Stime)
		atime := []byte(frame.Atime)
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

	if typeVal == AUTH_REQ || typeVal == AUTH_RES {
		token := string(data[2:])
		status := uint8(0)
		if typeVal == AUTH_RES {
			status = data[1]
		}
		return &Frame{Type: typeVal, Status: status, Token: token}, nil
	} else if typeVal == PING_FRAME {
		stime := string(data[1:14])

		return &Frame{Type: typeVal, Stime: stime}, nil
	} else if typeVal == PONG_FRAME {
		stime := string(data[1:14])
		atime := string(data[14:27])
		return &Frame{Type: typeVal, Stime: stime, Atime: atime}, nil
	} else if typeVal == TUNNEL_REQ {
		proto := data[1]
		message := string(data[2:])
		parts := strings.Split(message, ":")
		port := uint16(0)
		subdomain := ""
		if proto == 0x1 {
			num, err := strconv.ParseUint(parts[1], 10, 16)
			if err != nil {
				fmt.Println("无法转换为 uint16")
				return nil, errors.New("无法转换为 uint16")
			}
			uint16Val := uint16(num)
			port = uint16Val
		} else {
			subdomain = parts[1]
		}
		return &Frame{Type: typeVal, Name: parts[0], TunType: proto, Port: port, Subdomain: subdomain}, nil
	} else if typeVal == TUNNEL_RES {
		status := data[1]
		message := string(data[2:])
		return &Frame{Type: typeVal, Status: status, Message: message}, nil
	} else {
		streamID := hex.EncodeToString(data[1:17])
		dataBuf := data[17:]
		return &Frame{Type: typeVal, StreamID: streamID, Data: dataBuf}, nil
	}
}

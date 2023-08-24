package auth

import (
	"errors"
	"github/xuxihai123/go-gwk/v1/src/protocol"
	"github/xuxihai123/go-gwk/v1/src/transport"
	"github/xuxihai123/go-gwk/v1/src/types"
)

type AuthFunc func(authstr string) *types.StatusMsg

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
 */
func buildPacket(protype uint8, status uint8, msg string) []byte {
	prefix := []byte{protype, status}
	msgbuf := []byte(msg)
	ret := append(prefix, msgbuf...)
	return ret
}

func HandleAuthReq(tsport *transport.TcpTransport, authtoken string) error {
	pack := buildPacket(protocol.AUTH_REQ, 0x0, authtoken)
	err := tsport.SendPacket(pack)
	if err != nil {
		return err
	}
	packet, err := tsport.ReadPacket()
	if err != nil {
		return errors.New("protocol error!")
	}

	if packet[0] != protocol.AUTH_RES {
		return errors.New("protocol error1!")
	}
	if packet[1] != 0x1 {
		return errors.New("tunnel auth failed!")
	}
	return nil
}

func HandleAuthRes(tsport *transport.TcpTransport, handler AuthFunc) error {
	packet, err := tsport.ReadPacket()
	if err != nil {
		return errors.New("protocol error!")
	}
	if packet[0] != protocol.AUTH_REQ {
		return errors.New("protocol error!")
	}
	authstr := string(packet[2:])

	stausMsg := handler(authstr)

	pack := buildPacket(protocol.AUTH_RES, stausMsg.Status, stausMsg.Message)

	err = tsport.SendPacket(pack)
	if err != nil {
		return err
	}
	return nil
}

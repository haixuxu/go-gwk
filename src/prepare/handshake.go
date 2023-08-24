package prepare

import (
	"errors"
	"fmt"
	"github/xuxihai123/go-gwk/v1/src/protocol"
	"github/xuxihai123/go-gwk/v1/src/transport"
	"github/xuxihai123/go-gwk/v1/src/types"
	"strconv"
	"strings"
)

type ReqTunFun func(opt *types.TunnelOpts) *types.StatusMsg

/**
 * @param {*} TUNNEL_REQ frame
 * |<--type[1]-->|----pro----|----port/subdomain----|
 * |----- 1 -----|----- 1----|--------name:port--------|
 * |----- 1 -----|----- 1----|--------name:domain------|
 *
 * @param {*} TUNNEL_RES frame
 * |<--type[1]-->|----status----|------message-------|
 * |----- 1 -----|----- 1-------|--------------------|
 */

func HandleTunnelReq(tsport *transport.TcpTransport, opts *types.TunnelOpts) (msg string, err error) {
	prefix := []byte{protocol.TUNNEL_REQ}
	if opts.Type == "tcp" {
		prefix = append(prefix, 0x1)
		msgbuf := []byte(fmt.Sprintf("%s:%d", opts.Name, opts.RemotePort))
		prefix = append(prefix, msgbuf...)
	} else {
		prefix = append(prefix, 0x2)
		msgbuf := []byte(fmt.Sprintf("%s:%s", opts.Name, opts.Subdomain))
		prefix = append(prefix, msgbuf...)
	}

	err = tsport.SendPacket(prefix)
	if err != nil {
		return "", err
	}

	packet, err := tsport.ReadPacket()
	if err != nil {
		return "", errors.New("protocol error!")
	}

	if packet[0] != protocol.TUNNEL_RES {
		return "", errors.New("protocol error1!")
	}
	if packet[1] != 0x1 {
		return "", errors.New("tunnel prepare failed!")
	}

	message := string(packet[2:])
	return message, nil
}

func HandleTunnelRes(tsport *transport.TcpTransport, handler ReqTunFun) error {
	packet, err := tsport.ReadPacket()
	if err != nil {
		return errors.New("protocol error!")
	}
	if packet[0] != protocol.TUNNEL_REQ {
		return errors.New("protocol error!")
	}
	proto := packet[1]
	message := string(packet[2:])
	parts := strings.Split(message, ":")
	port := uint16(0)
	subdomain := ""
	if proto == 0x1 {
		num, err := strconv.ParseUint(parts[1], 10, 16)
		if err != nil {
			return errors.New("protocol error!")
		}
		uint16Val := uint16(num)
		port = uint16Val
	} else {
		subdomain = parts[1]
	}
	tunopts := types.TunnelOpts{
		Name:       parts[0],
		Type:       types.GetTypeByNo(proto),
		LocalPort:  0,
		RemotePort: int(port),
		Subdomain:  subdomain,
	}

	stausMsg := handler(&tunopts)

	prefix := []byte{protocol.TUNNEL_RES, stausMsg.Status}
	msgbuf := []byte(stausMsg.Message)
	pack := append(prefix, msgbuf...)

	err = tsport.SendPacket(pack)
	if err != nil {
		return err
	}
	return nil
}

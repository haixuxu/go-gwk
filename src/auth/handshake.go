package auth

import (
	"errors"
	"fmt"
	 "github/xuxihai123/go-gwk/v1/src/protocol"
	"github/xuxihai123/go-gwk/v1/src/transport"
)

// TODO

func HandleAuthReq(authtoken string, tsport *transport.TcpTransport) error {

	authReqFm := protocol.Frame{Type:protocol.AUTH_REQ,Status:0x0,Token:authtoken}
	tsport.SendPacket(protocol.Encode(&authReqFm))
	packet, err := tsport.ReadPacket()
	if err != nil {
		fmt.Println("transport read packet err;", err.Error())
		return errors.New("transport read packet error!")
	}

	fm,err := protocol.Decode(packet)
	if err!= nil {
		return errors.New("protocol error1!")
	}


	if fm.Status!=1{
		return errors.New("tunnel auth failed!")
	}
	return nil
}

func HandleAuthRes(authtoken string, tsport *transport.TcpTransport)  {
	packet, err := tsport.ReadPacket()
	if err != nil {
		fmt.Println("transport read packet err;", err.Error())
		return
	}
	fm,err := protocol.Decode(packet)
	if err!= nil {
		fmt.Errorf("protol error\n")
		return
	}

	if(fm.Type==protocol.AUTH_REQ){
		fmt.Println("auth req token:",fm.Token, authtoken)

		authResFm := protocol.Frame{Type:protocol.AUTH_RES,Status:0x1,Token:authtoken}
		tsport.SendPacket(protocol.Encode(&authResFm))
	}
}
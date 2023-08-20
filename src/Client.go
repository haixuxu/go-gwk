package gwk

import (
	"errors"
	"fmt"
	"github.com/bbk47/toolbox"
	"github/xuxihai123/go-gwk/v1/src/auth"
	"github/xuxihai123/go-gwk/v1/src/protocol"
	"github/xuxihai123/go-gwk/v1/src/transport"
	"github/xuxihai123/go-gwk/v1/src/tunnel"
	"github/xuxihai123/go-gwk/v1/src/utils"
	"log"
	"net"
	"strconv"
	"sync"
)

type Client struct {
	opts   *ClientOpts
	logger *toolbox.Logger
	// inner attr
	tunnelStatus uint8
}

func NewClient(opts *ClientOpts) Client {
	cli := Client{}

	cli.opts = opts
	cli.logger = utils.NewLogger("C", opts.LogLevel)
	return cli
}

func (cli *Client) handlePrepare(tunopts *TunnelConfig, tsport *transport.TcpTransport) (msg string, err error) {
	protype := 0x1
	if tunopts.Protocol == "web" {
		protype = 0x2
	}
	tunnelReqFm := &protocol.Frame{Type: protocol.TUNNEL_REQ, Protocol: uint8(protype), Name: tunopts.Name, Port: uint16(tunopts.RemotePort), Subdomain: tunopts.Subdomain}

	tsport.SendPacket(protocol.Encode(tunnelReqFm))

	fmt.Println("send packet prepare..")
	packet, err := tsport.ReadPacket()
	if err != nil {
		fmt.Println("transport read packet err;", err.Error())
		return "", errors.New("transport read packet error!")
	}

	respFm, err := protocol.Decode(packet)
	if err != nil {
		return "", errors.New("protocol error!11")
	}

	fmt.Println(respFm)
	if respFm.Status != 1 {
		return "", errors.New("tunnel prepare failed!" + respFm.Message)
	}

	return respFm.Message, nil
}

func (cli *Client) handleStream(worker *tunnel.TunnelStub, tunopts *TunnelConfig, stream *tunnel.GwkStream) {
	defer stream.Close()

	targetAddr := fmt.Sprintf("%s:%d", "127.0.0.1", tunopts.LocalPort)
	cli.logger.Infof("REQ CONNECT=>%s\n", targetAddr)
	tsocket, err := net.Dial("tcp", targetAddr)
	if err != nil {
		return
	}
	defer tsocket.Close()
	cli.logger.Infof("DIAL SUCCESS==>%s\n", targetAddr, stream.Cid)
	worker.SetReady(stream)
	err = tunnel.Relay(tsocket, stream)
	if err != nil {
		cli.logger.Errorf("stream err:%s\n", err.Error())
	}
}

func (cli *Client) setupTunnel(name string) {

	fmt.Println("setupTunnel name:", name)
	tunopts := cli.opts.Tunnels[name]
	// 1. auth
	// 2. prepare
	// 3. setup stub
	// 4. listen stream
	tunnelHost := cli.opts.TunnelHost
	tunnelPort := cli.opts.TunnelAddr
	tsport, err := transport.NewTcpTransport(tunnelHost, strconv.Itoa(tunnelPort))
	if err != nil {
		log.Fatal(err)
		return
	}
	cli.logger.Infof("auth tunnel:%s\n", tunopts.Name)
	err = auth.HandleAuthReq("xxxxx", tsport)
	if err != nil {
		fmt.Println("auth tunnel err:", err.Error())
		return
	}

	cli.logger.Infof("auth tunnel:%s ok\n", tunopts.Name)

	cli.logger.Infof("prepare tunnel:%s\n", tunopts.Name)
	msg, err := cli.handlePrepare(&tunopts, tsport)
	if err != nil {
		fmt.Println("prepare tunnel err:", err.Error())
		return
	}
	cli.logger.Infof("prepare tunnel:%s ok, %s =>tcp:%s%d\n", tunopts.Name, msg, "127.0.0.1:", tunopts.LocalPort)
	tunnelworker := tunnel.NewTunnelStub(tsport)
	for {
		fmt.Println("tunnelworker.Accept....", tunopts)
		stream, err := tunnelworker.Accept()

		fmt.Println("stream commine")
		if err != nil {
			// transport error
			cli.logger.Errorf("stream accept err:%s\n", err.Error())
			return
		}
		go cli.handleStream(tunnelworker, &tunopts, stream)
	}
}

func (cli *Client) Bootstrap() {
	var wg sync.WaitGroup
	for key, _ := range cli.opts.Tunnels {
		wg.Add(1)
		// call setupTunnel
		go func(name string) {
			defer wg.Done()
			cli.setupTunnel(name)
		}(key)
	}

	wg.Wait()
	println("all goroutine finished!")
}

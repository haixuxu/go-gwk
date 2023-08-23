package gwk

import (
	"fmt"
	"github.com/bbk47/toolbox"
	"github/xuxihai123/go-gwk/v1/src/transport"
	"github/xuxihai123/go-gwk/v1/src/tunnel"
	. "github/xuxihai123/go-gwk/v1/src/types"
	"github/xuxihai123/go-gwk/v1/src/utils"
	"net"
	"strconv"
	"sync"
	"time"
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

func (cli *Client) handleStream(worker *tunnel.TunnelStub, tunopts *TunnelOpts, stream *tunnel.GwkStream) {
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
	} else {
		cli.logger.Infof("stream close====>", stream.Cid)
	}
}

func (cli *Client) setupTunnel(name string) {
	defer func() {
		//fmt.Println("last close====>")
		time.Sleep(3 * time.Second)
		cli.setupTunnel(name)
	}()
	// 1. auth
	// 2. prepare
	// 3. setup stub
	// 4. listen stream
	fmt.Println("setupTunnel name:", name)
	tunopts := cli.opts.Tunnels[name]
	tunnelHost := cli.opts.TunnelHost
	tunnelPort := cli.opts.TunnelAddr
	tsport, err := transport.NewTcpTransport(tunnelHost, strconv.Itoa(tunnelPort))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tsport.Close()

	tunnelworker := tunnel.NewTunnelStub(tsport)
	tunnelworker.DoWork()
	_, err = tunnelworker.StartAuth("test:test123")
	if err != nil {
		cli.logger.Errorf("auth err:%s\n", err.Error())
		return
	}
	message, err := tunnelworker.PrepareTunnel(&tunopts)
	if err != nil {
		cli.logger.Errorf("err:%s\n", err.Error())
		return
	}
	cli.logger.Infof("%10s, tunnel ok, %s =>tcp:%s%d\n", tunopts.Name, message, "127.0.0.1:", tunopts.LocalPort)
	for {
		stream, err := tunnelworker.Accept()
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

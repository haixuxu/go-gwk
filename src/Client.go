package gwk

import (
	"fmt"
	"github.com/bbk47/toolbox"
	"github/xuxihai123/go-gwk/v1/src/console"
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
	pr     *console.Printer
	opts   *ClientOpts
	logger *toolbox.Logger
	// inner attr
	tunnelStatus uint8
}

type ConsoleMsg struct {
	name       string
	statusText string
}

func NewClient(opts *ClientOpts) Client {
	cli := Client{}

	cli.opts = opts
	cli.pr = console.NewPrinter()
	cli.logger = utils.NewLogger("C", opts.LogLevel)

	return cli
}

func (cli *Client) handleStream(worker *tunnel.TunnelStub, tunopts *TunnelOpts, stream *tunnel.GwkStream, sucessMsg string) {
	defer stream.Close()

	targetAddr := fmt.Sprintf("%s:%d", "127.0.0.1", tunopts.LocalPort)
	//cli.logger.Infof("REQ CONNECT=>%s\n", targetAddr)
	cli.updateConsole(tunopts, fmt.Sprintf("%s \033[32m->\033[0m", sucessMsg))
	tsocket, err := net.Dial("tcp", targetAddr)
	if err != nil {
		return
	}
	defer tsocket.Close()
	cli.updateConsole(tunopts, fmt.Sprintf("%s \033[32m<->\033[0m", sucessMsg))
	//cli.logger.Infof("DIAL SUCCESS==>%s\n", targetAddr, stream.Cid)
	worker.SetReady(stream)
	err = tunnel.Relay(tsocket, stream)
	if err != nil {
		cli.updateConsole(tunopts, fmt.Sprintf("%s stream err:\033[31m%s\033[0m", sucessMsg, err.Error()))
		//cli.logger.Errorf("stream err:%s\n", err.Error())
	} else {
		cli.updateConsole(tunopts, sucessMsg)
		//cli.logger.Infof("stream close====>", stream.Cid)
	}
}

func (cli *Client) updateConsole(tunopts *TunnelOpts, statusText string) {
	tunopts.Status = statusText
	msg := "tunnel list:\n"
	for _, value := range cli.opts.Tunnels {
		msg = fmt.Sprintf("%s%s\n", msg, value.Status)
	}
	cli.pr.Flush(msg)
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
	tunopts := cli.opts.Tunnels[name]
	tunnelHost := cli.opts.TunnelHost
	tunnelPort := cli.opts.TunnelAddr
	cli.updateConsole(tunopts, "connecting tunnel:"+name)
	tsport, err := transport.NewTcpTransport(tunnelHost, strconv.Itoa(tunnelPort))
	if err != nil {
		//fmt.Println(err)
		cli.updateConsole(tunopts, "create conn err:"+err.Error())
		return
	}
	defer tsport.Close()

	tunnelworker := tunnel.NewTunnelStub(tsport)
	tunnelworker.DoWork()
	_, err = tunnelworker.StartAuth("test:test123")
	if err != nil {
		cli.updateConsole(tunopts, "auth err:"+err.Error())
		//cli.logger.Errorf("auth err:%s\n", err.Error())
		return
	}
	message, err := tunnelworker.PrepareTunnel(tunopts)
	if err != nil {
		cli.updateConsole(tunopts, "prepare err:"+err.Error())
		//cli.logger.Errorf("err:%s\n", err.Error())
		return
	}
	sucmsg := fmt.Sprintf("%-10s \033[32mok\033[0m, %s =>tcp://127.0.0.1:%d", tunopts.Name, message, tunopts.LocalPort)
	cli.updateConsole(tunopts, sucmsg)
	//cli.logger.Infof("sucmsg\n", sucmsg)
	for {
		stream, err := tunnelworker.Accept()
		if err != nil {
			// transport error
			cli.updateConsole(tunopts, "stream accept err:"+err.Error())
			//cli.logger.Errorf("stream accept err:%s\n", err.Error())
			return
		}
		go cli.handleStream(tunnelworker, tunopts, stream, sucmsg)
	}

}

func (cli *Client) Bootstrap() {
	var wg sync.WaitGroup

	go cli.pr.Start()

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

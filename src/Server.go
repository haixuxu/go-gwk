package gwk

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/bbk47/toolbox"
	"github/xuxihai123/go-gwk/v1/src/transport"
	"github/xuxihai123/go-gwk/v1/src/tunnel"
	. "github/xuxihai123/go-gwk/v1/src/types"
	"github/xuxihai123/go-gwk/v1/src/utils"
	"log"
	"net"
	"regexp"
	"sync"
)

type ConnectObj struct {
	uid     string
	tunnel  *tunnel.TunnelStub
	tunopts *TunnelOpts
	rtt     int
	url     string
	ln      net.Listener
}

type Server struct {
	opts        *ServerOpts
	logger      *toolbox.Logger
	connections map[string]*ConnectObj
	webTunnels  map[string]*ConnectObj //线程共享变量
	rlock       sync.Mutex
}

func NewServer(opt *ServerOpts) Server {
	s := Server{}
	s.opts = opt
	s.webTunnels = make(map[string]*ConnectObj)
	s.connections = make(map[string]*ConnectObj)
	s.logger = utils.NewLogger("S", opt.LogLevel)
	return s
}

func (servss *Server) handleTcpPipe(worker *tunnel.TunnelStub, listener net.Listener) {
	defer listener.Close()
	// 处理新连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go func() {
			defer conn.Close()
			newstream := worker.CreateStream()
			fmt.Println("create stream ok...")
			err = tunnel.Relay(conn, newstream)
			if err != nil {
				servss.logger.Errorf("stream err:%s\n", err.Error())
			}
		}()
	}
}

func (servss *Server) handleTcpTunnel(connobj *ConnectObj, tunopts *TunnelOpts) *tunnel.StatusMsg {
	//servss.logger.Infof("handle tcp tunnel===>", tunopts)
	remoteAddr := fmt.Sprintf("%s:%d", "127.0.0.1", tunopts.RemotePort)
	listener, err := net.Listen("tcp", remoteAddr)
	if err != nil {
		return &tunnel.StatusMsg{Status: tunnel.FAIELD, Message: err.Error()}
	}

	// 获取监听的地址和端口号
	addr := listener.Addr().(*net.TCPAddr)
	go servss.handleTcpPipe(connobj.tunnel, listener)

	connobj.ln = listener

	msg := fmt.Sprintf("tcp://%s:%d", servss.opts.ServerHost, addr.Port)
	return &tunnel.StatusMsg{Status: tunnel.OK, Message: msg}
}

func (servss *Server) handleWebTunnel(connobj *ConnectObj, tunopts *TunnelOpts) *tunnel.StatusMsg {
	//servss.logger.Infof("handle web tunnel===>", tunopts)
	fulldomain := fmt.Sprintf("%s.%s", tunopts.Subdomain, servss.opts.ServerHost)

	if servss.webTunnels[fulldomain] != nil {
		return &tunnel.StatusMsg{Status: tunnel.FAIELD, Message: "subdomain existed!"}
	}
	connobj.url = "http://" + fulldomain
	servss.webTunnels[fulldomain] = connobj
	return &tunnel.StatusMsg{Status: tunnel.OK, Message: connobj.url}
}

func (servss *Server) handleConnection(conn net.Conn) {
	tsport := transport.WrapConn(conn)

	tunnelworker := tunnel.NewTunnelStub(tsport)
	connobj := &ConnectObj{tunnel: tunnelworker, rtt: 0, uid: utils.GetUUID()}

	tunnelworker.RegisterAuth(func(authstr string) *tunnel.StatusMsg {
		fmt.Println("hand auth===>", authstr)
		if authstr == "test:test123" {
			return &tunnel.StatusMsg{Status: tunnel.OK, Message: "success"}
		} else {
			return &tunnel.StatusMsg{Status: tunnel.FAIELD, Message: "user/pass error!"}
		}
	})

	tunnelworker.RegisterReqTun(func(tunops *TunnelOpts) *tunnel.StatusMsg {
		tunopsstr, _ := json.Marshal(tunops)
		fmt.Println("tunopts:", string(tunopsstr))
		connobj.tunopts = tunops
		if tunops.Type == "tcp" {
			return servss.handleTcpTunnel(connobj, tunops)
		} else {
			return servss.handleWebTunnel(connobj, tunops)
		}
	})

	tunnelworker.DoWork()
	tunnelworker.AwaitClose()
	//fmt.Println("clear =====>")
	// clear
	_ = conn.Close()
	servss.rlock.Lock()
	defer servss.rlock.Unlock()

	delete(servss.connections, connobj.uid)
	if connobj.tunopts == nil {
		return
	}
	tunopts := connobj.tunopts
	if tunopts.Type == "web" {
		fulldomain := connobj.url[7:]
		fmt.Printf("remove web fulldomain:%s\n", fulldomain)
		delete(servss.webTunnels, fulldomain)
	}

	if connobj.ln != nil {
		fmt.Printf("stop server on 127.0.0.1:%d\n", tunopts.RemotePort)
		_ = connobj.ln.Close()
	}
}

func (servss *Server) handleHttpRequest(conn net.Conn) {
	defer conn.Close()

	req, err := utils.ParseHttpHeader(conn)
	if err != nil {
		conn.Write([]byte("HTTP/1.1 200 OK\n\n invalid http request\r\n"))
		return
	}

	// 获取请求头部的 Host
	host := req.Headers["Host"]
	re := regexp.MustCompile(`:\d+`)
	targetDomain := re.ReplaceAllString(host, "")

	connobj := servss.webTunnels[targetDomain]
	if connobj == nil {
		conn.Write([]byte("HTTP/1.1 200 OK\n\n server host missing!\r\n"))
		return
	}
	newstream := connobj.tunnel.CreateStream()
	fmt.Println("create stream ok..for host.", host)
	newstream.Write(req.RawBuffer)
	newstream.Write([]byte("\r\n"))
	err = tunnel.Relay(conn, newstream)
	if err != nil {
		servss.logger.Errorf("stream err:%s\n", err.Error())
	} else {
		servss.logger.Infof("stream close:%s\n", host)
	}
}

func (servss *Server) listenSocket(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go servss.handleHttpRequest(conn)
	}
}

func (servss *Server) initTcpServer(wg *sync.WaitGroup) {
	defer wg.Done()
	opts := servss.opts

	address := fmt.Sprintf("%s:%d", "127.0.0.1", opts.TunnelAddr)
	tcpserver, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}

	servss.logger.Infof("server listen on tcp://127.0.0.1:%d\n", opts.TunnelAddr)

	for {
		conn, err := tcpserver.Accept()
		if err != nil {
			continue
		}
		go servss.handleConnection(conn)
	}
}

func (servss *Server) initHttpsServer(wg *sync.WaitGroup) {
	defer wg.Done()
	listenPort := servss.opts.HttpsAddr
	address := fmt.Sprintf("%s:%d", "127.0.0.1", listenPort)
	cer, err := tls.LoadX509KeyPair(servss.opts.TlsCrt, servss.opts.TlsKey)
	if err != nil {
		log.Fatal(err)
	}
	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	ln, err := tls.Listen("tcp", address, config)
	if err != nil {
		log.Fatal(err)
	}
	servss.logger.Infof("https server listen on %s\n", address)
	servss.listenSocket(ln)
}

func (servss *Server) initHttpServer(wg *sync.WaitGroup) {
	defer wg.Done()
	listenPort := servss.opts.HttpAddr
	address := fmt.Sprintf("%s:%d", "127.0.0.1", listenPort)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	servss.logger.Infof("http server listen on %s\n", address)
	servss.listenSocket(ln)
}

func (servss *Server) Bootstrap() {
	opts := servss.opts
	var wg sync.WaitGroup

	wg.Add(1)
	go servss.initTcpServer(&wg)

	if opts.HttpAddr != 0 {
		wg.Add(1)
		go servss.initHttpServer(&wg)
	}

	if opts.HttpsAddr != 0 {
		wg.Add(1)
		go servss.initHttpsServer(&wg)
	}

	wg.Wait()
	println("all goroutine finished!")
}

package gwk

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/bbk47/toolbox"
	"github/xuxihai123/go-gwk/v1/src/auth"
	"github/xuxihai123/go-gwk/v1/src/prepare"
	"github/xuxihai123/go-gwk/v1/src/stub"
	"github/xuxihai123/go-gwk/v1/src/transport"
	. "github/xuxihai123/go-gwk/v1/src/types"
	"github/xuxihai123/go-gwk/v1/src/utils"
	"log"
	"net"
	"regexp"
	"sync"
	"time"
)

type ConnectObj struct {
	uid     string
	tunnel  *stub.TunnelStub
	tunopts *TunnelOpts
	rtt     int64
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

func (servss *Server) handleTcpPipe(worker *stub.TunnelStub, listener net.Listener, addr *net.TCPAddr) {
	defer listener.Close()
	// 处理新连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go func() {
			defer conn.Close()

			newstream, err := worker.CreateStream()
			if err != nil {
				servss.logger.Infof("create stream err:%s\n", err.Error())
				return
			}
			servss.logger.Infof("create stream ok:%s\n", addr.String())
			err = stub.Relay(conn, newstream)
			if err != nil {
				servss.logger.Errorf("stream err:%s\n", err.Error())
			} else {
				servss.logger.Infof("stream close:%s\n", addr.String())

			}
		}()
	}
}

func (servss *Server) handleTcpTunnel(connobj *ConnectObj, tunopts *TunnelOpts) *StatusMsg {
	//servss.logger.Infof("handle tcp stub===>", tunopts)
	remoteAddr := fmt.Sprintf("%s:%d", "127.0.0.1", tunopts.RemotePort)
	listener, err := net.Listen("tcp", remoteAddr)
	if err != nil {
		return &StatusMsg{Status: stub.FAIELD, Message: err.Error()}
	}

	// 获取监听的地址和端口号
	addr := listener.Addr().(*net.TCPAddr)
	go servss.handleTcpPipe(connobj.tunnel, listener, addr)

	connobj.ln = listener

	msg := fmt.Sprintf("tcp://%s:%d", servss.opts.ServerHost, addr.Port)
	return &StatusMsg{Status: stub.OK, Message: msg}
}

func (servss *Server) handleWebTunnel(connobj *ConnectObj, tunopts *TunnelOpts) *StatusMsg {
	//servss.logger.Infof("handle web stub===>", tunopts)
	fulldomain := fmt.Sprintf("%s.%s", tunopts.Subdomain, servss.opts.ServerHost)

	if servss.webTunnels[fulldomain] != nil {
		return &StatusMsg{Status: stub.FAIELD, Message: "subdomain existed!"}
	}
	connobj.url = "http://" + fulldomain
	servss.webTunnels[fulldomain] = connobj
	return &StatusMsg{Status: stub.OK, Message: connobj.url}
}

func (servss *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	tsport := transport.WrapConn(conn)
	err := auth.HandleAuthRes(tsport, func(authstr string) *StatusMsg {
		servss.logger.Infof("hand auth===>%s\n", authstr)
		if authstr == "test:test123" {
			return &StatusMsg{Status: stub.OK, Message: "success"}
		} else {
			return &StatusMsg{Status: stub.FAIELD, Message: "user/pass error!"}
		}
	})

	if err != nil {
		return
	}
	connobj := &ConnectObj{rtt: 0, uid: utils.GetUUID()}

	err = prepare.HandleTunnelRes(tsport, func(tunopts *TunnelOpts) *StatusMsg {
		tunopsstr, _ := json.Marshal(tunopts)
		servss.logger.Infof("tunopts:%s\n", string(tunopsstr))
		connobj.tunopts = tunopts
		if tunopts.Type == "tcp" {
			return servss.handleTcpTunnel(connobj, tunopts)
		} else {
			return servss.handleWebTunnel(connobj, tunopts)
		}
	})

	if err != nil {
		return
	}

	tunnelworker := stub.NewTunnelStub(tsport)
	connobj.tunnel = tunnelworker
	servss.connections[connobj.uid] = connobj
	tunnelworker.NotifyPong(func(up, down int64) {
		connobj.rtt = up + down
		//servss.logger.Infof("stub %s: up:%d,down:%d", connobj.tunopts.Name, up, down)
	})
	tunnelworker.AwaitClose()
	//fmt.Println("clear =====>")
	// clear
	servss.rlock.Lock()
	defer servss.rlock.Unlock()
	delete(servss.connections, connobj.uid)

	if connobj.tunopts == nil {
		return
	}
	tunopts := connobj.tunopts
	if tunopts.Type == "web" {
		fulldomain := connobj.url[7:]
		servss.logger.Infof("remove web fulldomain:%s\n", fulldomain)
		delete(servss.webTunnels, fulldomain)
	}

	if connobj.ln != nil {
		servss.logger.Infof("stop server on 127.0.0.1:%d\n", tunopts.RemotePort)
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
	newstream, err := connobj.tunnel.CreateStream()
	if err != nil {
		conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\n\n%s\r\n", err.Error())))
		return
	}
	servss.logger.Infof("create stream ok..for %s\n", host)
	newstream.Write(req.RawBuffer)
	newstream.Write([]byte("\r\n"))
	err = stub.Relay(conn, newstream)
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
func (svc *Server) keepPingWs() {
	ticker := time.Tick(time.Second * 5)
	for range ticker {
		for _, value := range svc.connections {
			tunnelobj := value.tunnel
			if tunnelobj != nil {
				tunnelobj.Ping()
			}
		}
	}
}

func (servss *Server) Bootstrap() {
	opts := servss.opts
	var wg sync.WaitGroup

	wg.Add(2)
	go servss.initTcpServer(&wg)
	go servss.keepPingWs()

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

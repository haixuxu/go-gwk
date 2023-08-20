package gwk

import (
	"crypto/tls"
	"fmt"
	"github.com/bbk47/toolbox"
	"github/xuxihai123/go-gwk/v1/src/auth"
	"github/xuxihai123/go-gwk/v1/src/protocol"
	"github/xuxihai123/go-gwk/v1/src/transport"
	"github/xuxihai123/go-gwk/v1/src/tunnel"
	"github/xuxihai123/go-gwk/v1/src/utils"
	"io"
	"log"
	"net"
	"regexp"
	"strings"
	"sync"
)

type Server struct {
	opts       *ServerOpts
	logger     *toolbox.Logger
	webTunnels map[string]*tunnel.TunnelStub //线程共享变量
}

func NewServer(opt *ServerOpts) Server {
	s := Server{}
	s.opts = opt
	s.webTunnels = make(map[string]*tunnel.TunnelStub)
	s.logger = utils.NewLogger("S", opt.LogLevel)
	return s
}

func (servss *Server) handlePrepare(tsport *transport.TcpTransport) {
	packet, err := tsport.ReadPacket()
	//fmt.Printf("receive====:%x\n", packet) // a6010b7463703030313a37323030
	//fmt.Printf("transport read data:len:%d\n", len(packet))

	if err != nil {
		fmt.Println("transport read packet err;", err.Error())
		authResFm := &protocol.Frame{Type: protocol.TUNNEL_RES, Status: 0x2, Message: err.Error()}
		tsport.SendPacket(protocol.Encode(authResFm))
		return
	}
	reqfm, err := protocol.Decode(packet)
	if err != nil {
		authResFm := &protocol.Frame{Type: protocol.TUNNEL_RES, Status: 0x2, Message: err.Error()}
		tsport.SendPacket(protocol.Encode(authResFm))
		return
	}

	if reqfm.Protocol == 0x1 {

		remoteAddr := fmt.Sprintf("%s:%d", "127.0.0.1", reqfm.Port)
		//listener, err := net.Listen("tcp", ":0") // 监听随机可用的端口
		listener, err := net.Listen("tcp", remoteAddr)
		if err != nil {
			fmt.Println("无法监听端口:", err)
			authResFm := &protocol.Frame{Type: protocol.TUNNEL_RES, Status: 0x2, Message: err.Error()}
			tsport.SendPacket(protocol.Encode(authResFm))
			return
		}

		authResFm := &protocol.Frame{Type: protocol.TUNNEL_RES, Status: 0x1, Message: "success"}
		tsport.SendPacket(protocol.Encode(authResFm))
		defer listener.Close()

		// 获取监听的地址和端口号
		addr := listener.Addr().(*net.TCPAddr)
		fmt.Println("监听地址:", addr.IP)
		fmt.Println("监听端口:", addr.Port)

		stub := tunnel.NewTunnelStub(tsport)

		// 处理新连接
		for {
			conn, err := listener.Accept()

			fmt.Println("connection====")
			if err != nil {
				fmt.Println("接受连接时发生错误:", err)
				continue
			}
			go func() {
				newstream := stub.CreateStream()

				fmt.Println("create stream...")
				<-newstream.Ready
				fmt.Println("create stream ok...")
				err = tunnel.Relay(conn, newstream)
				if err != nil {
					servss.logger.Errorf("stream err:%s\n", err.Error())
				}
			}()
			// 在这里处理新连接...
			//_ = conn.Close()
		}
	} else {

		fulldomain := fmt.Sprintf("%s.%s", reqfm.Subdomain, servss.opts.ServerHost)

		stub := tunnel.NewTunnelStub(tsport)
		if servss.webTunnels[fulldomain] != nil {
			authResFm := &protocol.Frame{Type: protocol.TUNNEL_RES, Status: 0x2, Message: "subdomain existed!"}
			tsport.SendPacket(protocol.Encode(authResFm))
			return
		}
		servss.webTunnels[fulldomain] = stub
		authResFm := &protocol.Frame{Type: protocol.TUNNEL_RES, Status: 0x1, Message: "http://" + fulldomain}
		tsport.SendPacket(protocol.Encode(authResFm))
		return
	}

}

func (servss *Server) handleConnection(conn net.Conn) {
	fmt.Println("connection====>")
	tsport := transport.WrapConn(conn)

	auth.HandleAuthRes("bbbb", tsport)
	fmt.Println("auth ---ok")
	servss.handlePrepare(tsport)
}

func (servss *Server) initTcpServer() {
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

func (servss *Server) handleHttpRequest(conn net.Conn) {
	defer conn.Close()

	servss.logger.Infof("handle http request==")
	// 创建缓冲读取器以从连接中读取数据
	reader := io.Reader(conn)

	// 读取第一行，解析请求方法和路径
	requestLine, err := utils.ReadOneLine(reader)
	if err != nil {
		log.Println("无法读取请求:", err)
		conn.Write([]byte("HTTP/1.1 200 OK\n\n target service invalid\r\n"))
		return
	}

	// 解析请求行
	parts := strings.Split(requestLine, " ")
	if len(parts) < 3 {
		log.Println("无效的请求行:", requestLine)
		conn.Write([]byte("HTTP/1.1 200 OK\n\n target service invalid\r\n"))
		return
	}
	method := parts[0]
	path := parts[1]

	fmt.Println("request :", method, path)

	cache := []byte(fmt.Sprintf("%s %s %s\r\n", method, path, parts[2]))
	// 解析请求头部
	headers := make(map[string]string)
	for {
		line, err := utils.ReadOneLine(reader)
		if err != nil || line == "" {
			break
		}
		cache = append(cache, []byte(line+"\r\n")...)
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			headerName := strings.TrimSpace(parts[0])
			headerValue := strings.TrimSpace(parts[1])
			headers[headerName] = headerValue
		}
	}

	// 获取请求头部的 Host
	host := headers["Host"]
	re := regexp.MustCompile(`:\d+`)
	targetDomain := re.ReplaceAllString(host, "")

	tunworker := servss.webTunnels[targetDomain]
	if tunworker == nil {
		conn.Write([]byte("HTTP/1.1 200 OK\n\n target service invalid\r\n"))
		return
	}

	newstream := tunworker.CreateStream()

	fmt.Println("create stream...", host)
	<-newstream.Ready
	fmt.Println("create stream ok..for host.", host)
	fmt.Println(string(cache))
	newstream.Write(cache)
	newstream.Write([]byte("\r\n"))
	err = tunnel.Relay(conn, newstream)
	if err != nil {
		servss.logger.Errorf("stream err:%s\n", err.Error())
	} else {
		servss.logger.Infof("stream close:%s\n", host)
	}
}

func (servss *Server) initHttpsServer() {

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

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go servss.handleHttpRequest(conn)
	}
}

func (servss *Server) initHttpServer() {
	listenPort := servss.opts.HttpAddr
	address := fmt.Sprintf("%s:%d", "127.0.0.1", listenPort)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	servss.logger.Infof("http server listen on %s\n", address)
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go servss.handleHttpRequest(conn)
	}
}

func (servss *Server) Bootstrap() {
	opts := servss.opts
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		servss.initTcpServer()
	}()

	if opts.HttpAddr != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			servss.initHttpServer()
		}()
	}

	if opts.HttpsAddr != 0 {
		wg.Add(1)
		wg.Add(1)
		go func() {
			defer wg.Done()
			servss.initHttpsServer()
		}()
	}

	wg.Wait()
	println("all goroutine finished!")
}

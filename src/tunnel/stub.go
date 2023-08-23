package tunnel

import (
	"errors"
	"fmt"
	"github.com/bbk47/toolbox"
	"github/xuxihai123/go-gwk/v1/src/protocol"
	"github/xuxihai123/go-gwk/v1/src/transport"
	. "github/xuxihai123/go-gwk/v1/src/types"
	"github/xuxihai123/go-gwk/v1/src/utils"
)

const OK = 0x1
const FAIELD = 0x2

type PongFunc func(up, down int64)
type AuthFunc func(authstr string) *StatusMsg
type ReqTunFun func(opt *TunnelOpts) *StatusMsg

type TunnelStub struct {
	tsport      *transport.TcpTransport
	streams     map[string]*GwkStream
	streamch    chan *GwkStream
	sendch      chan *protocol.Frame
	closech     chan uint8
	state       string // "init"=>"authed"=>"ready"
	errmsg      string // close msg
	authedch    chan *StatusMsg
	tunnelResCh chan *StatusMsg
	seq         uint32
	//wlock    sync.Mutex
	pongFunc   PongFunc
	authFunc   AuthFunc
	reqtunFunc ReqTunFun
}

type StatusMsg struct {
	Status  uint8
	Message string
}

func NewTunnelStub(tsport *transport.TcpTransport) *TunnelStub {
	stub := TunnelStub{tsport: tsport}
	stub.streamch = make(chan *GwkStream, 1024)
	stub.sendch = make(chan *protocol.Frame, 1024)
	stub.streams = make(map[string]*GwkStream)
	stub.closech = make(chan uint8)
	stub.authedch = make(chan *StatusMsg)
	stub.tunnelResCh = make(chan *StatusMsg)
	stub.state = "init"
	return &stub
}

func (ts *TunnelStub) NotifyPong(handler func(up, down int64)) {
	ts.pongFunc = handler
}

func (ts *TunnelStub) RegisterAuth(handler AuthFunc) {
	ts.authFunc = handler
}

func (ts *TunnelStub) RegisterReqTun(handler ReqTunFun) {
	ts.reqtunFunc = handler
}

func (ts *TunnelStub) DoWork() {
	go ts.readWorker()
	go ts.writeWorker()
}

func (ts *TunnelStub) AwaitClose() {
	<-ts.closech
}

func (ts *TunnelStub) sendTinyFrame(frame *protocol.Frame) error {
	// 序列化数据
	binaryData := protocol.Encode(frame)

	//ts.wlock.Lock()
	//defer ts.wlock.Unlock()
	// 发送数据
	//log.Printf("write tunnel cid:%s, data[%d]bytes, frame type:%d\n", frame.StreamID, len(binaryData), frame.Type)
	return ts.tsport.SendPacket(binaryData)
}

func (ts *TunnelStub) sendDataFrame(streamId string, data []byte) {
	frame := &protocol.Frame{Type: protocol.STREAM_DATA, StreamID: streamId, Data: data}
	ts.sendch <- frame
}

func (ts *TunnelStub) sendFrame(frame *protocol.Frame) error {
	frames := protocol.SplitFrame(frame)
	for _, smallframe := range frames {
		err := ts.sendTinyFrame(smallframe)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ts *TunnelStub) closeStream(streamId string) {
	ts.destroyStream(streamId)

	frame := &protocol.Frame{Type: protocol.STREAM_FIN, StreamID: streamId, Data: []byte{0x1, 0x1}}
	ts.sendch <- frame
}

func (ts *TunnelStub) resetStream(streamId string) {
	ts.destroyStream(streamId)
	frame := &protocol.Frame{Type: protocol.STREAM_RST, StreamID: streamId, Data: []byte{0x1, 0x2}}
	ts.sendch <- frame
}

func (ts *TunnelStub) writeWorker() {
	//fmt.Println("writeWorker====")
	for {
		select {
		case ref := <-ts.sendch:
			ts.sendFrame(ref)
		case <-ts.closech:
			return
		}
	}
}

func (ts *TunnelStub) readWorker() {
	//fmt.Println("readworker====")
	defer func() {
		close(ts.closech)
	}()
	for {
		packet, err := ts.tsport.ReadPacket()
		//fmt.Printf("receive====:%d\n", len(packet))
		//fmt.Printf("transport read data:len:%d\n", len(packet))
		if err != nil {
			ts.errmsg = "read packet err:" + err.Error()
			return
		}
		respFrame, err := protocol.Decode(packet)
		if err != nil {
			ts.errmsg = "protocol error"
			return
		}

		//log.Printf("read  tunnel cid:%s, data[%d]bytes, frame type:%d\n", respFrame.StreamID, len(packet), respFrame.Type)
		if respFrame.Type == protocol.AUTH_REQ {
			authRet := ts.authFunc(respFrame.Token)
			authResFm := &protocol.Frame{Type: protocol.AUTH_RES, Status: authRet.Status, Token: authRet.Message}
			ts.sendch <- authResFm
		} else if respFrame.Type == protocol.AUTH_RES {
			ts.state = "authed"
			ts.authedch <- &StatusMsg{Status: respFrame.Status, Message: respFrame.Message}
		} else if respFrame.Type == protocol.TUNNEL_REQ {
			tuntypestr := GetTypeByNo(respFrame.TunType)
			tunops := &TunnelOpts{Name: respFrame.Name, RemotePort: int(respFrame.Port), Type: tuntypestr, Subdomain: respFrame.Subdomain}
			prepareRet := ts.reqtunFunc(tunops)
			tunnelResFm := &protocol.Frame{Type: protocol.TUNNEL_RES, Status: prepareRet.Status, Message: prepareRet.Message}
			ts.sendch <- tunnelResFm
		} else if respFrame.Type == protocol.TUNNEL_RES {
			ts.state = "ready"
			ts.tunnelResCh <- &StatusMsg{Status: respFrame.Status, Message: respFrame.Message}
		} else if respFrame.Type == protocol.PING_FRAME {
			timebs := toolbox.GetNowInt64Bytes()
			data := append(respFrame.Data, timebs...)
			pongFrame := &protocol.Frame{StreamID: respFrame.StreamID, Type: protocol.PONG_FRAME, Data: data}
			ts.sendch <- pongFrame
		} else if respFrame.Type == protocol.PONG_FRAME {
			//ts.pongFunc(respFrame.Atime-respFrame.Stime, downms)
		} else if respFrame.Type == protocol.STREAM_INIT {
			st := NewGwkStream(respFrame.StreamID, ts)
			ts.streams[st.Cid] = st
			ts.streamch <- st
		} else if respFrame.Type == protocol.STREAM_EST {
			streamId := respFrame.StreamID
			stream := ts.streams[streamId]
			if stream == nil {
				ts.resetStream(streamId)
				continue
			}
			stream.Ready <- 1
			ts.streamch <- stream
		} else if respFrame.Type == protocol.STREAM_DATA {
			// find stream , write stream
			streamId := respFrame.StreamID
			stream := ts.streams[streamId]
			if stream == nil {
				ts.resetStream(streamId)
				continue
			}
			err := stream.produce(respFrame.Data)
			if err != nil {
				fmt.Println("produce err:", err)
				ts.closeStream(streamId)
			}
		} else if respFrame.Type == protocol.STREAM_FIN {
			ts.destroyStream(respFrame.StreamID)
		} else if respFrame.Type == protocol.STREAM_RST {
			ts.resetStream(respFrame.StreamID)
		} else {
			if ts.state == "init" {
				ts.state = "authed"
				ts.authedch <- &StatusMsg{Status: respFrame.Status, Message: "protocol err"}
			}
			fmt.Println("eception frame type:", respFrame.Type)
		}
	}
}

func (ts *TunnelStub) CreateStream() *GwkStream {
	streamId := utils.GetUUID()
	stream := NewGwkStream(streamId, ts)
	fmt.Println("start stream===>", streamId)
	ts.streams[streamId] = stream
	frame := &protocol.Frame{Type: protocol.STREAM_INIT, StreamID: streamId}
	ts.sendch <- frame
	<-stream.Ready

	return stream
}
func (ts *TunnelStub) SetReady(stream *GwkStream) {
	frame := &protocol.Frame{Type: protocol.STREAM_EST, StreamID: stream.Cid}
	ts.sendch <- frame
}

func (ts *TunnelStub) destroyStream(streamId string) {
	stream := ts.streams[streamId]
	if stream != nil {
		stream.Close()
		delete(ts.streams, streamId)
	}
}
func (ts *TunnelStub) StartAuth(authtoken string) (message string, err error) {

	authReqFm := &protocol.Frame{Type: protocol.AUTH_REQ, Status: 0x0, Token: authtoken}
	ts.sendch <- authReqFm
	smsg := <-ts.authedch
	if smsg.Status != 0x1 {
		return "", errors.New(smsg.Message)
	}
	return smsg.Message, nil
}

func (ts *TunnelStub) PrepareTunnel(tunopts *TunnelOpts) (msg string, err error) {
	tunReqFm := &protocol.Frame{Type: protocol.TUNNEL_REQ, TunType: tunopts.GetTypeNo(), Port: uint16(tunopts.RemotePort), Subdomain: tunopts.Subdomain, Name: tunopts.Name}
	ts.sendch <- tunReqFm
	smsg := <-ts.tunnelResCh
	if smsg.Status != 0x1 {
		return "", errors.New(smsg.Message)
	}
	return smsg.Message, nil
}

func (ts *TunnelStub) Ping() {
	stime := utils.GetNowInt64String()
	frame := &protocol.Frame{Stime: stime}
	ts.sendch <- frame
}

func (ts *TunnelStub) Accept() (*GwkStream, error) {

	select {
	case st := <-ts.streamch: // 收到stream
		return st, nil
	case <-ts.closech:
		// close transport
		return nil, errors.New(ts.errmsg)
	}
}

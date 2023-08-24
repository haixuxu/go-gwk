package tunnel

import (
	"errors"
	"fmt"
	"github.com/bbk47/toolbox"
	"github/xuxihai123/go-gwk/v1/src/protocol"
	"github/xuxihai123/go-gwk/v1/src/transport"
	"github/xuxihai123/go-gwk/v1/src/utils"
)

const OK = 0x1
const FAIELD = 0x2

type TunnelStub struct {
	tsport   *transport.TcpTransport
	streams  map[string]*GwkStream
	streamch chan *GwkStream
	sendch   chan *protocol.Frame
	closech  chan uint8
	errmsg   string // close msg
	seq      uint32
	//wlock    sync.Mutex
	pongFunc func(up, down int64)
}

func NewTunnelStub(tsport *transport.TcpTransport) *TunnelStub {
	stub := TunnelStub{tsport: tsport}
	stub.streamch = make(chan *GwkStream, 1024)
	stub.sendch = make(chan *protocol.Frame, 1024)
	stub.streams = make(map[string]*GwkStream)
	stub.closech = make(chan uint8)
	go stub.readWorker()
	go stub.writeWorker()
	return &stub
}

func (ts *TunnelStub) NotifyPong(handler func(up, down int64)) {
	ts.pongFunc = handler
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
		if respFrame.Type == protocol.PING_FRAME {
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
				//fmt.Println("produce err:", err)
				ts.closeStream(streamId)
			}
		} else if respFrame.Type == protocol.STREAM_FIN {
			ts.destroyStream(respFrame.StreamID)
		} else if respFrame.Type == protocol.STREAM_RST {
			ts.resetStream(respFrame.StreamID)
		} else {
			fmt.Println("eception frame type:", respFrame.Type)
		}
	}
}

func (ts *TunnelStub) CreateStream() *GwkStream {
	streamId := utils.GetUUID()
	stream := NewGwkStream(streamId, ts)
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

func (ts *TunnelStub) Ping() {
	frame := &protocol.Frame{Stime: utils.GetNowInt64String()}
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

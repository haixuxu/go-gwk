package stub

import (
	"io"
	"sync"
)

type GwkStream struct {
	Cid   string
	Ready chan uint8
	ts    *TunnelStub
	rp    *io.PipeReader
	wp    *io.PipeWriter
}

func NewGwkStream(cid string, ts *TunnelStub) *GwkStream {
	s := &GwkStream{}
	s.Cid = cid
	s.ts = ts
	rp, wp := io.Pipe()
	s.rp = rp
	s.wp = wp
	s.Ready = make(chan uint8)

	return s
}

func (s *GwkStream) produce(data []byte) error {
	//fmt.Printf("produce wp====:%x\n", data)
	_, err := s.wp.Write(data)
	return err
}

func (s *GwkStream) Read(data []byte) (n int, err error) {
	n, err = s.rp.Read(data)
	//fmt.Printf("target read====:%x  len:%d\n", data[:n], n)
	return n, err
}

func (s *GwkStream) Write(p []byte) (n int, err error) {
	//fmt.Printf("write stream[%s] data:%x\n", s.Cid, p)
	buf2 := make([]byte, len(p))
	// go中使用io.Copy时，底层使用slice作为buffer cache,传入的p一直是同一个切片, 实现的目标 Writer 不能及时消费写入的数据，会导致数据覆盖
	copy(buf2, p) // io.Copy buf must copy data
	s.ts.sendDataFrame(s.Cid, buf2)
	return len(p), nil
}

func (s *GwkStream) Close() error {
	//log.Println("closeing ch")
	s.rp.Close()
	s.wp.Close()
	return nil
}

func Relay(left, right io.ReadWriteCloser) error {
	var err, err1 error
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err1 = io.Copy(right, left)
	}()
	_, err = io.Copy(left, right)
	wg.Wait()

	if err != nil {
		return err
	}

	if err1 != nil {
		return err1
	}
	return nil
}

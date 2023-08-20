package protocol


const DATA_MAX_SIZE = 1024 * 2

func SplitFrame(frame *Frame) []*Frame {
	var frames []*Frame
	leng := 0
	if frame.Data != nil {
		leng = len(frame.Data)
	}

	if frame.Type <0x10{
		frames = append(frames, frame)
		return frames
	}


	if leng <= DATA_MAX_SIZE {
		frames = append(frames, frame)
		return frames
	}
	offset := 0
	ldata := frame.Data
	for {
		offset2 := offset + DATA_MAX_SIZE
		if offset2 > leng {
			offset2 = leng
		}
		buf2 := make([]byte, offset2-offset)
		// 多个切片共享底层数组：当多个切片共享同一个底层数组时，修改其中一个切片的值可能会影响其他切片的值。
		copy(buf2, ldata[offset:offset2])
		frame2 := &Frame{Type:frame.Type,StreamID:frame.StreamID,Data:buf2}
		frames = append(frames, frame2)
		offset = offset2
		if offset2 == leng {
			break
		}
	}
	return frames
}
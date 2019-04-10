package phpserialize

import (
	"bytes"
	"sync"
)

const scratch_buffer_size = 128

type scratchBuffer [scratch_buffer_size]byte

type scrachBuffersPool struct {
	sync.Pool
}

func newScratchBuffersPool() *scrachBuffersPool {
	return &scrachBuffersPool{
		sync.Pool{
			New: func() interface{} {
				return &scratchBuffer{}
			},
		},
	}
}

func (s *scrachBuffersPool) Get() *scratchBuffer {
	return s.Pool.Get().(*scratchBuffer)
}

func (s *scrachBuffersPool) Put(sb *scratchBuffer) {
	s.Pool.Put(sb)
}

type bigScratchBufferPool struct {
	sync.Pool
}

func newBigScratchBufferPool() *bigScratchBufferPool {
	return &bigScratchBufferPool{
		sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
}

func (s *bigScratchBufferPool) Get() *bytes.Buffer {
	return s.Pool.Get().(*bytes.Buffer)
}

func (s *bigScratchBufferPool) Put(sb *bytes.Buffer) {
	sb.Reset()
	s.Pool.Put(sb)
}

var (
	sbPool  = newScratchBuffersPool()
	bsbPool = newBigScratchBufferPool()
)

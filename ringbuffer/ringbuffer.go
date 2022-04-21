package ringbuffer

import (
	"errors"
)

type RingBuffer struct {
	buf     []byte
	size    int
	r       int
	w       int
	isEmpty bool
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		buf:     make([]byte, size),
		size:    size,
		r:       0,
		w:       0,
		isEmpty: true,
	}
}

// Reset everything to zero
func (ringBuffer *RingBuffer) Reset() {
	ringBuffer.isEmpty = true
	ringBuffer.w = 0
	ringBuffer.r = 0
}

func (ringBuffer RingBuffer) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, errors.New("slice p is empty")
	}
	if ringBuffer.isEmpty {
		return 0, errors.New("ring buffer is empty now")
	} else if ringBuffer.w > ringBuffer.r {
		n = ringBuffer.w - ringBuffer.r
		if len(p) < n {
			n = len(p)
		}
		copy(p, ringBuffer.buf[ringBuffer.r:ringBuffer.r+n])
		ringBuffer.r = (ringBuffer.r + n) % ringBuffer.size
	} else {
		n = ringBuffer.size - ringBuffer.r + ringBuffer.w
		if len(p) < n {
			n = len(p)
		}
		c1 := ringBuffer.size - ringBuffer.r
		if c1 >= n {
			copy(p, ringBuffer.buf[ringBuffer.r:ringBuffer.r+n])
			ringBuffer.r = (ringBuffer.r + n) % ringBuffer.size
		} else {
			c2 := n - c1
			copy(p, ringBuffer.buf[ringBuffer.r:])
			copy(p, ringBuffer.buf[:c2])
			ringBuffer.r = (ringBuffer.r + n) % ringBuffer.size
		}
	}

	if ringBuffer.r == ringBuffer.w {
		ringBuffer.isEmpty = true
	}

	return
}

func (ringBuffer *RingBuffer) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, errors.New("slice p is empty")
	}
	for ringBuffer.Available() < len(p) {
		ringBuffer.Expand()
	}
	n = len(p)
	if ringBuffer.w < ringBuffer.r {
		copy(ringBuffer.buf[ringBuffer.w:ringBuffer.w+n], p)
	} else { // ringBuffer.w >= ringBuffer.r
		if n <= ringBuffer.size-ringBuffer.w {
			copy(ringBuffer.buf[ringBuffer.w:ringBuffer.w+n], p)
		} else {
			m := ringBuffer.size - ringBuffer.w
			copy(ringBuffer.buf[ringBuffer.w:], p[:m])
			copy(ringBuffer.buf[:n-m], p[m:])

		}
	}
	ringBuffer.w = (ringBuffer.w + n) % ringBuffer.size
	ringBuffer.isEmpty = false
	return
}

// Length available byte length
func (ringBuffer *RingBuffer) Length() int {
	if ringBuffer.isEmpty {
		return 0
	} else if ringBuffer.w > ringBuffer.r {
		return ringBuffer.w - ringBuffer.r
	} else {
		return ringBuffer.size - ringBuffer.r + ringBuffer.w
	}
}

// Available remain available byte length
func (ringBuffer *RingBuffer) Available() int {
	if ringBuffer.isEmpty {
		return ringBuffer.size
	} else if ringBuffer.w > ringBuffer.r {
		return ringBuffer.size - ringBuffer.w + ringBuffer.r
	} else {
		return ringBuffer.r - ringBuffer.w
	}
}

func (ringBuffer *RingBuffer) IsFull() bool {
	//return ringBuffer.Available() == ringBuffer.size
	if ringBuffer.isEmpty {
		return false
	} else {
		return ringBuffer.r == ringBuffer.w
	}
}

func (ringBuffer *RingBuffer) Discard(n int) {
	if n <= 0 {
		return
	} else if n > ringBuffer.Length() {
		n = ringBuffer.Length()
	}
	ringBuffer.r = (ringBuffer.r + n) % ringBuffer.size
	if ringBuffer.r == ringBuffer.w {
		ringBuffer.isEmpty = true
	}
}

// PeekAll without advancing ringBuffer.r
func (ringBuffer *RingBuffer) PeekAll() (head []byte, tail []byte) {
	if ringBuffer.isEmpty {
		return
	}
	if ringBuffer.w > ringBuffer.r {
		head = ringBuffer.buf[ringBuffer.r:ringBuffer.w]
	} else {
		head = ringBuffer.buf[ringBuffer.r:]
		tail = ringBuffer.buf[:ringBuffer.w]
	}
	return
}

// Peek without advancing
func (ringBuffer *RingBuffer) Peek(n int) (head []byte, tail []byte) {
	if ringBuffer.isEmpty || n == 0 {
		return
	}
	if ringBuffer.w > ringBuffer.r {
		if n > ringBuffer.w-ringBuffer.r {
			n = ringBuffer.w - ringBuffer.r
		}
		head = ringBuffer.buf[ringBuffer.r : ringBuffer.r+n]
	} else {
		if n > ringBuffer.size-ringBuffer.r+ringBuffer.w {
			n = ringBuffer.size - ringBuffer.r + ringBuffer.w
		}
		if ringBuffer.r+n <= ringBuffer.size {
			head = ringBuffer.buf[ringBuffer.r : ringBuffer.r+n]
		} else {
			head = ringBuffer.buf[ringBuffer.r:]
			tail = ringBuffer.buf[:n-ringBuffer.size+ringBuffer.r]
		}
	}
	return
}

// Expand capacity of underlying slice
func (ringBuffer *RingBuffer) Expand() {
	if ringBuffer.Length()*2 < ringBuffer.size {
		return
	}
	buf := make([]byte, ringBuffer.size*2)
	if ringBuffer.w > ringBuffer.r {
		copy(buf, ringBuffer.buf[ringBuffer.r:ringBuffer.w])
	} else {
		copy(buf, ringBuffer.buf[ringBuffer.r:])
		copy(buf, ringBuffer.buf[:ringBuffer.w])
	}
	ringBuffer.r = 0
	ringBuffer.w = ringBuffer.size
	ringBuffer.buf = buf
	ringBuffer.size = cap(buf)
}

package dt

type RingBuffer[T any] struct {
	buf  []T
	r, w int
}

func NewRingBuffer[T any](capacity int) RingBuffer[T] {
	return RingBuffer[T]{
		buf: make([]T, capacity),
		r:   0,
		w:   0,
	}
}
func (rb *RingBuffer[T]) Cap() int {
	return cap(rb.buf)
}
func (rb *RingBuffer[T]) Len() int {
	return rb.w - rb.r
}

// return value indicates that the buffer overwrote unread data
func (rb *RingBuffer[T]) Put(item T) bool {
	rb.w = (rb.w + 1) % len(rb.buf)
	rb.buf[rb.w] = item
	full := rb.w == rb.r
	if full {
		rb.r = (rb.r + 1) % len(rb.buf)
	}
	return full
}
// return value indicates wether the buffer is empty
func (rb *RingBuffer[T]) Pop() (T, bool) {
	var item T
	if rb.r == rb.w {
		return item, false
	} else {
		item = rb.buf[rb.r]
		rb.r = (rb.r + 1) % len(rb.buf)
		return item, true
	}
}
// return value indicates wether the buffer is empty
func (rb *RingBuffer[T]) Peek() (T, bool) {
	var item T
	if rb.r == rb.w {
		return item, false
	} else {
		return rb.buf[rb.r], true
	}
}

package emiya

import "bytes"

// getBuffer 从 pool 中获取 buffer
func (r *registry) getBuffer() *bytes.Buffer {
	return r.bufferPool.Get().(*bytes.Buffer)
}

// putBuffer 将 buffer 放回 pool
func (r *registry) putBuffer(buf *bytes.Buffer) {
	buf.Reset()
	r.bufferPool.Put(buf)
}

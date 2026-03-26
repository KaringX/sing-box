package buf

import (
	"io"
	"sync/atomic"

	"github.com/sagernet/sing-box/common/xray/bytespool"
	"github.com/sagernet/sing-box/common/xray/net"
	E "github.com/sagernet/sing/common/exceptions"
)

const (
	// Size of a regular buffer.
	Size = 8192
)

var ErrBufferFull = E.New("buffer is full")

var zero = [Size * 10]byte{0}

var pool = bytespool.GetPool(Size)

// ownership represents the data owner of the buffer.
type ownership uint8

const (
	managed ownership = iota
	unmanaged
	bytespools
)

// Buffer is a recyclable allocation of a byte array. Buffer.Release() recycles
// the buffer into an internal buffer pool, in order to recreate a buffer more
// quickly.
type Buffer struct {
	v         []byte
	start     atomic.Int32 //karing
	end       atomic.Int32 //karing
	ownership ownership
	UDP       *net.Destination
}

// New creates a Buffer with 0 length and 8K capacity, managed.
func New() *Buffer {
	buf := pool.Get().([]byte)
	if cap(buf) >= Size {
		buf = buf[:Size]
	} else {
		buf = make([]byte, Size)
	}

	return &Buffer{
		v: buf,
	}
}

// NewExisted creates a standard size Buffer with an existed bytearray, managed.
func NewExisted(b []byte) *Buffer {
	if cap(b) < Size {
		panic("Invalid buffer")
	}

	oLen := len(b)
	if oLen < Size {
		b = b[:Size]
	}

	buf := &Buffer{
		v: b,
	}
	buf.end.Store(int32(oLen))
	return buf
}

// FromBytes creates a Buffer with an existed bytearray, unmanaged.
func FromBytes(b []byte) *Buffer {
	buf := &Buffer{
		v:         b,
		ownership: unmanaged,
	}
	buf.end.Store(int32(len(b)))
	return buf
}

// StackNew creates a new Buffer object on stack, managed.
// This method is for buffers that is released in the same function.
func StackNew() Buffer {
	buf := pool.Get().([]byte)
	if cap(buf) >= Size {
		buf = buf[:Size]
	} else {
		buf = make([]byte, Size)
	}

	return Buffer{
		v: buf,
	}
}

// NewWithSize creates a Buffer with 0 length and capacity with at least the given size, bytespool's.
func NewWithSize(size int32) *Buffer {
	return &Buffer{
		v:         bytespool.Alloc(size),
		ownership: bytespools,
	}
}

// Release recycles the buffer into an internal buffer pool.
func (b *Buffer) Release() {
	if b == nil || b.v == nil || b.ownership == unmanaged {
		return
	}

	p := b.v
	b.v = nil
	b.Clear()

	switch b.ownership {
	case managed:
		if cap(p) == Size {
			pool.Put(p)
		}
	case bytespools:
		bytespool.Free(p)
	}
	b.UDP = nil
}

// Clear clears the content of the buffer, results an empty buffer with
// Len() = 0.
func (b *Buffer) Clear() {
	b.start.Store(0)
	b.end.Store(0)
}

// Byte returns the bytes at index.
func (b *Buffer) Byte(index int32) byte {
	return b.v[b.start.Load()+index]
}

// SetByte sets the byte value at index.
func (b *Buffer) SetByte(index int32, value byte) {
	b.v[b.start.Load()+index] = value
}

// Bytes returns the content bytes of this Buffer.
func (b *Buffer) Bytes() []byte {
	return b.v[b.start.Load():b.end.Load()]
}

// Extend increases the buffer size by n bytes, and returns the extended part.
// It panics if result size is larger than buf.Size.
func (b *Buffer) Extend(n int32) []byte {
	currentEnd := b.end.Load()
	newEnd := currentEnd + n
	if newEnd > int32(len(b.v)) {
		panic("extending out of bound")
	}
	ext := b.v[currentEnd:newEnd]
	b.end.Store(newEnd)
	copy(ext, zero[:])
	return ext
}

// BytesRange returns a slice of this buffer with given from and to boundary.
func (b *Buffer) BytesRange(from, to int32) []byte {
	if from < 0 {
		from += b.Len()
	}
	if to < 0 {
		to += b.Len()
	}
	start := b.start.Load()
	return b.v[start+from : start+to]
}

// BytesFrom returns a slice of this Buffer starting from the given position.
func (b *Buffer) BytesFrom(from int32) []byte {
	if from < 0 {
		from += b.Len()
	}
	return b.v[b.start.Load()+from : b.end.Load()]
}

// BytesTo returns a slice of this Buffer from start to the given position.
func (b *Buffer) BytesTo(to int32) []byte {
	if to < 0 {
		to += b.Len()
	}
	if to < 0 {
		to = 0
	}
	start := b.start.Load()
	return b.v[start : start+to]
}

// Check makes sure that 0 <= b.start <= b.end.
func (b *Buffer) Check() {
	start := b.start.Load()
	end := b.end.Load()
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = 0
	}
	if start > end {
		start = end
	}
	b.start.Store(start)
	b.end.Store(end)
}

// Resize cuts the buffer at the given position.
func (b *Buffer) Resize(from, to int32) {
	oldEnd := b.end.Load()
	start := b.start.Load()
	if from < 0 {
		from += b.Len()
	}
	if to < 0 {
		to += b.Len()
	}
	if to < from {
		panic("Invalid slice")
	}
	newEnd := start + to
	newStart := start + from
	b.end.Store(newEnd)
	b.start.Store(newStart)
	b.Check()
	if newEnd > oldEnd {
		copy(b.v[oldEnd:newEnd], zero[:])
	}
}

// Advance cuts the buffer at the given position.
func (b *Buffer) Advance(from int32) {
	if from < 0 {
		from += b.Len()
	}
	b.start.Add(from)
	b.Check()
}

// Len returns the length of the buffer content.
func (b *Buffer) Len() int32 {
	if b == nil {
		return 0
	}
	return b.end.Load() - b.start.Load()
}

// Cap returns the capacity of the buffer content.
func (b *Buffer) Cap() int32 {
	if b == nil {
		return 0
	}
	return int32(len(b.v))
}

// Available returns the available capacity of the buffer content.
func (b *Buffer) Available() int32 {
	if b == nil {
		return 0
	}
	return int32(len(b.v)) - b.end.Load()
}

// IsEmpty returns true if the buffer is empty.
func (b *Buffer) IsEmpty() bool {
	return b.Len() == 0
}

// IsFull returns true if the buffer has no more room to grow.
func (b *Buffer) IsFull() bool {
	return b != nil && b.end.Load() == int32(len(b.v))
}

// Write implements Write method in io.Writer.
func (b *Buffer) Write(data []byte) (int, error) {
	currentEnd := b.end.Load()
	nBytes := copy(b.v[currentEnd:], data)
	if nBytes < len(data) {
		return nBytes, ErrBufferFull
	}
	b.end.Store(currentEnd + int32(nBytes))
	return nBytes, nil
}

// WriteByte writes a single byte into the buffer.
func (b *Buffer) WriteByte(v byte) error {
	if b.IsFull() {
		return E.New("buffer full")
	}
	currentEnd := b.end.Load()
	b.v[currentEnd] = v
	b.end.Store(currentEnd + 1)
	return nil
}

// WriteString implements io.StringWriter.
func (b *Buffer) WriteString(s string) (int, error) {
	return b.Write([]byte(s))
}

// ReadByte implements io.ByteReader
func (b *Buffer) ReadByte() (byte, error) {
	start := b.start.Load()
	end := b.end.Load()
	if start == end {
		return 0, io.EOF
	}

	nb := b.v[start]
	b.start.Store(start + 1)
	return nb, nil
}

// ReadBytes implements bufio.Reader.ReadBytes
func (b *Buffer) ReadBytes(length int32) ([]byte, error) {
	start := b.start.Load()
	end := b.end.Load()
	if end-start < length {
		return nil, io.EOF
	}

	nb := b.v[start : start+length]
	b.start.Store(start + length)
	return nb, nil
}

// Read implements io.Reader.Read().
func (b *Buffer) Read(data []byte) (int, error) {
	start := b.start.Load()
	end := b.end.Load()
	length := end - start
	if length == 0 {
		return 0, io.EOF
	}
	nBytes := copy(data, b.v[start:end])
	if int32(nBytes) == length {
		b.Clear()
	} else {
		b.start.Store(start + int32(nBytes))
	}
	return nBytes, nil
}

// ReadFrom implements io.ReaderFrom.
func (b *Buffer) ReadFrom(reader io.Reader) (int64, error) {
	currentEnd := b.end.Load()
	n, err := reader.Read(b.v[currentEnd:])
	b.end.Store(currentEnd + int32(n))
	return int64(n), err
}

// ReadFullFrom reads exact size of bytes from given reader, or until error occurs.
func (b *Buffer) ReadFullFrom(reader io.Reader, size int32) (int64, error) {
	currentEnd := b.end.Load()
	newEnd := currentEnd + size
	if newEnd > int32(len(b.v)) {
		return 0, E.New("out of bound: ", newEnd)
	}
	n, err := io.ReadFull(reader, b.v[currentEnd:newEnd])
	b.end.Store(currentEnd + int32(n))
	return int64(n), err
}

// String returns the string form of this Buffer.
func (b *Buffer) String() string {
	return string(b.Bytes())
}

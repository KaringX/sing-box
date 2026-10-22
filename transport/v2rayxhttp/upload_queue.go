package xhttp

// upload_queue is a specialized priorityqueue + channel to reorder generic
// packets by a sequence number

import (
	"container/heap"
	"io"
	"sync"

	E "github.com/sagernet/sing/common/exceptions"
)

type Packet struct {
	Reader  io.ReadCloser
	Payload []byte
	Seq     uint64
}

type uploadQueue struct {
	reader        io.ReadCloser
	nomore        bool
	pushedPackets chan Packet
	done          chan struct{}
	heap          uploadHeap
	nextSeq       uint64
	closed        bool
	maxPackets    int

	mtx sync.Mutex
}

func NewUploadQueue(maxPackets int) *uploadQueue {
	return &uploadQueue{
		pushedPackets: make(chan Packet, maxPackets),
		done:          make(chan struct{}),
		heap:          uploadHeap{},
		nextSeq:       0,
		closed:        false,
		maxPackets:    maxPackets,
	}
}

func (h *uploadQueue) Push(p Packet) error {
	h.mtx.Lock()
	if h.closed {
		h.mtx.Unlock()
		return E.New("packet queue closed")
	}
	if h.nomore {
		h.mtx.Unlock()
		return E.New("h.reader already exists")
	}
	if p.Reader != nil {
		h.nomore = true
	}
	h.mtx.Unlock()
	select {
	case h.pushedPackets <- p:
		return nil
	case <-h.done:
		return E.New("packet queue closed")
	}
}

func (h *uploadQueue) Close() error {
	h.mtx.Lock()
	if h.closed {
		h.mtx.Unlock()
		return nil
	}
	h.closed = true
	close(h.done)
	h.mtx.Unlock()

	for {
		select {
		case p := <-h.pushedPackets:
			if p.Reader != nil {
				p.Reader.Close()
			}
		default:
			if h.reader != nil {
				return h.reader.Close()
			}
			return nil
		}
	}
}

func (h *uploadQueue) Read(b []byte) (int, error) {
	h.mtx.Lock()
	if h.closed {
		h.mtx.Unlock()
		return 0, io.EOF
	}
	h.mtx.Unlock()
	if h.reader != nil {
		return h.reader.Read(b)
	}
	if len(h.heap) == 0 {
		select {
		case packet, more := <-h.pushedPackets:
			if !more {
				return 0, io.EOF
			}
			if packet.Reader != nil {
				h.mtx.Lock()
				if h.closed {
					packet.Reader.Close()
					h.mtx.Unlock()
					return 0, io.EOF
				}
				h.reader = packet.Reader
				h.mtx.Unlock()
				return h.reader.Read(b)
			}
			heap.Push(&h.heap, packet)
		case <-h.done:
			return 0, io.EOF
		}
	}
	for len(h.heap) > 0 {
		packet := heap.Pop(&h.heap).(Packet)
		n := 0

		if packet.Seq == h.nextSeq {
			copy(b, packet.Payload)
			n = min(len(b), len(packet.Payload))

			if n < len(packet.Payload) {
				// partial read
				packet.Payload = packet.Payload[n:]
				heap.Push(&h.heap, packet)
			} else {
				h.nextSeq = packet.Seq + 1
			}

			return n, nil
		}
		// misordered packet
		if packet.Seq > h.nextSeq {
			if len(h.heap) > h.maxPackets {
				// the "reassembly buffer" is too large, and we want to
				// constrain memory usage somehow. let's tear down the
				// connection, and hope the application retries.
				return 0, E.New("packet queue is too large")
			}
			heap.Push(&h.heap, packet)
			select {
			case packet2, more := <-h.pushedPackets:
				if !more {
					return 0, io.EOF
				}
				heap.Push(&h.heap, packet2)
			case <-h.done:
				return 0, io.EOF
			}
		}
	}
	return 0, nil
}

// heap code directly taken from https://pkg.go.dev/container/heap
type uploadHeap []Packet

func (h uploadHeap) Len() int           { return len(h) }
func (h uploadHeap) Less(i, j int) bool { return h[i].Seq < h[j].Seq }
func (h uploadHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *uploadHeap) Push(x any) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(Packet))
}

func (h *uploadHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

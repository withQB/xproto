package router

import (
	"sync"

	"github.com/withqb/xproto/types"
)

type lifoQueue struct { // nolint:unused
	frames []*types.Frame
	size   int
	count  int
	mutex  sync.Mutex
	notifs chan struct{}
}

func newLIFOQueue(size int) *lifoQueue { // nolint:unused,deadcode
	q := &lifoQueue{
		frames: make([]*types.Frame, size),
		size:   size,
		notifs: make(chan struct{}, size),
	}
	return q
}

func (q *lifoQueue) queuecount() int { // nolint:unused
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return q.count
}

func (q *lifoQueue) queuesize() int { // nolint:unused
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return q.size
}

func (q *lifoQueue) push(frame *types.Frame) bool { // nolint:unused
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if q.count == q.size {
		return false
	}
	index := q.size - q.count - 1
	q.frames[index] = frame
	q.count++
	select {
	case q.notifs <- struct{}{}:
	default:
		panic("this should be impossible")
	}
	return true
}

func (q *lifoQueue) pop() (*types.Frame, bool) { // nolint:unused
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if q.count == 0 {
		return nil, false
	}
	index := q.size - q.count
	frame := q.frames[index]
	q.frames[index] = nil
	q.count--
	return frame, true
}

func (q *lifoQueue) ack() { // nolint:unused
	// no-op on this queue type
}

func (q *lifoQueue) reset() { // nolint:unused
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.count = 0
	for i := range q.frames {
		if q.frames[i] != nil {
			framePool.Put(q.frames[i])
			q.frames[i] = nil
		}
	}
	close(q.notifs)
	for range q.notifs {
	}
	q.notifs = make(chan struct{}, q.size)
}

func (q *lifoQueue) wait() <-chan struct{} { // nolint:unused
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return q.notifs
}

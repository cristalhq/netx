package netx

import (
	"sync/atomic"
)

// atomicCounter is a false sharing safe counter.
type atomicCounter struct {
	count uint64
	_     [7]uint64
}

type cacheLine [64]byte

// Stats object that can be queried to obtain certain metrics and get better observability.
type Stats struct {
	_            cacheLine
	accepts      atomicCounter
	acceptErrors atomicCounter
	activeConns  atomicCounter
	conns        atomicCounter
	closeErrors  atomicCounter

	readCalls    atomicCounter
	readBytes    atomicCounter
	readErrors   atomicCounter
	readTimeouts atomicCounter

	writeCalls    atomicCounter
	writtenBytes  atomicCounter
	writeErrors   atomicCounter
	writeTimeouts atomicCounter

	_ cacheLine
}

func (s *Stats) Accepts() uint64      { return atomic.LoadUint64(&s.accepts.count) }
func (s *Stats) AcceptErrors() uint64 { return atomic.LoadUint64(&s.acceptErrors.count) }
func (s *Stats) Conns() uint64        { return atomic.LoadUint64(&s.conns.count) }
func (s *Stats) CloseErrors() uint64  { return atomic.LoadUint64(&s.closeErrors.count) }

func (s *Stats) ReadCalls() uint64    { return atomic.LoadUint64(&s.readCalls.count) }
func (s *Stats) ReadBytes() uint64    { return atomic.LoadUint64(&s.readBytes.count) }
func (s *Stats) ReadErrors() uint64   { return atomic.LoadUint64(&s.readErrors.count) }
func (s *Stats) ReadTimeouts() uint64 { return atomic.LoadUint64(&s.readTimeouts.count) }

func (s *Stats) WriteCalls() uint64    { return atomic.LoadUint64(&s.writeCalls.count) }
func (s *Stats) WrittenBytes() uint64  { return atomic.LoadUint64(&s.writtenBytes.count) }
func (s *Stats) WriteErrors() uint64   { return atomic.LoadUint64(&s.writeErrors.count) }
func (s *Stats) WriteTimeouts() uint64 { return atomic.LoadUint64(&s.writeTimeouts.count) }

func (s *Stats) acceptsInc()      { atomic.AddUint64(&s.accepts.count, 1) }
func (s *Stats) acceptErrorsInc() { atomic.AddUint64(&s.acceptErrors.count, 1) }
func (s *Stats) activeConnsInc()  { atomic.AddUint64(&s.activeConns.count, 1) }
func (s *Stats) connsInc()        { atomic.AddUint64(&s.conns.count, 1) }
func (s *Stats) closeErrorsInc()  { atomic.AddUint64(&s.closeErrors.count, 1) }

func (s *Stats) readBytesAdd(n int) {
	atomic.AddUint64(&s.readCalls.count, 1)
	atomic.AddUint64(&s.readBytes.count, uint64(n))
}
func (s *Stats) readTimeoutsInc() { atomic.AddUint64(&s.readTimeouts.count, 1) }
func (s *Stats) readErrorsInc()   { atomic.AddUint64(&s.readErrors.count, 1) }

func (s *Stats) writtenBytesAdd(n int) {
	atomic.AddUint64(&s.writeCalls.count, 1)
	atomic.AddUint64(&s.writtenBytes.count, uint64(n))
}
func (s *Stats) writeTimeoutsInc() { atomic.AddUint64(&s.writeTimeouts.count, 1) }
func (s *Stats) writeErrorsInc()   { atomic.AddUint64(&s.writeErrors.count, 1) }

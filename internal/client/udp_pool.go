package client

import (
	"hash/maphash"
	"paqet/internal/flog"
	"paqet/internal/tr"
	"sync"
	"time"
)

type udpSess struct {
	strm       tr.Strm
	lastActive time.Time
}

type udpPool struct {
	sesses map[uint64]*udpSess
	mu     sync.RWMutex
	hasher maphash.Hash
}

func (p *udpPool) sessKey(localAddr, targetAddr string) uint64 {
	p.hasher.Reset()
	p.hasher.WriteString(localAddr)
	p.hasher.WriteString(targetAddr)
	return p.hasher.Sum64()
}

func (c *Client) CloseUDP(key uint64) error {
	return c.udpPool.closeSess(key)
}

func (p *udpPool) closeSess(key uint64) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if sess, exists := p.sesses[key]; exists {
		flog.Debugf("closing UDP session stream %d", sess.strm.SID())
		sess.strm.Close()
	} else {
		flog.Debugf("UDP session key %d not found for close", key)
	}
	delete(p.sesses, key)

	return nil
}

func (p *udpPool) close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.sesses) > 0 {
		flog.Infof("closing %d UDP sessions", len(p.sesses))
		for _, sess := range p.sesses {
			sess.strm.Close()
		}
		p.sesses = make(map[uint64]*udpSess)
	}
	return nil
}

func (p *udpPool) startGC() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			p.sweep()
		}
	}()
}

func (p *udpPool) sweep() {
	p.mu.Lock()
	defer p.mu.Unlock()

	cutoff := time.Now().Add(-2 * time.Minute)
	for key, sess := range p.sesses {
		if sess.lastActive.Before(cutoff) {
			flog.Debugf("cleaning up idle UDP session %d (last active: %v)", sess.strm.SID(), sess.lastActive)
			sess.strm.Close()
			delete(p.sesses, key)
		}
	}
}

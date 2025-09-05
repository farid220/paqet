package client

import (
	"net"
	"paqet/internal/flog"
	"paqet/internal/protocol"
	"paqet/internal/tr"
	"time"
)

func (c *Client) UDP(lAddr, tAddr string) (tr.Strm, bool, uint64, error) {
	key := c.udpPool.sessKey(lAddr, tAddr)
	c.udpPool.mu.RLock()
	if sess, exists := c.udpPool.sesses[key]; exists {
		sess.lastActive = time.Now()
		c.udpPool.mu.RUnlock()
		flog.Debugf("reusing UDP stream %d for %s -> %s", sess.strm.SID(), lAddr, tAddr)
		return sess.strm, false, key, nil
	}
	c.udpPool.mu.RUnlock()
	flog.Debugf("creating new UDP stream for %s -> %s", lAddr, tAddr)

	strm, err := c.newStrm()
	if err != nil {
		flog.Debugf("failed to create stream for UDP %s -> %s: %v", lAddr, tAddr, err)
		return nil, false, 0, err
	}

	taddr, err := net.ResolveUDPAddr("udp", tAddr)
	if err != nil {
		flog.Debugf("invalid UDP address %s: %v", tAddr, err)
		strm.Close()
		return nil, false, 0, err
	}
	p := protocol.Proto{Type: protocol.PUDP, Addr: taddr}
	err = p.Write(strm)
	if err != nil {
		flog.Debugf("failed to write UDP protocol header for %s -> %s on stream %d: %v", lAddr, tAddr, strm.SID(), err)
		strm.Close()
		return nil, false, 0, err
	}

	c.udpPool.mu.Lock()
	udpSess := udpSess{
		strm:       strm,
		lastActive: time.Now(),
	}
	c.udpPool.sesses[key] = &udpSess
	c.udpPool.mu.Unlock()

	flog.Debugf("established UDP stream %d for %s -> %s", strm.SID(), lAddr, tAddr)
	return strm, true, key, nil
}

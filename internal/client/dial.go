package client

import (
	"paqet/internal/flog"
	"paqet/internal/tr"
	"time"
)

func (c *Client) newConn() (tr.Conn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	autoExpire := 300
	tc := c.iter.Next()
	go tc.sendTCPF(tc.conn)
	err := tc.conn.Ping(false)
	if err != nil {
		flog.Infof("connection lost, retrying....")
		if tc.conn != nil {
			tc.conn.Close()
		}
		tc.conn = tc.waitConn()
		tc.expire = time.Now().Add(time.Duration(autoExpire) * time.Second)
	}
	return tc.conn, nil
}

func (c *Client) newStrm() (tr.Strm, error) {
	conn, err := c.newConn()
	if err != nil {
		flog.Debugf("session creation failed, retrying")
		return c.newStrm()
	}
	strm, err := conn.OpenStrm()
	if err != nil {
		flog.Debugf("failed to open stream, retrying: %v", err)
		return c.newStrm()
	}
	flog.Debugf("new stream %d created successfully", strm.SID())
	return strm, nil
}

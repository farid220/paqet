package client

import (
	"context"
	"paqet/internal/conf"
	"paqet/internal/flog"
	"paqet/internal/pkg/iterator"
	"sync"
)

type Client struct {
	cfg     *conf.Conf
	iter    *iterator.Iterator[*timedConn]
	udpPool *udpPool
	mu      sync.Mutex
}

func New(cfg *conf.Conf) (*Client, error) {
	c := &Client{
		cfg:     cfg,
		iter:    &iterator.Iterator[*timedConn]{},
		udpPool: &udpPool{sesses: make(map[uint64]*udpSess)},
	}
	c.udpPool.startGC()
	return c, nil
}

func (c *Client) Start(ctx context.Context) error {
	for i := range c.cfg.Transport.Conn {
		tc, err := newTimedConn(ctx, c.cfg)
		if err != nil {
			flog.Errorf("failed to establish connection %d: %v", i+1, err)
			return err
		}
		flog.Debugf("client connection %d established successfully", i+1)
		c.iter.Items = append(c.iter.Items, tc)
	}
	go func() {
		<-ctx.Done()
		for _, tc := range c.iter.Items {
			tc.Close()
		}
		c.udpPool.close()
		flog.Infof("client shutdown complete")
	}()

	flog.Infof("Client started: %s -> %s (%d connections)", &c.cfg.Network.LocalAddr, c.cfg.Server.Addr, len(c.iter.Items))
	return nil
}

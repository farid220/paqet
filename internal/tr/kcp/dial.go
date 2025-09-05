package kcp

import (
	"fmt"
	"net"
	"paqet/internal/conf"
	"paqet/internal/flog"
	"paqet/internal/pconn"
	"paqet/internal/tr"

	"github.com/xtaci/kcp-go/v5"
	"github.com/xtaci/smux"
)

func Dial(addr *net.UDPAddr, cfg *conf.KCP, pConn *pconn.PacketConn) (tr.Conn, error) {
	block, err := newBlock(cfg.Block, cfg.Key)
	if err != nil {
		return nil, err
	}
	flog.Debugf("creating KCP connection to %s", addr)
	conn, err := kcp.NewConn(addr.String(), block, cfg.Dshard, cfg.Pshard, pConn)
	if err != nil {
		return nil, fmt.Errorf("connection attempt failed: %v", err)
	}
	aplConf(conn, cfg)
	flog.Debugf("KCP connection established, creating smux session")

	sess, err := smux.Client(conn, smuxConf(cfg))
	if err != nil {
		return nil, fmt.Errorf("failed to create smux session: %w", err)
	}
	flog.Debugf("smux session established successfully")
	return &Conn{conn, sess}, nil
}

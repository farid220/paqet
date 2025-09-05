package client

import (
	"context"
	"fmt"
	"paqet/internal/conf"
	"paqet/internal/pconn"
	"paqet/internal/protocol"
	"paqet/internal/tr"
	"paqet/internal/tr/kcp"
	"time"
)

type timedConn struct {
	cfg    *conf.Conf
	conn   tr.Conn
	expire time.Time
	ctx    context.Context
}

func newTimedConn(ctx context.Context, cfg *conf.Conf) (*timedConn, error) {
	var err error
	tc := timedConn{cfg: cfg, ctx: ctx}
	tc.conn, err = tc.createConn()
	if err != nil {
		return nil, err
	}

	return &tc, nil
}

func (tc *timedConn) createConn() (tr.Conn, error) {
	netCfg := tc.cfg.Network
	pConn, err := pconn.New(tc.ctx, &netCfg)
	if err != nil {
		return nil, fmt.Errorf("could not create raw packet conn: %w", err)
	}

	conn, err := kcp.Dial(tc.cfg.Server.Addr, tc.cfg.Transport.KCP, pConn)
	if err != nil {
		return nil, err
	}
	err = tc.sendTCPF(conn)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (tc *timedConn) waitConn() tr.Conn {
	for {
		if c, err := tc.createConn(); err == nil {
			return c
		} else {
			time.Sleep(time.Second)
		}
	}
}

func (tc *timedConn) sendTCPF(conn tr.Conn) error {
	strm, err := conn.OpenStrm()
	if err != nil {
		return err
	}
	p := protocol.Proto{Type: protocol.PTCPF, TCPF: tc.cfg.Network.TCP.RF}
	err = p.Write(strm)
	if err != nil {
		return err
	}
	return nil
}

// func (tc *timedConn) sendTCPFP(ctx context.Context) {
// 	ticker := time.NewTicker(1 * time.Second)
// 	defer ticker.Stop()

// 	for {
// 		fmt.Println("ticker start ticked.")
// 		select {
// 		case <-ticker.C:
// 			err := tc.sendTCPF(tc.conn)
// 			if err != nil {
// 				fmt.Println(err)
// 			}
// 			fmt.Println("ticker ticked.")
// 		case <-ctx.Done():
// 			fmt.Println("ctx ticker ticked.")
// 			return
// 		}
// 	}
// }

func (tc *timedConn) Close() {
	if tc.conn != nil {
		tc.conn.Close()
	}
}

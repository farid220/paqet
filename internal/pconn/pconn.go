package pconn

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"os"
	"paqet/internal/conf"
	"paqet/internal/flog"
	"paqet/internal/pkg/iterator"
	"sync/atomic"
	"time"
)

type PacketConn struct {
	cfg           *conf.Network
	sendHandle    *SendHandle
	recvHandle    *RecvHandle
	readDeadline  atomic.Value
	writeDeadline atomic.Value

	ctx    context.Context
	cancel context.CancelFunc
}

// &OpError{Op: "listen", Net: network, Source: nil, Addr: nil, Err: err}
func New(ctx context.Context, cfg *conf.Network) (*PacketConn, error) {
	if cfg.LocalAddr.Port == 0 {
		cfg.LocalAddr.Port = 32768 + rand.Intn(32768)
	}
	flog.Warnf("PCONN Port: %d", cfg.LocalAddr.Port)
	sendHandle, err := NewSendHandle(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create send handle on %s: %v", cfg.Interface.Name, err)
	}

	recvHandle, err := NewRecvHandle(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create receive handle on %s: %v", cfg.Interface.Name, err)
	}

	ctx, cancel := context.WithCancel(ctx)
	conn := &PacketConn{
		cfg:        cfg,
		sendHandle: sendHandle,
		recvHandle: recvHandle,
		ctx:        ctx,
		cancel:     cancel,
	}

	return conn, nil
}

func (c *PacketConn) ReadFrom(data []byte) (n int, addr net.Addr, err error) {
	var timer *time.Timer
	var deadline <-chan time.Time
	if d, ok := c.readDeadline.Load().(time.Time); ok && !d.IsZero() {
		timer = time.NewTimer(time.Until(d))
		defer timer.Stop()
		deadline = timer.C
	}

	select {
	case <-c.ctx.Done():
		return 0, nil, c.ctx.Err()
	case <-deadline:
		return 0, nil, os.ErrDeadlineExceeded
	default:
	}

	payload, addr, err := c.recvHandle.Read()
	if err != nil {
		return 0, nil, err
	}
	n = copy(data, payload)

	return n, addr, nil
}

func (c *PacketConn) WriteTo(data []byte, addr net.Addr) (n int, err error) {
	var timer *time.Timer
	var deadline <-chan time.Time
	if d, ok := c.writeDeadline.Load().(time.Time); ok && !d.IsZero() {
		timer = time.NewTimer(time.Until(d))
		defer timer.Stop()
		deadline = timer.C
	}

	select {
	case <-c.ctx.Done():
		return 0, c.ctx.Err()
	case <-deadline:
		return 0, os.ErrDeadlineExceeded
	default:
	}

	daddr, ok := addr.(*net.UDPAddr)
	if !ok {
		return 0, net.InvalidAddrError("invalid address")
	}

	err = c.sendHandle.Write(data, daddr)
	if err != nil {
		return 0, err
	}

	return len(data), nil
}

func (c *PacketConn) Close() error {
	c.cancel()

	if c.sendHandle != nil {
		go c.sendHandle.Close()
	}
	if c.recvHandle != nil {
		go c.recvHandle.Close()
	}

	return nil
}

func (c *PacketConn) LocalAddr() net.Addr {
	return &net.UDPAddr{
		IP:   append([]byte(nil), c.cfg.LocalAddr.IP...),
		Port: c.cfg.LocalAddr.Port,
		Zone: c.cfg.LocalAddr.Zone,
	}
}

func (c *PacketConn) SetDeadline(t time.Time) error {
	c.readDeadline.Store(t)
	c.writeDeadline.Store(t)
	return nil
}

func (c *PacketConn) SetReadDeadline(t time.Time) error {
	c.readDeadline.Store(t)
	return nil
}

func (c *PacketConn) SetWriteDeadline(t time.Time) error {
	c.writeDeadline.Store(t)
	return nil
}

func (c *PacketConn) SetDSCP(dscp int) error {
	return nil
}

func (c *PacketConn) SetClientTCPF(addr net.Addr, f []conf.TCPF) {
	a := *addr.(*net.UDPAddr)
	c.sendHandle.cTCPF[hashIPAddr(a.IP, uint16(a.Port))] = &iterator.Iterator[conf.TCPF]{Items: f}
}

func hashIPAddr(ip net.IP, port uint16) uint8 {
	if len(ip) == 4 {
		hash := uint64(binary.BigEndian.Uint32(ip))<<16 | uint64(port)
		return uint8(hash)
	}
	ip16 := ip.To16()
	hash := binary.BigEndian.Uint64(ip16[0:8]) ^ binary.BigEndian.Uint64(ip16[8:16])
	hash = hash ^ (uint64(port) << 48)
	return uint8(hash)
}

package conf

import (
	"fmt"
	"time"
)

type QUIC struct {
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
	CAFile   string `yaml:"ca_file"`
	Insecure bool   `yaml:"insecure"`

	MaxIdleTimeout       int `yaml:"max_idle_timeout"`
	KeepAlivePeriod      int `yaml:"keep_alive_period"`
	HandshakeIdleTimeout int `yaml:"handshake_idle_timeout"`

	InitialStreamReceiveWindow     int64 `yaml:"initial_stream_receive_window"`
	MaxStreamReceiveWindow         int64 `yaml:"max_stream_receive_window"`
	InitialConnectionReceiveWindow int64 `yaml:"initial_connection_receive_window"`
	MaxConnectionReceiveWindow     int64 `yaml:"max_connection_receive_window"`

	MaxIncomingStreams    int64 `yaml:"max_incoming_streams"`
	MaxIncomingUniStreams int64 `yaml:"max_incoming_uni_streams"`

	Allow0RTT               bool `yaml:"allow_0rtt"`
	EnableDatagrams         bool `yaml:"enable_datagrams"`
	DisablePathMTUDiscovery bool `yaml:"disable_path_mtu_discovery"`
}

func (q *QUIC) setDefaults(role string) {
	if q.MaxIdleTimeout == 0 {
		if role == "server" {
			q.MaxIdleTimeout = 60
		} else {
			q.MaxIdleTimeout = 30
		}
	}
	if q.KeepAlivePeriod == 0 {
		q.KeepAlivePeriod = 10
	}
	if q.HandshakeIdleTimeout == 0 {
		if role == "server" {
			q.HandshakeIdleTimeout = 10
		} else {
			q.HandshakeIdleTimeout = 5
		}
	}

	if q.InitialStreamReceiveWindow == 0 {
		q.InitialStreamReceiveWindow = 2 * 1024 * 1024
	}
	if q.MaxStreamReceiveWindow == 0 {
		q.MaxStreamReceiveWindow = 2 * 1024 * 1024
	}
	if q.InitialConnectionReceiveWindow == 0 {
		q.InitialConnectionReceiveWindow = 4 * 1024 * 1024
	}
	if q.MaxConnectionReceiveWindow == 0 {
		q.MaxConnectionReceiveWindow = 4 * 1024 * 1024
	}

	if q.MaxIncomingStreams == 0 {
		if role == "server" {
			q.MaxIncomingStreams = 1024 * 1024
		} else {
			q.MaxIncomingStreams = 1024
		}
	}
	if q.MaxIncomingUniStreams == 0 {
		if role == "server" {
			q.MaxIncomingUniStreams = 1024
		} else {
			q.MaxIncomingUniStreams = 128
		}
	}

	if !q.Allow0RTT {
		q.Allow0RTT = true
	}
}

func (q *QUIC) validate() []error {
	var errors []error

	if q.MaxIdleTimeout < 1 || q.MaxIdleTimeout > 3600 {
		errors = append(errors, fmt.Errorf("QUIC max_idle_timeout (%d) must be between 1-3600 seconds", q.MaxIdleTimeout))
	}
	if q.KeepAlivePeriod < 1 || q.KeepAlivePeriod > 3600 {
		errors = append(errors, fmt.Errorf("QUIC keep_alive_period (%d) must be between 1-3600 seconds", q.KeepAlivePeriod))
	}
	if q.HandshakeIdleTimeout < 1 || q.HandshakeIdleTimeout > 300 {
		errors = append(errors, fmt.Errorf("QUIC handshake_idle_timeout (%d) must be between 1-300 seconds", q.HandshakeIdleTimeout))
	}

	if q.KeepAlivePeriod >= q.MaxIdleTimeout {
		errors = append(errors, fmt.Errorf("QUIC keep_alive_period (%d) should be less than max_idle_timeout (%d)", q.KeepAlivePeriod, q.MaxIdleTimeout))
	}

	if q.InitialStreamReceiveWindow < 1024 {
		errors = append(errors, fmt.Errorf("QUIC initial_stream_receive_window must be >= 1024 bytes"))
	}
	if q.MaxStreamReceiveWindow < q.InitialStreamReceiveWindow {
		errors = append(errors, fmt.Errorf("QUIC max_stream_receive_window must be >= initial_stream_receive_window"))
	}
	if q.InitialConnectionReceiveWindow < 1024 {
		errors = append(errors, fmt.Errorf("QUIC initial_connection_receive_window must be >= 1024 bytes"))
	}
	if q.MaxConnectionReceiveWindow < q.InitialConnectionReceiveWindow {
		errors = append(errors, fmt.Errorf("QUIC max_connection_receive_window must be >= initial_connection_receive_window"))
	}

	if q.MaxIncomingStreams < 1 {
		errors = append(errors, fmt.Errorf("QUIC max_incoming_streams must be > 0"))
	}
	if q.MaxIncomingUniStreams < 1 {
		errors = append(errors, fmt.Errorf("QUIC max_incoming_uni_streams must be > 0"))
	}

	return errors
}

func (q *QUIC) ToQuicConfig() any {
	return map[string]any{
		"MaxIdleTimeout":                 time.Duration(q.MaxIdleTimeout) * time.Second,
		"KeepAlivePeriod":                time.Duration(q.KeepAlivePeriod) * time.Second,
		"HandshakeIdleTimeout":           time.Duration(q.HandshakeIdleTimeout) * time.Second,
		"InitialStreamReceiveWindow":     uint64(q.InitialStreamReceiveWindow),
		"MaxStreamReceiveWindow":         uint64(q.MaxStreamReceiveWindow),
		"InitialConnectionReceiveWindow": uint64(q.InitialConnectionReceiveWindow),
		"MaxConnectionReceiveWindow":     uint64(q.MaxConnectionReceiveWindow),
		"MaxIncomingStreams":             q.MaxIncomingStreams,
		"MaxIncomingUniStreams":          q.MaxIncomingUniStreams,
		"Allow0RTT":                      q.Allow0RTT,
		"EnableDatagrams":                q.EnableDatagrams,
		"DisablePathMTUDiscovery":        q.DisablePathMTUDiscovery,
	}
}

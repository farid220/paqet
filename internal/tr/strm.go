package tr

import (
	"net"
)

type Strm interface {
	net.Conn
	SID() int
}

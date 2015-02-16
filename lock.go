package gobfile

import (
	"fmt"
	"net"
	"time"
)

type PortLocker struct {
	port int
	ln   net.Listener
}

func (l *PortLocker) Lock() {
	for {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", l.port))
		if err != nil {
			time.Sleep(time.Second * 1)
			continue
		}
		l.ln = ln
		break
	}
}

func (l *PortLocker) Unlock() {
	l.ln.Close()
}

func NewPortLocker(port int) *PortLocker {
	return &PortLocker{
		port: port,
	}
}

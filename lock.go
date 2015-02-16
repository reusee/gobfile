package gobfile

import (
	"fmt"
	"net"
	"os"
	"syscall"
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

type FileLocker struct {
	file *os.File
}

func NewFileLocker(path string) *FileLocker {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(fmt.Sprintf("open lock file %s error %v", path, err))
	}
	return &FileLocker{
		file: file,
	}
}

func (l *FileLocker) Lock() {
	syscall.Flock(int(l.file.Fd()), syscall.LOCK_EX)
}

func (l *FileLocker) Unlock() {
	syscall.Flock(int(l.file.Fd()), syscall.LOCK_UN)
	l.file.Close()
}

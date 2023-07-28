package counter

import (
	"net"

	"github.com/sagernet/sing/common/buf"

	M "github.com/sagernet/sing/common/metadata"
	"github.com/sagernet/sing/common/network"
)

type ConnCounter struct {
	net.Conn
	storage *TrafficStorage
}

func NewConnCounter(conn net.Conn, s *TrafficStorage) net.Conn {
	return &ConnCounter{
		Conn:    conn,
		storage: s,
	}
}

func (c *ConnCounter) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	c.storage.DownCounter.Store(int64(n))
	return
}

func (c *ConnCounter) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	c.storage.UpCounter.Store(int64(n))
	return
}

type PacketConnCounter struct {
	network.PacketConn
	storage *TrafficStorage
}

func NewPacketConnCounter(conn network.PacketConn, s *TrafficStorage) network.PacketConn {
	return &PacketConnCounter{
		PacketConn: conn,
		storage:    s,
	}
}

func (p *PacketConnCounter) ReadPacket(buff *buf.Buffer) (destination M.Socksaddr, err error) {
	destination, err = p.PacketConn.ReadPacket(buff)
	if err != nil {
		return
	}
	p.storage.DownCounter.Add(int64(buff.Len()))
	return
}

func (p *PacketConnCounter) WritePacket(buff *buf.Buffer, destination M.Socksaddr) (err error) {
	err = p.PacketConn.WritePacket(buff, destination)
	if err != nil {
		return
	}
	p.storage.UpCounter.Add(int64(buff.Len()))
	return
}

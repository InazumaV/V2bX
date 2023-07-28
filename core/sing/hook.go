package sing

import (
	"net"

	"github.com/Yuzuki616/V2bX/common/counter"
	"github.com/inazumav/sing-box/adapter"
	"github.com/inazumav/sing-box/log"
	N "github.com/sagernet/sing/common/network"
)

type HookServer struct {
	hooker *Hooker
}

func NewHookServer(logger log.Logger) *HookServer {
	return &HookServer{
		hooker: &Hooker{
			logger:  logger,
			counter: make(map[string]*counter.TrafficCounter),
		},
	}
}

func (h *HookServer) Start() error {
	return nil
}

func (h *HookServer) Close() error {
	return nil
}

func (h *HookServer) StatsService() adapter.V2RayStatsService {
	return h.hooker
}

func (h *HookServer) Hooker() *Hooker {
	return h.hooker
}

type Hooker struct {
	logger  log.Logger
	counter map[string]*counter.TrafficCounter
}

func (h *Hooker) RoutedConnection(inbound string, outbound string, user string, conn net.Conn) net.Conn {
	if c, ok := h.counter[inbound]; ok {
		return counter.NewConnCounter(conn, c.GetCounter(user))
	}
	return conn
}

func (h *Hooker) RoutedPacketConnection(inbound string, outbound string, user string, conn N.PacketConn) N.PacketConn {
	if c, ok := h.counter[inbound]; ok {
		return counter.NewPacketConnCounter(conn, c.GetCounter(user))
	}
	return conn
}

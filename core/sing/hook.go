package sing

import (
	"net"
	"strings"
	"sync"

	"github.com/InazumaV/V2bX/common/rate"

	"github.com/InazumaV/V2bX/limiter"

	"github.com/InazumaV/V2bX/common/counter"
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
			counter: sync.Map{},
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
	counter sync.Map
}

func (h *Hooker) RoutedConnection(metadata adapter.InboundContext, outbound string, user string, conn net.Conn) net.Conn {
	inbound := metadata.Inbound
	l, err := limiter.GetLimiter(inbound)
	if err != nil {
		log.Error("get limiter for ", inbound, " error: ", err)
	}
	if l.CheckDomainRule(metadata.Domain) {
		conn.Close()
		h.logger.Error("[", inbound, "] ", "Limited ", user, " access domain ", metadata.Domain, " reject by rule")
		return conn
	}
	ip, _, _ := strings.Cut(conn.RemoteAddr().String(), ":")
	if b, r := l.CheckLimit(user, ip, true); r {
		conn.Close()
		h.logger.Error("[", inbound, "] ", "Limited ", user, " by ip or conn")
		return conn
	} else if b != nil {
		conn = rate.NewConnRateLimiter(conn, b)
	}
	if c, ok := h.counter.Load(inbound); ok {
		return counter.NewConnCounter(conn, c.(*counter.TrafficCounter).GetCounter(user))
	} else {
		c := counter.NewTrafficCounter()
		h.counter.Store(inbound, c)
		return counter.NewConnCounter(conn, c.GetCounter(user))
	}
}

func (h *Hooker) RoutedPacketConnection(metadata adapter.InboundContext, outbound string, user string, conn N.PacketConn) N.PacketConn {
	inbound := metadata.Inbound
	l, err := limiter.GetLimiter(inbound)
	if err != nil {
		log.Error("get limiter for ", inbound, " error: ", err)
	}
	if l.CheckDomainRule(metadata.Domain) {
		conn.Close()
		h.logger.Error("[", inbound, "] ", "Limited ", user, " access domain ", metadata.Domain, " reject by rule")
		return conn
	}
	if c, ok := h.counter.Load(inbound); ok {
		return counter.NewPacketConnCounter(conn, c.(*counter.TrafficCounter).GetCounter(user))
	} else {
		c := counter.NewTrafficCounter()
		h.counter.Store(inbound, c)
		return counter.NewPacketConnCounter(conn, c.GetCounter(user))
	}
}

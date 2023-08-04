package sing

import (
	"context"
	"github.com/inazumav/sing-box/common/urltest"
	"net"
	"sync"

	"github.com/InazumaV/V2bX/common/rate"

	"github.com/InazumaV/V2bX/limiter"

	"github.com/InazumaV/V2bX/common/counter"
	"github.com/inazumav/sing-box/adapter"
	"github.com/inazumav/sing-box/log"
	N "github.com/sagernet/sing/common/network"
)

type HookServer struct {
	logger  log.Logger
	counter sync.Map
}

func NewHookServer(logger log.Logger) *HookServer {
	return &HookServer{
		logger:  logger,
		counter: sync.Map{},
	}
}

func (h *HookServer) Start() error {
	return nil
}

func (h *HookServer) Close() error {
	return nil
}

func (h *HookServer) PreStart() error {
	return nil
}

func (h *HookServer) RoutedConnection(_ context.Context, conn net.Conn, m adapter.InboundContext, _ adapter.Rule) (net.Conn, adapter.Tracker) {
	l, err := limiter.GetLimiter(m.Inbound)
	if err != nil {
		log.Error("get limiter for ", m.Inbound, " error: ", err)
	}
	if l.CheckDomainRule(m.Domain) {
		conn.Close()
		h.logger.Error("[", m.Inbound, "] ",
			"Limited ", m.User, "access to", m.Domain, " by domain rule")
		return conn, &Tracker{l: func() {}}
	}
	if l.CheckProtocolRule(m.Protocol) {
		conn.Close()
		h.logger.Error("[", m.Inbound, "] ",
			"Limited ", m.User, "use", m.Domain, " by protocol rule")
		return conn, &Tracker{l: func() {}}
	}
	ip := m.Source.Addr.String()
	if b, r := l.CheckLimit(m.User, ip, true); r {
		conn.Close()
		h.logger.Error("[", m.Inbound, "] ", "Limited ", m.User, " by ip or conn")
		return conn, &Tracker{l: func() {}}
	} else if b != nil {
		conn = rate.NewConnRateLimiter(conn, b)
	}
	t := &Tracker{
		l: func() {
			l.ConnLimiter.DelConnCount(m.User, ip)
		},
	}
	if c, ok := h.counter.Load(m.Inbound); ok {
		return counter.NewConnCounter(conn, c.(*counter.TrafficCounter).GetCounter(m.Inbound)), t
	} else {
		c := counter.NewTrafficCounter()
		h.counter.Store(m.Inbound, c)
		return counter.NewConnCounter(conn, c.GetCounter(m.Inbound)), t
	}
}

func (h *HookServer) RoutedPacketConnection(_ context.Context, conn N.PacketConn, m adapter.InboundContext, _ adapter.Rule) (N.PacketConn, adapter.Tracker) {
	t := &Tracker{
		l: func() {},
	}
	l, err := limiter.GetLimiter(m.Inbound)
	if err != nil {
		log.Error("get limiter for ", m.Inbound, " error: ", err)
	}
	if l.CheckDomainRule(m.Domain) {
		conn.Close()
		h.logger.Error("[", m.Inbound, "] ",
			"Limited ", m.User, "access to", m.Domain, " by domain rule")
		return conn, t
	}
	if l.CheckProtocolRule(m.Protocol) {
		conn.Close()
		h.logger.Error("[", m.Inbound, "] ",
			"Limited ", m.User, "use", m.Domain, " by protocol rule")
		return conn, t
	}
	ip := m.Source.Addr.String()
	if b, r := l.CheckLimit(m.User, ip, true); r {
		conn.Close()
		h.logger.Error("[", m.Inbound, "] ", "Limited ", m.User, " by ip or conn")
		return conn, &Tracker{l: func() {}}
	} else if b != nil {
		conn = rate.NewPacketConnCounter(conn, b)
	}
	if c, ok := h.counter.Load(m.Inbound); ok {
		return counter.NewPacketConnCounter(conn, c.(*counter.TrafficCounter).GetCounter(m.User)), t
	} else {
		c := counter.NewTrafficCounter()
		h.counter.Store(m.Inbound, c)
		return counter.NewPacketConnCounter(conn, c.GetCounter(m.User)), t
	}
}

// not need

func (h *HookServer) Mode() string {
	return ""
}
func (h *HookServer) StoreSelected() bool {
	return false
}
func (h *HookServer) CacheFile() adapter.ClashCacheFile {
	return nil
}
func (h *HookServer) HistoryStorage() *urltest.HistoryStorage {
	return nil
}

func (h *HookServer) StoreFakeIP() bool {
	return false
}

type Tracker struct {
	l func()
}

func (t *Tracker) Leave() {
	t.l()
}

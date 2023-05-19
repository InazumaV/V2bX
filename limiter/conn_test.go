package limiter

import (
	"sync"
	"testing"
	"time"
)

var c *ConnLimiter

func init() {
	c = NewConnLimiter(1, 1, true)
}

func TestConnLimiter_AddConnCount(t *testing.T) {
	t.Log(c.AddConnCount("1", "1", true))
	t.Log(c.AddConnCount("1", "2", true))
}

func TestConnLimiter_DelConnCount(t *testing.T) {
	t.Log(c.AddConnCount("1", "1", true))
	t.Log(c.AddConnCount("1", "2", true))
	c.DelConnCount("1", "1")
	t.Log(c.AddConnCount("1", "2", true))
}

func TestConnLimiter_ClearOnlineIP(t *testing.T) {
	t.Log(c.AddConnCount("1", "1", false))
	t.Log(c.AddConnCount("1", "2", false))
	c.ClearOnlineIP()
	t.Log(c.AddConnCount("1", "2", true))
	c.DelConnCount("1", "2")
	t.Log(c.AddConnCount("1", "1", false))
	// not realtime
	c.realtime = false
	t.Log(c.AddConnCount("3", "2", true))
	c.ClearOnlineIP()
	t.Log(c.ip.Load("3"))
	time.Sleep(time.Minute)
	c.ClearOnlineIP()
	t.Log(c.ip.Load("3"))
}

func BenchmarkConnLimiter(b *testing.B) {
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			c.AddConnCount("1", "2", true)
			c.DelConnCount("1", "2")
			wg.Done()
		}()
	}
	wg.Wait()

}

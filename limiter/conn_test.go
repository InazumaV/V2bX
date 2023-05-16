package limiter

import (
	"sync"
	"testing"
)

var c *ConnLimiter

func init() {
	c = NewConnLimiter(1, 1)
}

func TestConnLimiter_AddConnCount(t *testing.T) {
	t.Log(c.AddConnCount("1", "1"))
	t.Log(c.AddConnCount("1", "2"))
}

func TestConnLimiter_DelConnCount(t *testing.T) {
	t.Log(c.AddConnCount("1", "1"))
	t.Log(c.AddConnCount("1", "2"))
	c.DelConnCount("1", "1")
	t.Log(c.AddConnCount("1", "2"))
}

func BenchmarkConnLimiter(b *testing.B) {
	wg := sync.WaitGroup{}
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			c.AddConnCount("1", "2")
			c.DelConnCount("1", "2")
			wg.Done()
		}()
	}
	wg.Wait()

}

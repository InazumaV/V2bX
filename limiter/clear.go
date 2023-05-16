package limiter

import "log"

func ClearPacketOnlineIP() error {
	log.Println("Limiter: Clear packet online ip...")
	limitLock.RLock()
	for _, l := range limiter {
		l.ConnLimiter.ClearPacketOnlineIP()
	}
	limitLock.RUnlock()
	log.Println("Limiter: Clear packet online ip done")
	return nil
}

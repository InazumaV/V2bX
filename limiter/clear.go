package limiter

import "log"

func ClearOnlineIP() error {
	log.Println("Limiter: Clear online ip...")
	limitLock.RLock()
	for _, l := range limiter {
		l.ConnLimiter.ClearOnlineIP()
	}
	limitLock.RUnlock()
	log.Println("Limiter: Clear online ip done")
	return nil
}

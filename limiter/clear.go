package limiter

import log "github.com/sirupsen/logrus"

func ClearOnlineIP() error {
	log.WithField("Type", "Limiter").
		Debug("Clear online ip...")
	limitLock.RLock()
	for _, l := range limiter {
		l.ConnLimiter.ClearOnlineIP()
	}
	limitLock.RUnlock()
	log.WithField("Type", "Limiter").
		Debug("Clear online ip done")
	return nil
}

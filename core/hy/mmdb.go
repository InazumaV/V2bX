package hy

import (
	"os"

	"github.com/oschwald/geoip2-golang"
	"github.com/sirupsen/logrus"
)

func loadMMDBReader(filename string) (*geoip2.Reader, error) {
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			logrus.Info("GeoLite2 database not found, downloading...")
			logrus.WithFields(logrus.Fields{
				"file": filename,
			}).Info("GeoLite2 database downloaded")
			return geoip2.Open(filename)
		} else {
			// some other error
			return nil, err
		}
	} else {
		// file exists, just open it
		return geoip2.Open(filename)
	}
}

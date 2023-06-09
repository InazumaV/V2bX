package hy

const (
	mbpsToBps   = 125000
	minSpeedBPS = 16384

	DefaultALPN = "hysteria"

	DefaultStreamReceiveWindow     = 16777216                           // 16 MB
	DefaultConnectionReceiveWindow = DefaultStreamReceiveWindow * 5 / 2 // 40 MB

	DefaultMaxIncomingStreams = 1024

	DefaultMMDBFilename = "GeoLite2-Country.mmdb"

	ServerMaxIdleTimeoutSec     = 60
	DefaultClientIdleTimeoutSec = 20

	DefaultClientHopIntervalSec = 10
)

func SpeedTrans(upM, downM int) (uint64, uint64) {
	up := uint64(upM) * mbpsToBps
	down := uint64(downM) * mbpsToBps
	return up, down
}

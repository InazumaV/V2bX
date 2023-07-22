//go:build !(windows || linux || darwin)

package systime

import (
	"os"
	"time"
)

func SetSystemTime(nowTime time.Time) error {
	return os.ErrInvalid
}

//go:build !linux

package sing

func rusageMaxRSS() float64 {
	return -1
}

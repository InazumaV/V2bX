package file

import "os"

func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

package cmd

import (
	"fmt"
	"strings"
)

const (
	red    = "\033[0;31m"
	green  = "\033[0;32m"
	yellow = "\033[0;33m"
	plain  = "\033[0m"
)

func checkRunning() (bool, error) {
	o, err := execCommand("systemctl status V2bX | grep Active")
	if err != nil {
		return false, err
	}
	return strings.Contains(o, "running"), nil
}

func Err(msg ...any) string {
	return red + fmt.Sprint(msg...) + plain
}

func Ok(msg ...any) string {
	return green + fmt.Sprint(msg...) + plain
}

func Warn(msg ...any) string {
	return yellow + fmt.Sprint(msg...) + plain
}

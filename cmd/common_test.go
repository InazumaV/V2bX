package cmd

import "testing"

func TestExecCommand(t *testing.T) {
	t.Log(execCommand("echo test"))
}

func Test_printFailed(t *testing.T) {
	t.Log(Err("test"))
}

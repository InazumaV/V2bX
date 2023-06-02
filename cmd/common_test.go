package cmd

import "testing"

func Test_printFailed(t *testing.T) {
	t.Log(Err("test"))
}

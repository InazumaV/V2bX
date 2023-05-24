package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
)

var (
	version  = "TempVersion" //use ldflags replace
	codename = "V2bX"
	intro    = "A V2board backend based on Xray-core"
)

func showVersion() {
	fmt.Printf("%s %s (%s) \n", codename, version, intro)
	// Warning
	fmt.Println(Warn("This version need V2board version >= 1.7.0."))
	fmt.Println(Warn("This version changed config file. Please check config file before running."))
}

var command = &cobra.Command{
	Use: "V2bX",
	PreRun: func(_ *cobra.Command, _ []string) {
		showVersion()
	},
}

func Run() {
	err := command.Execute()
	if err != nil {
		log.Println("execute failed, error:", err)
	}
}

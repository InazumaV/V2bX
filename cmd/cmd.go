package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

var command = &cobra.Command{
	Use: "V2bX",
}

func Run() {
	err := command.Execute()
	if err != nil {
		log.Println("execute failed, error:", err)
	}
}

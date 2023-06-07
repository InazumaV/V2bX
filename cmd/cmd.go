package cmd

import (
	"log"

	_ "github.com/Yuzuki616/V2bX/core/imports"
	"github.com/spf13/cobra"
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

package main

import (
	"fmt"
	"os"

	"github.com/davipatricio/dnpm/cmd/commands"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dnpm",
	Short: "dnpm is a package manager for Node.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to the dnpm!")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	commands.AddInstallCommand(rootCmd)
	commands.AddUpdateCommand(rootCmd)
}

func main() {
	Execute()
}

package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	updateLatest bool
	onlyDev      bool
	onlyProd     bool
)

var updateCmd = &cobra.Command{
	Use:     "update [options] [package]",
	Short:   "Updates all dependencies for a project or a list of packages",
	Example: "  update\n  update react@latest\n  update react@18\n  update -l eslint",
	Aliases: []string{"up", "upgrade", "in"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("You must specify a package to update")
			return
		}

		fmt.Printf("Updating package: %s\n", args[0])
		fmt.Printf("Latest flag: %v\n", updateLatest)
	},
}

func init() {
	updateCmd.Flags().BoolVarP(&updateLatest, "latest", "l", false, "Update to the latest version")
	updateCmd.Flags().BoolVarP(&onlyDev, "dev", "D", false, "Update only devDependencies")
	updateCmd.Flags().BoolVarP(&onlyProd, "prod", "P", false, "Update only regular dependencies")
	updateCmd.Flags().BoolVar(&ignoreOptional, "no-optional", false, "Ignore optional dependencies")
}

func AddUpdateCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(updateCmd)
}

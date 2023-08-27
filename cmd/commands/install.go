package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:     "install [options] [package]",
	Short:   "Installs all dependencies for a project or a list of packages",
	Example: "  install\n  install react\n  install -D eslint",
	Aliases: []string{"add", "i", "in"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("You must specify a package to install")
			return
		}

		fmt.Printf("Installing package: %s\n", args[0])
		fmt.Printf("Dev flag: %v\n", saveDev)
		fmt.Printf("Prod flag: %v\n", saveProd)
	},
}

func init() {
	installCmd.Flags().BoolVarP(&saveDev, "save-dev", "D", false, "Install the specified packages as devDependencies")
	installCmd.Flags().BoolVarP(&saveProd, "save-prod", "P", false, "Install the specified packages as regular dependencies")
	installCmd.Flags().BoolVar(&ignoreOptional, "no-optional", false, "Ignore optional dependencies")
}

func AddInstallCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(installCmd)
}

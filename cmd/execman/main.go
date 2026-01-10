package main

import (
	"fmt"
	"os"

	"github.com/sfkleach/execman/pkg/check"
	"github.com/sfkleach/execman/pkg/forget"
	initpkg "github.com/sfkleach/execman/pkg/init"
	"github.com/sfkleach/execman/pkg/install"
	"github.com/sfkleach/execman/pkg/list"
	"github.com/sfkleach/execman/pkg/remove"
	"github.com/sfkleach/execman/pkg/update"
	"github.com/sfkleach/execman/pkg/version"
	"github.com/spf13/cobra"
)

var versionFlag bool

// Install command flags.
var (
	installInto               string
	installYes                bool
	installIncludePrereleases bool
)

var rootCmd = &cobra.Command{
	Use:   "execman",
	Short: "Execman - Executable manager",
	Long:  `Execman is a command-line tool for managing executables.`,
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			if err := version.ShowVersion(false, false); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		} else {
			_ = cmd.Help()
		}
	},
}

var installCmd = &cobra.Command{
	Use:   "install <github.com/owner/repo>[@version]",
	Short: "Install an executable from GitHub",
	Long:  `Install an executable from a GitHub release`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		opts := install.Options{
			Source:             args[0],
			Into:               installInto,
			Yes:                installYes,
			IncludePrereleases: installIncludePrereleases,
		}
		if err := install.Run(opts); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&versionFlag, "version", false, "Print version information")

	// Install command flags.
	installCmd.Flags().StringVarP(&installInto, "into", "d", "", "Install to specified directory")
	installCmd.Flags().BoolVarP(&installYes, "yes", "y", false, "Skip confirmation prompts")
	installCmd.Flags().BoolVar(&installIncludePrereleases, "include-prereleases", false, "Allow installing prerelease versions")

	rootCmd.AddCommand(version.NewVersionCommand())
	rootCmd.AddCommand(initpkg.NewInitCommand())
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(list.NewListCommand())
	rootCmd.AddCommand(check.NewCheckCommand())
	rootCmd.AddCommand(update.NewUpdateCommand())
	rootCmd.AddCommand(remove.NewRemoveCommand())
	rootCmd.AddCommand(forget.NewForgetCommand())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

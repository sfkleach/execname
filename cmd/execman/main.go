package main

import (
	"fmt"
	"os"

	"github.com/sfkleach/execman/pkg/check"
	"github.com/sfkleach/execman/pkg/forget"
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
			fmt.Println(version.GetVersion())
		} else {
			_ = cmd.Help()
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of execman",
	Long:  `All software has versions. This is execman's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.GetVersion())
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

var adoptCmd = &cobra.Command{
	Use:   "adopt",
	Short: "Adopt an existing executable (TBD)",
	Long:  `Adopt an existing executable (TBD)`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("adopt subcommand - TBD")
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&versionFlag, "version", false, "Print version information")

	// Install command flags.
	installCmd.Flags().StringVarP(&installInto, "into", "d", "", "Install to specified directory")
	installCmd.Flags().BoolVarP(&installYes, "yes", "y", false, "Skip confirmation prompts")
	installCmd.Flags().BoolVar(&installIncludePrereleases, "include-prereleases", false, "Allow installing prerelease versions")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(list.NewListCommand())
	rootCmd.AddCommand(check.NewCheckCommand())
	rootCmd.AddCommand(update.NewUpdateCommand())
	rootCmd.AddCommand(remove.NewRemoveCommand())
	rootCmd.AddCommand(forget.NewForgetCommand())
	rootCmd.AddCommand(adoptCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

package cmd

import (
	"fmt"
	"nt-folly-xmaxx-comp/internal/pkg/build"

	"github.com/spf13/cobra"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "prints api server build version",
	Long:  "Prints API Server build version.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Version:", build.Version)
		fmt.Println("Build Hash:", build.BuildHash)
	},
}

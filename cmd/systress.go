package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	duration int64
)

var rootCmd = &cobra.Command{
	Use:   "systress",
	Short: "System stress tool box",
	Long:  `Bunch of tool to generate load on a system in a timed box manner`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(processCmd)

	rootCmd.PersistentFlags().Int64VarP(&duration, "duration", "", 5, "specify duration(s) of the test")
}

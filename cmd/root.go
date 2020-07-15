package cmd

import (
	"github.com/spf13/cobra"
)

var rootDir string

func init() {
	RootCmd.AddCommand(InitCmd)
	RootCmd.AddCommand(RunCmd)
	RootCmd.PersistentFlags().StringVar(&rootDir, "home", "./tmhome", "Home directory of Data Blockchain")
}

var RootCmd = cobra.Command{
	Use:   "dbc-node",
	Short: "Data Blockchain node",
}

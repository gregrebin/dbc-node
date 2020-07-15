package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/config"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize config files",
	Run:   initialize,
}

func initialize(cmd *cobra.Command, args []string) {
	configuration := config.DefaultConfig()
	configuration.SetRoot(rootDir)
	configuration.ValidateBasic()
	config.EnsureRoot(configuration.RootDir)
}

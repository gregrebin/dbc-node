package cmd

import (
	"encoding/hex"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/types"
	"time"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize config files",
	Run:   initialize,
}

func initialize(cmd *cobra.Command, args []string) {
	configuration := config.DefaultConfig()
	configuration.SetRoot(rootDir)
	config.EnsureRoot(configuration.RootDir)

	configuration.LogLevel = "consensus:error,*:info"
	configuration.RPC.CORSAllowedOrigins = []string{"*"}
	configuration.P2P.AllowDuplicateIP = true
	configuration.Consensus.CreateEmptyBlocksInterval = time.Duration(10) * time.Second
	configuration.ValidateBasic()
	config.WriteConfigFile(rootDir+"/config/config.toml", configuration)

	privValKeyFile := configuration.PrivValidatorKeyFile()
	privValStateFile := configuration.PrivValidatorStateFile()
	privVal := privval.GenFilePV(privValKeyFile, privValStateFile)
	privVal.Save()

	nodeKeyFile := configuration.NodeKeyFile()
	p2p.LoadOrGenNodeKey(nodeKeyFile)

	genFile := configuration.GenesisFile()
	genDoc := types.GenesisDoc{
		ChainID:         "datablockchain",
		GenesisTime:     time.Now(),
		ConsensusParams: types.DefaultConsensusParams(),
	}
	for _, key := range genValidatorKeys {
		genDoc.Validators = append(genDoc.Validators, types.GenesisValidator{
			Address: key.Address(),
			PubKey:  key,
			Power:   genValidators[hex.EncodeToString(key[:])],
		})
	}
	genDoc.SaveAs(genFile)
}

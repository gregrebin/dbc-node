package main

import (
	"flag"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/cli/flags"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	dataBlockChain := NewDataBlockChain()

	rootDir := flag.String("home", "./tmhome", "Tendermint home testDirectory path")
	flag.Parse()

	configuration := config.DefaultConfig()
	viper.SetConfigFile(*rootDir + "/config/config.toml")
	viper.ReadInConfig()
	viper.Unmarshal(configuration)
	configuration.SetRoot(*rootDir)
	configuration.ValidateBasic()

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	logger, _ = flags.ParseLogLevel(configuration.LogLevel, logger, config.DefaultLogLevel())

	pv := privval.LoadFilePV(
		configuration.PrivValidatorKeyFile(),
		configuration.PrivValidatorStateFile(),
	)

	nodeKey, _ := p2p.LoadNodeKey(configuration.NodeKeyFile())

	node, _ := node.NewNode(
		configuration,
		pv,
		nodeKey,
		proxy.NewLocalClientCreator(dataBlockChain),
		node.DefaultGenesisDocProviderFunc(configuration),
		node.DefaultDBProvider,
		node.DefaultMetricsProvider(configuration.Instrumentation),
		logger)

	node.Start()
	defer func() {
		node.Stop()
		node.Wait()
	}()

	sign := make(chan os.Signal, 1)
	signal.Notify(sign, syscall.SIGINT, syscall.SIGTERM)
	<-sign
	os.Exit(0)
}


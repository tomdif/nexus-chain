package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/spf13/cobra"

	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	cmttypes "github.com/cometbft/cometbft/types"

	"nexus/app"
	miningcli "nexus/x/mining/client/cli"
)

func NewRootCmd() *cobra.Command {
	encodingConfig := app.MakeEncodingConfig()

	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Codec).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithHomeDir(app.DefaultNodeHome)

	rootCmd := &cobra.Command{
		Use:   "nexusd",
		Short: "NEXUS Chain",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}
			return nil
		},
	}

	rootCmd.AddCommand(
		InitCmd(),
		StartCmd(),
		TxCmd(),
		QueryCmd(),
		keys.Commands(),
		VersionCmd(),
	)
	rootCmd.PersistentFlags().String(flags.FlagHome, app.DefaultNodeHome, "home")
	return rootCmd
}

func TxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx",
		Short: "Transaction commands",
	}
	cmd.AddCommand(miningcli.GetTxCmd())
	return cmd
}

func QueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "query",
		Short:   "Query commands",
		Aliases: []string{"q"},
	}
	cmd.AddCommand(miningcli.GetQueryCmd())
	return cmd
}

func InitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "init [moniker]",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			home, _ := cmd.Flags().GetString(flags.FlagHome)
			chainID := "nexus-testnet-1"
			configDir := filepath.Join(home, "config")
			dataDir := filepath.Join(home, "data")
			os.MkdirAll(configDir, 0755)
			os.MkdirAll(dataDir, 0755)

			cmtConfig := cmtcfg.DefaultConfig()
			cmtConfig.SetRoot(home)
			cmtConfig.Moniker = args[0]
			cmtcfg.WriteConfigFile(filepath.Join(configDir, "config.toml"), cmtConfig)

			pvKeyFile := filepath.Join(configDir, "priv_validator_key.json")
			pvStateFile := filepath.Join(dataDir, "priv_validator_state.json")
			pv := privval.GenFilePV(pvKeyFile, pvStateFile)
			pv.Save()

			nodeKeyFile := filepath.Join(configDir, "node_key.json")
			p2p.LoadOrGenNodeKey(nodeKeyFile)

			pubKey, _ := pv.GetPubKey()
			pubKeyB64 := base64.StdEncoding.EncodeToString(pubKey.Bytes())

			genFile := filepath.Join(configDir, "genesis.json")
			enc := app.MakeEncodingConfig()
			appGenState := app.ModuleBasics.DefaultGenesis(enc.Codec)
			appState, _ := json.MarshalIndent(appGenState, "", "  ")

			genDoc := cmttypes.GenesisDoc{
				ChainID:         chainID,
				GenesisTime:     time.Now(),
				ConsensusParams: cmttypes.DefaultConsensusParams(),
				AppState:        appState,
				Validators: []cmttypes.GenesisValidator{{
					Address: pubKey.Address(),
					PubKey:  pubKey,
					Power:   1000,
					Name:    args[0],
				}},
			}
			genDoc.SaveAs(genFile)

			appCfg := serverconfig.DefaultConfig()
			appCfg.MinGasPrices = "0unexus"
			serverconfig.WriteConfigFile(filepath.Join(configDir, "app.toml"), appCfg)

			cmd.Printf("Initialized! Home: %s\nValidator PubKey: %s\n", home, pubKeyB64)
			return nil
		},
	}
	return cmd
}

func StartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the NEXUS node",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, _ := cmd.Flags().GetString(flags.FlagHome)
			configDir := filepath.Join(home, "config")
			dataDir := filepath.Join(home, "data")

			// Ensure directories exist
			os.MkdirAll(configDir, 0755)
			os.MkdirAll(dataDir, 0755)

			// Load CometBFT config
			cmtConfig := cmtcfg.DefaultConfig()
			cmtConfig.SetRoot(home)
			configFile := filepath.Join(configDir, "config.toml")
			if _, err := os.Stat(configFile); err == nil {
				cmtConfig = cmtcfg.LoadConfigFile(configFile)
			} else {
				// Write default config if it doesn't exist
				cmtcfg.WriteConfigFile(configFile, cmtConfig)
			}

			// Open application database
			dbPath := filepath.Join(home, "data")
			db, err := dbm.NewGoLevelDB("application", dbPath, nil)
			if err != nil {
				return err
			}
			defer db.Close()

			// Create logger
			logger := log.NewLogger(cmd.OutOrStdout())

			// Create NEXUS application
			nexusApp := app.New(logger, db, nil, true, nil)

			// Load genesis document
			genFile := filepath.Join(configDir, "genesis.json")
			genDoc, err := cmttypes.GenesisDocFromFile(genFile)
			if err != nil {
				return err
			}

			// Load validator private key
			pvKeyFile := filepath.Join(configDir, "priv_validator_key.json")
			pvStateFile := filepath.Join(dataDir, "priv_validator_state.json")
			pv := privval.LoadFilePV(pvKeyFile, pvStateFile)

			// Load node key
			nodeKeyFile := filepath.Join(configDir, "node_key.json")
			nodeKey, err := p2p.LoadNodeKey(nodeKeyFile)
			if err != nil {
				return err
			}

			// Create CometBFT node with NEXUS app as ABCI application
			cmtNode, err := node.NewNode(
				cmtConfig,
				pv,
				nodeKey,
				proxy.NewLocalClientCreator(nexusApp),
				node.DefaultGenesisDocProviderFunc(cmtConfig),
				cmtcfg.DefaultDBProvider,
				node.DefaultMetricsProvider(cmtConfig.Instrumentation),
				logger.With("module", "node"),
			)
			if err != nil {
				return err
			}

			// Start the node
			if err := cmtNode.Start(); err != nil {
				return err
			}

			cmd.Println("========================================")
			cmd.Println("  NEXUS Chain - Proof of Useful Work")
			cmd.Println("========================================")
			cmd.Printf("  Chain ID: %s\n", genDoc.ChainID)
			cmd.Printf("  Home: %s\n", home)
			cmd.Printf("  Node ID: %s\n", nodeKey.ID())
			cmd.Println("  Status: Node started, producing blocks")
			cmd.Println("========================================")
			cmd.Println("")
			cmd.Println("  Press Ctrl+C to stop")
			cmd.Println("")

			// Setup signal handling for graceful shutdown
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

			// Wait for shutdown signal
			<-sigCh

			cmd.Println("\nShutting down gracefully...")

			// Stop the node
			if err := cmtNode.Stop(); err != nil {
				logger.Error("Error stopping node", "error", err)
			}

			// Wait for node to stop
			cmtNode.Wait()

			cmd.Println("Node stopped")
			return nil
		},
	}
}

func VersionCmd() *cobra.Command {
	return &cobra.Command{
		Use: "version",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("NEXUS v0.1.0")
		},
	}
}

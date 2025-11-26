package cmd

import (
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
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
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
		Short: "Start the NEXUS node (test mode - simulated blocks)",
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
			if _, err := os.Stat(configFile); os.IsNotExist(err) {
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

			// Initialize chain with genesis
			var genesisState map[string]json.RawMessage
			if err := json.Unmarshal(genDoc.AppState, &genesisState); err != nil {
				return err
			}

			// Create initial context
			ctx := nexusApp.BaseApp.NewContext(true)
			ctx = ctx.WithBlockHeight(0).WithBlockTime(genDoc.GenesisTime)

			// Initialize chain
			_, err = nexusApp.InitChainer(ctx, &cmtproto.RequestInitChain{
				Time:            genDoc.GenesisTime,
				ChainId:         genDoc.ChainID,
				ConsensusParams: genDoc.ConsensusParams,
				Validators:      nil,
				AppStateBytes:   genDoc.AppState,
			})
			if err != nil {
				return err
			}

			// Commit genesis
			nexusApp.Commit()

			// Load node key for display
			nodeKeyFile := filepath.Join(configDir, "node_key.json")
			nodeKey, err := p2p.LoadNodeKey(nodeKeyFile)
			if err != nil {
				return err
			}

			cmd.Println("========================================")
			cmd.Println("  NEXUS Chain - Proof of Useful Work")
			cmd.Println("========================================")
			cmd.Printf("  Chain ID: %s\n", genDoc.ChainID)
			cmd.Printf("  Home: %s\n", home)
			cmd.Printf("  Node ID: %s\n", nodeKey.ID())
			cmd.Println("  Status: Test mode - simulating blocks")
			cmd.Println("  Block time: 2 seconds")
			cmd.Println("========================================")
			cmd.Println("")
			cmd.Println("  Press Ctrl+C to stop")
			cmd.Println("")

			// Setup signal handling for graceful shutdown
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

			// Block production loop (test mode)
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()

			height := int64(1)

			for {
				select {
				case <-sigCh:
					cmd.Println("\nShutting down gracefully...")
					return nil

				case blockTime := <-ticker.C:
					// Create block context
					ctx := nexusApp.BaseApp.NewContext(false)
					ctx = ctx.WithBlockHeight(height).WithBlockTime(blockTime)

					// PreBlock
					_, err := nexusApp.PreBlocker(ctx, &cmtproto.RequestFinalizeBlock{
						Height: height,
						Time:   blockTime,
					})
					if err != nil {
						logger.Error("PreBlock failed", "height", height, "error", err)
						continue
					}

					// BeginBlock
					_, err = nexusApp.BeginBlocker(ctx)
					if err != nil {
						logger.Error("BeginBlock failed", "height", height, "error", err)
						continue
					}

					// EndBlock
					_, err = nexusApp.EndBlocker(ctx)
					if err != nil {
						logger.Error("EndBlock failed", "height", height, "error", err)
						continue
					}

					// Commit
					nexusApp.Commit()

					logger.Info("Block produced",
						"height", height,
						"time", blockTime.Format(time.RFC3339),
					)

					height++
				}
			}
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

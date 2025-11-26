package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"

	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"nexus/app"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "nexusd",
		Short: "NEXUS Network - Proof of Useful Work Blockchain",
	}

	rootCmd.AddCommand(
		InitCmd(),
		StartCmd(),
		VersionCmd(),
	)

	rootCmd.PersistentFlags().String(flags.FlagHome, app.DefaultNodeHome, "home directory")
	return rootCmd
}

func InitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [moniker]",
		Short: "Initialize node configuration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			home, _ := cmd.Flags().GetString(flags.FlagHome)
			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			if chainID == "" {
				chainID = "nexus-testnet-1"
			}

			// Create directories
			configDir := filepath.Join(home, "config")
			dataDir := filepath.Join(home, "data")
			os.MkdirAll(configDir, 0755)
			os.MkdirAll(dataDir, 0755)

			// Create genesis file
			genFile := filepath.Join(configDir, "genesis.json")
			encodingConfig := app.MakeEncodingConfig()

			appGenState := app.ModuleBasics.DefaultGenesis(encodingConfig.Codec)
			appState, err := json.MarshalIndent(appGenState, "", "  ")
			if err != nil {
				return err
			}

			genDoc := map[string]interface{}{
				"chain_id":         chainID,
				"genesis_time":     "2024-01-01T00:00:00Z",
				"consensus_params": defaultConsensusParams(),
				"app_state":        json.RawMessage(appState),
			}

			genBytes, err := json.MarshalIndent(genDoc, "", "  ")
			if err != nil {
				return err
			}

			if err := os.WriteFile(genFile, genBytes, 0644); err != nil {
				return err
			}

			// Create minimal config.toml
			configFile := filepath.Join(configDir, "config.toml")
			configContent := `# NEXUS Node Configuration
moniker = "` + args[0] + `"
`
			if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
				return err
			}

			cmd.Println("Initialized NEXUS node!")
			cmd.Printf("  Moniker: %s\n", args[0])
			cmd.Printf("  Chain ID: %s\n", chainID)
			cmd.Printf("  Home: %s\n", home)
			return nil
		},
	}
	cmd.Flags().String(flags.FlagChainID, "nexus-testnet-1", "chain ID")
	return cmd
}

func StartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the node",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, _ := cmd.Flags().GetString(flags.FlagHome)
			dbPath := filepath.Join(home, "data")

			os.MkdirAll(dbPath, 0755)

			db, err := dbm.NewGoLevelDB("application", dbPath, nil)
			if err != nil {
				return err
			}
			defer db.Close()

			logger := log.NewLogger(os.Stdout)

			nexusApp := app.New(
				logger,
				db,
				nil,
				true,
				nil,
			)

			cmd.Println("========================================")
			cmd.Println("  NEXUS Chain - Proof of Useful Work")
			cmd.Println("========================================")
			cmd.Printf("  Home: %s\n", home)
			cmd.Printf("  App: %s\n", nexusApp.Name())
			cmd.Println("")
			cmd.Println("  Node is running! Press Ctrl+C to stop")
			cmd.Println("========================================")

			// Block forever (placeholder - real node would run CometBFT)
			select {}
		},
	}
	return cmd
}

func VersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("NEXUS Chain v0.1.0")
			cmd.Println("Mining Module: enabled")
			cmd.Println("Cosmos SDK: v0.50.8")
		},
	}
}

func defaultConsensusParams() map[string]interface{} {
	return map[string]interface{}{
		"block": map[string]interface{}{
			"max_bytes": "22020096",
			"max_gas":   "-1",
		},
		"evidence": map[string]interface{}{
			"max_age_num_blocks": "100000",
			"max_age_duration":   "172800000000000",
			"max_bytes":          "1048576",
		},
		"validator": map[string]interface{}{
			"pub_key_types": []string{"ed25519"},
		},
	}
}

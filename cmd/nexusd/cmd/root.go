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
	"cosmossdk.io/math"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/cobra"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometb ft/privval"
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

			// Create CometBFT config
			cmtConfig := cmtcfg.DefaultConfig()
			cmtConfig.SetRoot(home)
			cmtConfig.Moniker = args[0]
			cmtcfg.WriteConfigFile(filepath.Join(configDir, "config.toml"), cmtConfig)

			// Generate validator private key
			pvKeyFile := filepath.Join(configDir, "priv_validator_key.json")
			pvStateFile := filepath.Join(dataDir, "priv_validator_state.json")
			pv := privval.GenFilePV(pvKeyFile, pvStateFile)
			pv.Save()

			// Generate node key
			nodeKeyFile := filepath.Join(configDir, "node_key.json")
			p2p.LoadOrGenNodeKey(nodeKeyFile)

			// Get validator consensus public key
			valPubKey, _ := pv.GetPubKey()
			pubKeyB64 := base64.StdEncoding.EncodeToString(valPubKey.Bytes())

			// Create a keyring and generate a key for the validator account
			kr := keyring.NewInMemory(app.MakeEncodingConfig().Codec)
			
			// Generate a new key for the validator operator account
			mnemonic := ""
			record, mnemonic, err := kr.NewMnemonic(
				"validator",
				keyring.English,
				sdk.GetConfig().GetFullBIP44Path(),
				keyring.DefaultBIP39Passphrase,
				hd.Secp256k1,
			)
			if err != nil {
				return err
			}

			// Get the validator operator address
			valOperatorAddr, err := record.GetAddress()
			if err != nil {
				return err
			}

			// Get the validator operator public key
			valOperatorPubKey, err := record.GetPubKey()
			if err != nil {
				return err
			}

			// Convert to validator address
			valAddr := sdk.ValAddress(valOperatorAddr)

			// Create genesis with proper validator setup
			enc := app.MakeEncodingConfig()
			appGenState := app.ModuleBasics.DefaultGenesis(enc.Codec)

			// === AUTH MODULE: Add genesis account ===
			var authGenState authtypes.GenesisState
			enc.Codec.MustUnmarshalJSON(appGenState[authtypes.ModuleName], &authGenState)

			genAccount := authtypes.NewBaseAccount(valOperatorAddr, valOperatorPubKey, 0, 0)
			if err := genAccount.Validate(); err != nil {
				return err
			}

			genAccounts := []authtypes.GenesisAccount{genAccount}
			authGenState.Accounts, err = authtypes.PackAccounts(genAccounts)
			if err != nil {
				return err
			}

			appGenState[authtypes.ModuleName] = enc.Codec.MustMarshalJSON(&authGenState)

			// === BANK MODULE: Add initial balance ===
			initialBalance := sdk.NewCoins(sdk.NewInt64Coin("unexus", 1000000000000)) // 1M NEX
			selfDelegation := math.NewInt(100000000000) // 100K NEX = 100000000000 unexus
			
			var bankGenState banktypes.GenesisState
			enc.Codec.MustUnmarshalJSON(appGenState[banktypes.ModuleName], &bankGenState)

			// Add validator operator balance
			bankGenState.Balances = append(bankGenState.Balances, banktypes.Balance{
				Address: valOperatorAddr.String(),
				Coins:   initialBalance,
			})

			// Add bonded pool module account balance (required for staking)
			bondedPoolAddr := authtypes.NewModuleAddress(stakingtypes.BondedPoolName)
			bondedPoolCoins := sdk.NewCoins(sdk.NewInt64Coin("unexus", selfDelegation.Int64()))
			bankGenState.Balances = append(bankGenState.Balances, banktypes.Balance{
				Address: bondedPoolAddr.String(),
				Coins:   bondedPoolCoins,
			})

			// Update total supply
			bankGenState.Supply = bankGenState.Supply.Add(initialBalance...).Add(bondedPoolCoins...)

			appGenState[banktypes.ModuleName] = enc.Codec.MustMarshalJSON(&bankGenState)

			// === STAKING MODULE: Add validator with delegation ===
			var stakingGenState stakingtypes.GenesisState
			enc.Codec.MustUnmarshalJSON(appGenState[stakingtypes.ModuleName], &stakingGenState)

			// Set bond denom to unexus
			stakingGenState.Params.BondDenom = "unexus"

			// Create validator description
			description := stakingtypes.Description{
				Moniker:         args[0],
				Identity:        "",
				Website:         "",
				SecurityContact: "",
				Details:         "Genesis validator",
			}

			// Create commission rates
			commissionRates := stakingtypes.CommissionRates{
				Rate:          math.LegacyNewDecWithPrec(1, 1), // 10%
				MaxRate:       math.LegacyNewDecWithPrec(2, 1), // 20%
				MaxChangeRate: math.LegacyNewDecWithPrec(1, 2), // 1%
			}

			// Create commission with update time
			commission := stakingtypes.Commission{
				CommissionRates: commissionRates,
				UpdateTime:      time.Now(),
			}

			// Convert CometBFT pubkey to SDK pubkey
			sdkPubKey := &ed25519.PubKey{Key: valPubKey.Bytes()}

			// Convert consensus pubkey to Any
			pkAny, err := types.NewAnyWithValue(sdkPubKey)
			if err != nil {
				return err
			}

			// Create validator
			validator := stakingtypes.Validator{
				OperatorAddress:   valAddr.String(),
				ConsensusPubkey:   pkAny,
				Jailed:            false,
				Status:            stakingtypes.Bonded,
				Tokens:            selfDelegation,
				DelegatorShares:   math.LegacyNewDecFromInt(selfDelegation),
				Description:       description,
				UnbondingHeight:   0,
				UnbondingTime:     time.Unix(0, 0).UTC(),
				Commission:        commission,
				MinSelfDelegation: math.NewInt(1000000), // 1 NEX minimum
			}

			stakingGenState.Validators = append(stakingGenState.Validators, validator)

			// Create delegation
			delegation := stakingtypes.Delegation{
				DelegatorAddress: valOperatorAddr.String(),
				ValidatorAddress: valAddr.String(),
				Shares:           math.LegacyNewDecFromInt(selfDelegation),
			}

			stakingGenState.Delegations = append(stakingGenState.Delegations, delegation)

			appGenState[stakingtypes.ModuleName] = enc.Codec.MustMarshalJSON(&stakingGenState)

			// Marshal final app state
			appState, err := json.MarshalIndent(appGenState, "", "  ")
			if err != nil {
				return err
			}

			// Create genesis document
			genFile := filepath.Join(configDir, "genesis.json")
			genDoc := cmttypes.GenesisDoc{
				ChainID:         chainID,
				GenesisTime:     time.Now(),
				ConsensusParams: cmttypes.DefaultConsensusParams(),
				AppState:        appState,
				Validators: []cmttypes.GenesisValidator{{
					Address: valPubKey.Address(),
					PubKey:  valPubKey,
					Power:   selfDelegation.Quo(math.NewInt(1000000)).Int64(), // Convert to consensus power
					Name:    args[0],
				}},
			}

			if err := genDoc.SaveAs(genFile); err != nil {
				return err
			}

			// Create app config
			appCfg := serverconfig.DefaultConfig()
			appCfg.MinGasPrices = "0unexus"
			serverconfig.WriteConfigFile(filepath.Join(configDir, "app.toml"), appCfg)

			cmd.Println("========================================")
			cmd.Println("  NEXUS Chain Initialized")
			cmd.Println("========================================")
			cmd.Printf("  Home: %s\n", home)
			cmd.Printf("  Chain ID: %s\n", chainID)
			cmd.Printf("  Moniker: %s\n", args[0])
			cmd.Println("========================================")
			cmd.Printf("  Validator Operator Address: %s\n", valOperatorAddr.String())
			cmd.Printf("  Validator Address: %s\n", valAddr.String())
			cmd.Printf("  Validator PubKey: %s\n", pubKeyB64)
			cmd.Printf("  Initial Balance: %s\n", initialBalance.String())
			cmd.Printf("  Self-Delegation: %s\n", sdk.NewCoin("unexus", selfDelegation).String())
			cmd.Println("========================================")
			cmd.Println("  Mnemonic (SAVE THIS!):")
			cmd.Printf("  %s\n", mnemonic)
			cmd.Println("========================================")

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

			// Load genesis document first to get chain ID
			genFile := filepath.Join(configDir, "genesis.json")
			genDoc, err := cmttypes.GenesisDocFromFile(genFile)
			if err != nil {
				return err
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

			// Create NEXUS application with chain ID
			nexusApp := app.New(logger, db, nil, true, nil, genDoc.ChainID)

			// Convert ConsensusParams to proto type
			consensusParamsProto := genDoc.ConsensusParams.ToProto()

			// Initialize chain with genesis using InitChain
			_, err = nexusApp.InitChain(&abcitypes.RequestInitChain{
				Time:            genDoc.GenesisTime,
				ChainId:         genDoc.ChainID,
				ConsensusParams: &consensusParamsProto,
				Validators:      nil,
				AppStateBytes:   genDoc.AppState,
			})
			if err != nil {
				return err
			}

			// DO NOT call Commit() here - the first commit happens after first FinalizeBlock

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
					// Create FinalizeBlock request
					req := &abcitypes.RequestFinalizeBlock{
						Height: height,
						Time:   blockTime,
						Hash:   []byte{}, // Empty hash for test mode
					}

					// Process the block through FinalizeBlock
					_, err := nexusApp.FinalizeBlock(req)
					if err != nil {
						logger.Error("FinalizeBlock failed", "height", height, "error", err)
						continue
					}

					// Commit the block
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

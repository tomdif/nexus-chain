package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
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
	"github.com/spf13/viper"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtcfg "github.com/cometbft/cometbft/config"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	cmttypes "github.com/cometbft/cometbft/types"

	"nexus/app"
	miningcli "nexus/x/mining/client/cli"
        genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
        genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
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
                GenesisCmd(),
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
        encodingConfig := app.MakeEncodingConfig()
	cmd := &cobra.Command{
		Use:   "query",
                PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
                        cmd.SetOut(cmd.OutOrStdout())
                        cmd.SetErr(cmd.ErrOrStderr())
                        clientCtx := client.Context{}.WithCodec(encodingConfig.Codec).WithInterfaceRegistry(encodingConfig.InterfaceRegistry).WithTxConfig(encodingConfig.TxConfig).WithLegacyAmino(encodingConfig.Amino).WithInput(os.Stdin)
                        if err := client.SetCmdClientContext(cmd, clientCtx); err != nil {
                                return err
                        }
                        return nil
                },
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
			chainID, _ := cmd.Flags().GetString("chain-id")
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
        cmd.Flags().String("chain-id", "nexus-testnet-1", "Chain ID for the network")
	return cmd
}

func StartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the NEXUS node with CometBFT consensus",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, _ := cmd.Flags().GetString(flags.FlagHome)
			configDir := filepath.Join(home, "config")
			dataDir := filepath.Join(home, "data")

			// Ensure directories exist
			os.MkdirAll(configDir, 0755)
			os.MkdirAll(dataDir, 0755)

			// Load CometBFT config from file (respects custom ports)
			configFile := filepath.Join(configDir, "config.toml")
			cmtConfig := cmtcfg.DefaultConfig()
			cmtConfig.SetRoot(home)
			
			// Use viper to read and unmarshal the config file
			viper.SetConfigFile(configFile)
			if err := viper.ReadInConfig(); err == nil {
				if err := viper.Unmarshal(cmtConfig); err != nil {
					return fmt.Errorf("failed to parse config: %w", err)
				}
			}

			// Load genesis document to get chain ID
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

			// Load validator private key
			pvKeyFile := cmtConfig.PrivValidatorKeyFile()
			pvStateFile := cmtConfig.PrivValidatorStateFile()
			pv := privval.LoadFilePV(pvKeyFile, pvStateFile)

			// Load node key
			nodeKeyFile := cmtConfig.NodeKeyFile()
			nodeKey, err := p2p.LoadNodeKey(nodeKeyFile)
			if err != nil {
				return err
			}

			// Create CometBFT logger adapter
			cmtLogger := newCometLogger(logger)

			// Wrap the app to implement CometB FT's ABCI interface
			abciApp := &ABCIWrapper{App: nexusApp}

			// Create local client creator that wraps our ABCI app
			clientCreator := proxy.NewLocalClientCreator(abciApp)

			// Create CometBFT node
			cmtNode, err := node.NewNode(
				cmtConfig,
				pv,
				nodeKey,
				clientCreator,
				node.DefaultGenesisDocProviderFunc(cmtConfig),
				cmtcfg.DefaultDBProvider,
				node.DefaultMetricsProvider(cmtConfig.Instrumentation),
				cmtLogger,
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
			cmd.Println("  Status: Running with CometBFT consensus")
			cmd.Println("========================================")
			cmd.Println("")
			cmd.Println("  Press Ctrl+C to stop")
			cmd.Println("")

			// Setup signal handling for graceful shutdown
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

			// Wait for interrupt signal
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

func GenesisCmd() *cobra.Command {
        txConfig := app.MakeEncodingConfig().TxConfig
        cmd := &cobra.Command{
                Use:   "genesis",
                Short: "Genesis utilities",
        }
        cmd.AddCommand(
                genutilcli.AddGenesisAccountCmd(
                        app.DefaultNodeHome,
                        txConfig.SigningContext().AddressCodec(),
                ),
                genutilcli.GenTxCmd(
                        app.ModuleBasics,
                        txConfig,
                        banktypes.GenesisBalancesIterator{},
                        app.DefaultNodeHome,
                        txConfig.SigningContext().ValidatorAddressCodec(),
                ),
                genutilcli.CollectGenTxsCmd(
                        banktypes.GenesisBalancesIterator{},
                        app.DefaultNodeHome,
                        genutiltypes.DefaultMessageValidator,
                        txConfig.SigningContext().ValidatorAddressCodec(),
                ),
                genutilcli.ValidateGenesisCmd(app.ModuleBasics),
        )
        return cmd
}

// cometLogger adapts cosmossdk log.Logger to CometBFT's logger interface
type cometLogger struct {
	logger log.Logger
}

func newCometLogger(logger log.Logger) cmtlog.Logger {
	return &cometLogger{logger: logger}
}

func (l *cometLogger) Debug(msg string, keyvals ...interface{}) {
	l.logger.Debug(msg, keyvals...)
}

func (l *cometLogger) Info(msg string, keyvals ...interface{}) {
	l.logger.Info(msg, keyvals...)
}

func (l *cometLogger) Error(msg string, keyvals ...interface{}) {
	l.logger.Error(msg, keyvals...)
}

func (l *cometLogger) With(keyvals ...interface{}) cmtlog.Logger {
	return &cometLogger{logger: l.logger.With(keyvals...)}
}

// ABCIWrapper wraps the Cosmos SDK app to implement CometBFT's ABCI interface
// This is needed because Cosmos SDK v0.50 uses different method signatures
// CometBFT expects context.Context as first parameter, but BaseApp methods don't take it
type ABCIWrapper struct {
	App *app.App
}

// Implement abci.Application interface by wrapping BaseApp methods

func (w *ABCIWrapper) Info(ctx context.Context, req *abci.RequestInfo) (*abci.ResponseInfo, error) {
	return w.App.Info(req)
}

func (w *ABCIWrapper) Query(ctx context.Context, req *abci.RequestQuery) (*abci.ResponseQuery, error) {
	return w.App.Query(ctx, req)
}

func (w *ABCIWrapper) CheckTx(ctx context.Context, req *abci.RequestCheckTx) (*abci.ResponseCheckTx, error) {
	return w.App.CheckTx(req)
}

func (w *ABCIWrapper) InitChain(ctx context.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	return w.App.InitChain(req)
}

func (w *ABCIWrapper) PrepareProposal(ctx context.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
	return w.App.PrepareProposal(req)
}

func (w *ABCIWrapper) ProcessProposal(ctx context.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
	return w.App.ProcessProposal(req)
}

func (w *ABCIWrapper) FinalizeBlock(ctx context.Context, req *abci.RequestFinalizeBlock) (*abci.ResponseFinalizeBlock, error) {
	return w.App.FinalizeBlock(req)
}

func (w *ABCIWrapper) ExtendVote(ctx context.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
	return w.App.ExtendVote(ctx, req)
}

func (w *ABCIWrapper) VerifyVoteExtension(ctx context.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
	return w.App.VerifyVoteExtension(req)
}

func (w *ABCIWrapper) Commit(ctx context.Context, req *abci.RequestCommit) (*abci.ResponseCommit, error) {
	return w.App.Commit()
}

func (w *ABCIWrapper) ListSnapshots(ctx context.Context, req *abci.RequestListSnapshots) (*abci.ResponseListSnapshots, error) {
	return w.App.ListSnapshots(req)
}

func (w *ABCIWrapper) OfferSnapshot(ctx context.Context, req *abci.RequestOfferSnapshot) (*abci.ResponseOfferSnapshot, error) {
	return w.App.OfferSnapshot(req)
}

func (w *ABCIWrapper) LoadSnapshotChunk(ctx context.Context, req *abci.RequestLoadSnapshotChunk) (*abci.ResponseLoadSnapshotChunk, error) {
	return w.App.LoadSnapshotChunk(req)
}

func (w *ABCIWrapper) ApplySnapshotChunk(ctx context.Context, req *abci.RequestApplySnapshotChunk) (*abci.ResponseApplySnapshotChunk, error) {
	return w.App.ApplySnapshotChunk(req)
}

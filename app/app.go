package app

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	dbm "github.com/cosmos/cosmos-db"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	miningmodule "nexus/x/mining"
	miningkeeper "nexus/x/mining/keeper"
	miningtypes "nexus/x/mining/types"
	nexusante "nexus/app/ante"
)

const (
	AccountAddressPrefix = "nexus"
	Name                 = "nexus"
	BondDenom            = "unexus"
)

var (
	DefaultNodeHome string

	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
		miningmodule.AppModuleBasic{},
	)

	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:     nil,
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		miningtypes.ModuleName:         {authtypes.Minter, authtypes.Burner},
	}
)

func init() {
	userHomeDir, _ := os.UserHomeDir()
	DefaultNodeHome = filepath.Join(userHomeDir, ".nexus")

	// Set address prefixes
	cfg := sdk.GetConfig()
	cfg.SetBech32PrefixForAccount(AccountAddressPrefix, AccountAddressPrefix+"pub")
	cfg.SetBech32PrefixForValidator(AccountAddressPrefix+"valoper", AccountAddressPrefix+"valoperpub")
	cfg.SetBech32PrefixForConsensusNode(AccountAddressPrefix+"valcons", AccountAddressPrefix+"valconspub")
	cfg.Seal()
}

var _ servertypes.Application = (*App)(nil)

type App struct {
	*baseapp.BaseApp

	cdc               codec.Codec
	legacyAmino       *codec.LegacyAmino
	txConfig          client.TxConfig
	interfaceRegistry codectypes.InterfaceRegistry

	keys map[string]*storetypes.KVStoreKey

	AccountKeeper authkeeper.AccountKeeper
	BankKeeper    bankkeeper.Keeper
	StakingKeeper *stakingkeeper.Keeper
	MiningKeeper  miningkeeper.Keeper

	ModuleManager      *module.Manager
	BasicModuleManager module.BasicManager
}

func New(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	encodingConfig := MakeEncodingConfig()

	interfaceRegistry := encodingConfig.InterfaceRegistry
	cdc := encodingConfig.Codec
	legacyAmino := encodingConfig.Amino
	txConfig := encodingConfig.TxConfig

	bApp := baseapp.NewBaseApp(Name, logger, db, txConfig.TxDecoder(), baseAppOptions...)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetTxEncoder(txConfig.TxEncoder())

	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey,
		banktypes.StoreKey,
		stakingtypes.StoreKey,
		miningtypes.StoreKey,
	)

	app := &App{
		BaseApp:           bApp,
		cdc:               cdc,
		legacyAmino:       legacyAmino,
		txConfig:          txConfig,
		interfaceRegistry: interfaceRegistry,
		keys:              keys,
	}

	app.AccountKeeper = authkeeper.NewAccountKeeper(
		cdc,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		authcodec.NewBech32Codec(AccountAddressPrefix),
		AccountAddressPrefix,
		authtypes.NewModuleAddress(authtypes.ModuleName).String(),
	)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		cdc,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		app.AccountKeeper,
		nil,
		authtypes.NewModuleAddress(authtypes.ModuleName).String(),
		logger,
	)

	app.StakingKeeper = stakingkeeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress(authtypes.ModuleName).String(),
		authcodec.NewBech32Codec(AccountAddressPrefix+"valoper"),
		authcodec.NewBech32Codec(AccountAddressPrefix+"valcons"),
	)

	app.MiningKeeper = miningkeeper.NewKeeper(
		cdc,
		keys[miningtypes.StoreKey],
		nil,
		app.StakingKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress(authtypes.ModuleName).String(),
	)

	app.ModuleManager = module.NewManager(
		auth.NewAppModule(cdc, app.AccountKeeper, nil, nil),
		bank.NewAppModule(cdc, app.BankKeeper, app.AccountKeeper, nil),
		staking.NewAppModule(cdc, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, nil),
		genutil.NewAppModule(app.AccountKeeper, app.StakingKeeper, app, txConfig),
		miningmodule.NewAppModule(cdc, app.MiningKeeper),
	)

	app.BasicModuleManager = module.NewBasicManagerFromManager(
		app.ModuleManager,
		map[string]module.AppModuleBasic{},
	)

	app.ModuleManager.SetOrderBeginBlockers(
		stakingtypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		genutiltypes.ModuleName,
		miningtypes.ModuleName,
	)

	app.ModuleManager.SetOrderEndBlockers(
		stakingtypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		genutiltypes.ModuleName,
		miningtypes.ModuleName,
	)

	genesisModuleOrder := []string{
		authtypes.ModuleName,
		banktypes.ModuleName,
		stakingtypes.ModuleName,
		genutiltypes.ModuleName,
		miningtypes.ModuleName,
	}
	app.ModuleManager.SetOrderInitGenesis(genesisModuleOrder...)
	app.ModuleManager.SetOrderExportGenesis(genesisModuleOrder...)

	app.ModuleManager.RegisterServices(app.configurator())

	app.MountKVStores(keys)

	// Set up custom AnteHandler with 50% fee burn
	anteHandler, err := nexusante.NewAnteHandler(nexusante.HandlerOptions{
		AccountKeeper:  app.AccountKeeper,
		BankKeeper:     app.BankKeeper,
		FeegrantKeeper: nil,
		SigGasConsumer: nil,
		TxFeeChecker:   nil,
	})
	if err != nil {
		panic(err)
	}
	app.SetAnteHandler(anteHandler)

	app.SetInitChainer(app.InitChainer)
	app.SetPreBlocker(app.PreBlocker)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			panic(err)
		}
	}

	return app
}

func (app *App) configurator() module.Configurator {
	return module.NewConfigurator(app.cdc, app.MsgServiceRouter(), app.GRPCQueryRouter())
}

func (app *App) Name() string                             { return app.BaseApp.Name() }
func (app *App) LegacyAmino() *codec.LegacyAmino          { return app.legacyAmino }
func (app *App) AppCodec() codec.Codec                    { return app.cdc }
func (app *App) InterfaceRegistry() codectypes.InterfaceRegistry { return app.interfaceRegistry }
func (app *App) TxConfig() client.TxConfig                { return app.txConfig }
func (app *App) GetKey(storeKey string) *storetypes.KVStoreKey { return app.keys[storeKey] }

// InitChainer handles genesis initialization
func (app *App) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState map[string]json.RawMessage
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	return app.ModuleManager.InitGenesis(ctx, app.cdc, genesisState)
}

// PreBlocker runs before BeginBlock
func (app *App) PreBlocker(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.ModuleManager.PreBlock(ctx)
}

// BeginBlocker runs at the start of each block
func (app *App) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.ModuleManager.BeginBlock(ctx)
}

// EndBlocker runs at the end of each block
func (app *App) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.ModuleManager.EndBlock(ctx)
}

// RegisterAPIRoutes registers REST API routes
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	// Register API routes here
}

// RegisterTxService registers gRPC tx service
func (app *App) RegisterTxService(clientCtx client.Context) {
	// Register tx service
}

// RegisterTendermintService registers Tendermint service
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	// Register tendermint service
}

// RegisterNodeService registers node service
func (app *App) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	// Register node service
}

// DefaultGenesis returns default genesis state
func (app *App) DefaultGenesis() map[string]json.RawMessage {
	return app.BasicModuleManager.DefaultGenesis(app.cdc)
}

type EncodingConfig struct {
	InterfaceRegistry codectypes.InterfaceRegistry
	Codec             codec.Codec
	TxConfig          client.TxConfig
	Amino             *codec.LegacyAmino
}

func MakeEncodingConfig() EncodingConfig {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	amino := codec.NewLegacyAmino()

	std.RegisterLegacyAminoCodec(amino)
	std.RegisterInterfaces(interfaceRegistry)
	ModuleBasics.RegisterLegacyAminoCodec(amino)
	ModuleBasics.RegisterInterfaces(interfaceRegistry)

	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             cdc,
		TxConfig:          txConfig,
		Amino:             amino,
	}
}

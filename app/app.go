package app

import (
	"io"
	"os"
	"path/filepath"

	dbm "github.com/cosmos/cosmos-db"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
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
)

const (
	AccountAddressPrefix = "nexus"
	Name                 = "nexus"
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
}

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

	ModuleManager *module.Manager
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
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
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
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
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

	app.MountKVStores(keys)

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			panic(err)
		}
	}

	return app
}

func (app *App) Name() string                                      { return app.BaseApp.Name() }
func (app *App) LegacyAmino() *codec.LegacyAmino                   { return app.legacyAmino }
func (app *App) AppCodec() codec.Codec                             { return app.cdc }
func (app *App) InterfaceRegistry() codectypes.InterfaceRegistry   { return app.interfaceRegistry }
func (app *App) TxConfig() client.TxConfig                         { return app.txConfig }

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

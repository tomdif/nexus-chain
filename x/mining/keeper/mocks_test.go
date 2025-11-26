package keeper_test

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// MockBankKeeper implements types.BankKeeper for testing
type MockBankKeeper struct {
	Balances       map[string]sdk.Coins
	ModuleBalances map[string]sdk.Coins
	SendErrors     map[string]error
}

func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{
		Balances:       make(map[string]sdk.Coins),
		ModuleBalances: make(map[string]sdk.Coins),
		SendErrors:     make(map[string]error),
	}
}

func (m *MockBankKeeper) SetBalance(addr sdk.AccAddress, coins sdk.Coins) {
	m.Balances[addr.String()] = coins
}

func (m *MockBankKeeper) SetModuleBalance(moduleName string, coins sdk.Coins) {
	m.ModuleBalances[moduleName] = coins
}

func (m *MockBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	if err, ok := m.SendErrors["SendCoinsFromAccountToModule"]; ok {
		return err
	}
	senderKey := senderAddr.String()
	senderBalance := m.Balances[senderKey]
	if !senderBalance.IsAllGTE(amt) {
		return fmt.Errorf("insufficient funds: %s < %s", senderBalance, amt)
	}
	m.Balances[senderKey] = senderBalance.Sub(amt...)
	moduleBalance := m.ModuleBalances[recipientModule]
	m.ModuleBalances[recipientModule] = moduleBalance.Add(amt...)
	return nil
}

func (m *MockBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	if err, ok := m.SendErrors["SendCoinsFromModuleToAccount"]; ok {
		return err
	}
	recipientKey := recipientAddr.String()
	moduleBalance := m.ModuleBalances[senderModule]
	if !moduleBalance.IsAllGTE(amt) {
		return fmt.Errorf("insufficient module funds: %s < %s", moduleBalance, amt)
	}
	m.ModuleBalances[senderModule] = moduleBalance.Sub(amt...)
	recipientBalance := m.Balances[recipientKey]
	m.Balances[recipientKey] = recipientBalance.Add(amt...)
	return nil
}

func (m *MockBankKeeper) MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	if err, ok := m.SendErrors["MintCoins"]; ok {
		return err
	}
	moduleBalance := m.ModuleBalances[moduleName]
	m.ModuleBalances[moduleName] = moduleBalance.Add(amt...)
	return nil
}

func (m *MockBankKeeper) BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	if err, ok := m.SendErrors["BurnCoins"]; ok {
		return err
	}
	moduleBalance := m.ModuleBalances[moduleName]
	if !moduleBalance.IsAllGTE(amt) {
		return fmt.Errorf("insufficient module funds to burn: %s < %s", moduleBalance, amt)
	}
	m.ModuleBalances[moduleName] = moduleBalance.Sub(amt...)
	return nil
}

func (m *MockBankKeeper) GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	balance := m.Balances[addr.String()]
	return sdk.NewCoin(denom, balance.AmountOf(denom))
}

// MockStakingKeeper implements types.StakingKeeper for testing
type MockStakingKeeper struct {
	Validators  []MockValidator
	TotalBonded int64
}

type MockValidator struct {
	OperatorAddr string
	BondedTokens int64
}

func NewMockStakingKeeper() *MockStakingKeeper {
	return &MockStakingKeeper{
		Validators:  make([]MockValidator, 0),
		TotalBonded: 0,
	}
}

func (m *MockStakingKeeper) AddValidator(operatorAddr string, bondedTokens int64) {
	m.Validators = append(m.Validators, MockValidator{
		OperatorAddr: operatorAddr,
		BondedTokens: bondedTokens,
	})
	m.TotalBonded += bondedTokens
}

func (m *MockStakingKeeper) GetValidator(ctx context.Context, addr sdk.ValAddress) (stakingtypes.Validator, error) {
	return stakingtypes.Validator{}, nil
}

func (m *MockStakingKeeper) TotalBondedTokens(ctx context.Context) (math.Int, error) {
	return math.NewInt(m.TotalBonded), nil
}

func (m *MockStakingKeeper) IterateBondedValidatorsByPower(ctx context.Context, fn func(index int64, validator stakingtypes.ValidatorI) (stop bool)) error {
	for i, v := range m.Validators {
		mockVal := &mockValidatorI{
			operatorAddr: v.OperatorAddr,
			bondedTokens: v.BondedTokens,
		}
		if fn(int64(i), mockVal) {
			break
		}
	}
	return nil
}

func (m *MockStakingKeeper) GetValidatorByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (stakingtypes.Validator, error) {
	return stakingtypes.Validator{}, nil
}

// mockValidatorI implements stakingtypes.ValidatorI
type mockValidatorI struct {
	operatorAddr string
	bondedTokens int64
}

func (v *mockValidatorI) IsJailed() bool                             { return false }
func (v *mockValidatorI) GetMoniker() string                         { return "" }
func (v *mockValidatorI) GetStatus() stakingtypes.BondStatus         { return stakingtypes.Bonded }
func (v *mockValidatorI) IsBonded() bool                             { return true }
func (v *mockValidatorI) IsUnbonded() bool                           { return false }
func (v *mockValidatorI) IsUnbonding() bool                          { return false }
func (v *mockValidatorI) GetOperator() string                        { return v.operatorAddr }
func (v *mockValidatorI) ConsPubKey() (cryptotypes.PubKey, error)    { return nil, nil }
func (v *mockValidatorI) TmConsPublicKey() (cmtprotocrypto.PublicKey, error) {
	return cmtprotocrypto.PublicKey{}, nil
}
func (v *mockValidatorI) GetConsAddr() ([]byte, error)               { return nil, nil }
func (v *mockValidatorI) GetTokens() math.Int                        { return math.NewInt(v.bondedTokens) }
func (v *mockValidatorI) GetBondedTokens() math.Int                  { return math.NewInt(v.bondedTokens) }
func (v *mockValidatorI) GetConsensusPower(r math.Int) int64         { return v.bondedTokens }
func (v *mockValidatorI) GetCommission() math.LegacyDec              { return math.LegacyZeroDec() }
func (v *mockValidatorI) GetMinSelfDelegation() math.Int             { return math.ZeroInt() }
func (v *mockValidatorI) GetDelegatorShares() math.LegacyDec         { return math.LegacyZeroDec() }
func (v *mockValidatorI) TokensFromShares(shares math.LegacyDec) math.LegacyDec {
	return math.LegacyZeroDec()
}
func (v *mockValidatorI) TokensFromSharesTruncated(shares math.LegacyDec) math.LegacyDec {
	return math.LegacyZeroDec()
}
func (v *mockValidatorI) TokensFromSharesRoundUp(shares math.LegacyDec) math.LegacyDec {
	return math.LegacyZeroDec()
}
func (v *mockValidatorI) SharesFromTokens(amt math.Int) (math.LegacyDec, error) {
	return math.LegacyZeroDec(), nil
}
func (v *mockValidatorI) SharesFromTokensTruncated(amt math.Int) (math.LegacyDec, error) {
	return math.LegacyZeroDec(), nil
}

package keeper_test

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MockBankKeeper implements types.BankKeeper for testing
type MockBankKeeper struct {
	Balances       map[string]sdk.Coins // address -> coins
	ModuleBalances map[string]sdk.Coins // module name -> coins
	SendErrors     map[string]error     // optional errors to simulate failures
}

func NewMockBankKeeper() *MockBankKeeper {
	return &MockBankKeeper{
		Balances:       make(map[string]sdk.Coins),
		ModuleBalances: make(map[string]sdk.Coins),
		SendErrors:     make(map[string]error),
	}
}

// SetBalance sets an account's balance for testing
func (m *MockBankKeeper) SetBalance(addr sdk.AccAddress, coins sdk.Coins) {
	m.Balances[addr.String()] = coins
}

// SetModuleBalance sets a module's balance for testing
func (m *MockBankKeeper) SetModuleBalance(moduleName string, coins sdk.Coins) {
	m.ModuleBalances[moduleName] = coins
}

func (m *MockBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	if err, ok := m.SendErrors["SendCoinsFromAccountToModule"]; ok {
		return err
	}

	senderKey := senderAddr.String()
	
	// Check sender has enough
	senderBalance := m.Balances[senderKey]
	if !senderBalance.IsAllGTE(amt) {
		return fmt.Errorf("insufficient funds: %s < %s", senderBalance, amt)
	}

	// Deduct from sender
	m.Balances[senderKey] = senderBalance.Sub(amt...)

	// Add to module
	moduleBalance := m.ModuleBalances[recipientModule]
	m.ModuleBalances[recipientModule] = moduleBalance.Add(amt...)

	return nil
}

func (m *MockBankKeeper) SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	if err, ok := m.SendErrors["SendCoinsFromModuleToAccount"]; ok {
		return err
	}

	recipientKey := recipientAddr.String()

	// Check module has enough
	moduleBalance := m.ModuleBalances[senderModule]
	if !moduleBalance.IsAllGTE(amt) {
		return fmt.Errorf("insufficient module funds: %s < %s", moduleBalance, amt)
	}

	// Deduct from module
	m.ModuleBalances[senderModule] = moduleBalance.Sub(amt...)

	// Add to recipient
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

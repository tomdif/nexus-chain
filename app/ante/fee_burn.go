package ante

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

// BurnFeeDecorator burns 50% of transaction fees for deflationary tokenomics
type BurnFeeDecorator struct {
	bankKeeper bankkeeper.Keeper
	burnRate   math.LegacyDec
}

// NewBurnFeeDecorator creates a new BurnFeeDecorator
// burnRate should be between 0 and 1 (e.g., 0.5 for 50%)
func NewBurnFeeDecorator(bk bankkeeper.Keeper, burnRate math.LegacyDec) BurnFeeDecorator {
	return BurnFeeDecorator{
		bankKeeper: bk,
		burnRate:   burnRate,
	}
}

// AnteHandle implements the AnteDecorator interface
func (bfd BurnFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// Skip fee burning during simulation
	if simulate {
		return next(ctx, tx, simulate)
	}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.ErrTxDecode.Wrap("tx must be a FeeTx")
	}

	fees := feeTx.GetFee()
	if fees.IsZero() {
		return next(ctx, tx, simulate)
	}

	// Calculate burn amount (50% of fees)
	var burnCoins sdk.Coins
	for _, fee := range fees {
		burnAmount := math.LegacyNewDecFromInt(fee.Amount).Mul(bfd.burnRate).TruncateInt()
		if burnAmount.IsPositive() {
			burnCoins = burnCoins.Add(sdk.NewCoin(fee.Denom, burnAmount))
		}
	}

	// Burn coins from the fee collector
	if !burnCoins.IsZero() {
		err := bfd.bankKeeper.BurnCoins(ctx, authtypes.FeeCollectorName, burnCoins)
		if err != nil {
			// Log but don't fail - fee collector might not have coins yet
			ctx.Logger().Error("Failed to burn fee coins", "error", err, "coins", burnCoins)
		} else {
			ctx.Logger().Info("Burned transaction fees",
				"burned", burnCoins.String(),
				"total_fees", fees.String(),
			)

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"fee_burn",
					sdk.NewAttribute("amount", burnCoins.String()),
					sdk.NewAttribute("total_fees", fees.String()),
				),
			)
		}
	}

	return next(ctx, tx, simulate)
}

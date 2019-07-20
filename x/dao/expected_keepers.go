package dao

import sdk "github.com/cosmos/cosmos-sdk/types"

// expected bank keeper
type BankKeeper interface {
	GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins

	// TODO remove once governance doesn't require use of accounts
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) (sdk.Tags, sdk.Error)
}

type StakeKeeper interface {
	// TODO: TBD
	// Rebalancing
	// get total supply of index denom , etc
}
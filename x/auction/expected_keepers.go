package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type bankKeeper interface {
	SubtractCoins(sdk.Context, sdk.AccAddress, sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
	AddCoins(sdk.Context, sdk.AccAddress, sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
	HasCoins(sdk.Context, sdk.AccAddress, sdk.Coins) bool
}

type stakingKeeper interface {
	CollectFeePoolsForAuction(ctx sdk.Context) sdk.Coins
	RepatriateFeeEarnings(ctx sdk.Context, origCoin sdk.Coin, payCoin sdk.Coin)
	RollOverFeesFromAuction(ctx sdk.Context, coin sdk.Coin)
}

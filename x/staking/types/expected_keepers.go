package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// expected coin keeper
type DistributionKeeper interface {
	GetFeePoolCommunityCoins(ctx sdk.Context) sdk.DecCoins
}

// expected fee collection keeper
type FeeCollectionKeeper interface {
	GetCollectedFees(ctx sdk.Context) sdk.Coins
}

// expected bank keeper
type BankKeeper interface {
	DelegateCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Tags, sdk.Error)
	UndelegateCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Tags, sdk.Error)
	SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
	AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
	GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	HasCoins(ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) bool
}

// expected crisis keeper
type CrisisKeeper interface {
	RegisterRoute(moduleName, route string, invar sdk.Invariant)
}

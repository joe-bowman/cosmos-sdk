package auction

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker runs at the end of every block.
func EndBlocker(ctx sdk.Context, k Keeper) sdk.Tags {

	if ctx.BlockHeight()%FeeAuctionFrequency == 0 {
		// trigger auction starts for each fee denom in collective validatorsets.
		feePools := k.stakingKeeper.CollectFeePoolsForAuction(ctx)

		for _, coin := range feePools {
			k.StartForwardAuction(ctx, coin, sdk.Coin{"uatom", sdk.ZeroInt()}) // todo: fetch bond denom
		}
	}

	expiredAuctions := k.ExpireAuctions(ctx, ctx.BlockHeight())
	// loop through expired and close them - distribute funds, delete from store (and queue)
	for _, auctionRef := range expiredAuctions {
		fmt.Printf("Closing auction %d", auctionRef.Id)
		err := k.CloseAuction(ctx, auctionRef.Id)
		if err != nil {
			panic(err) // TODO how should errors be handled here?
		}
	}

	return sdk.Tags{}
}

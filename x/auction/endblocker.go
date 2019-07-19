package auction

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlocker runs at the end of every block.
func EndBlocker(ctx sdk.Context, k Keeper) sdk.Tags {

	if ctx.BlockHeight()%FeeAuctionFrequency == 0 {
		fmt.Println("Triggering auction, fuckers!")
		// trigger auction starts for each fee denom in collective validatorsets.
	}

	// get an iterator of expired auctions
	expiredAuctions := k.getQueueIterator(ctx, ctx.BlockHeight())
	defer expiredAuctions.Close()

	// loop through and close them - distribute funds, delete from store (and queue)
	for ; expiredAuctions.Valid(); expiredAuctions.Next() {
		var auctionID ID
		k.cdc.MustUnmarshalBinaryLengthPrefixed(expiredAuctions.Value(), &auctionID)

		err := k.CloseAuction(ctx, auctionID)
		if err != nil {
			panic(err) // TODO how should errors be handled here?
		}
	}

	return sdk.Tags{}
}

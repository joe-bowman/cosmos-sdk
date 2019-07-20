package auction

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

const (
	// StoreKey is the string key for the params store
	StoreKey = "auction"
)

type Keeper struct {
	bankKeeper    bankKeeper
	stakingKeeper stakingKeeper
	// The reference to the Param Keeper to get and set Global Params
	paramsKeeper params.Keeper
	// The reference to the Paramstore to get and set gov specific params
	paramSpace params.Subspace
	storeKey   sdk.StoreKey
	cdc        *codec.Codec
	// TODO codespace
}

type AuctionRef struct {
	Id      ID
	EndTime int64
}

type AuctionRefs struct {
	Refs []AuctionRef
}

// NewKeeper returns a new auction keeper.
func NewKeeper(cdc *codec.Codec, bankKeeper bankKeeper, stakingKeeper stakingKeeper,
	paramsKeeper params.Keeper, storeKey sdk.StoreKey) Keeper {
	return Keeper{
		bankKeeper:    bankKeeper,
		stakingKeeper: stakingKeeper,
		paramsKeeper:  paramsKeeper,
		storeKey:      storeKey,
		cdc:           cdc,
	}
}

// TODO these 3 start functions be combined or abstracted away?

// StartForwardAuction starts a normal auction. Known as flap in maker.
func (k Keeper) StartForwardAuction(ctx sdk.Context, lot sdk.Coin, initialBid sdk.Coin) (ID, sdk.Error) {
	// create auction
	fmt.Printf("STARTING AUCTION FOR %s\n", lot.Denom)
	auction := NewForwardAuction(lot, initialBid, ctx.BlockHeight()+MaxAuctionDuration)
	// start the auction
	auctionID, err := k.startAuction(ctx, &auction)
	if err != nil {
		return 0, err
	}
	return auctionID, nil
}

func (k Keeper) startAuction(ctx sdk.Context, auction Auction) (ID, sdk.Error) {
	// get ID
	newAuctionID, err := k.getNextAuctionID(ctx)
	if err != nil {
		return 0, err
	}
	// set ID
	auction.SetID(newAuctionID)

	// store auction
	k.setAuction(ctx, auction)
	k.incrementNextAuctionID(ctx)
	return newAuctionID, nil
}

// PlaceBid places a bid on any auction.
func (k Keeper) PlaceBid(ctx sdk.Context, auctionID ID, bidder sdk.AccAddress, bid sdk.Coin) sdk.Error {

	// get auction from store
	auction, found := k.GetAuction(ctx, auctionID)
	if !found {
		return sdk.ErrInternal("Auction doesn't exist")
	}

	if !k.bankKeeper.HasCoins(ctx, bidder, sdk.NewCoins(bid)) {
		return sdk.ErrInternal("Insufficient funds")
	}

	// place bid
	err := auction.PlaceBid(ctx.BlockHeight(), bidder, bid) // update auction according to what type of auction it is // TODO should this return updated Auction to be more immutable?
	if err != nil {
		return err
	}

	// subtract coins from new bid
	_, _, err = k.bankKeeper.SubtractCoins(ctx, bidder, sdk.NewCoins(bid)) // TODO handle errors properly here. All coin transfers should be atomic. InputOutputCoins may work
	if err != nil {
		panic(err)
	}

	if auction.GetBid().Amount.GT(sdk.ZeroInt()) { //we aren't the first bid.
		// refund tokens to previous bidder
		_, _, err = k.bankKeeper.AddCoins(ctx, auction.GetBidder(), sdk.NewCoins(auction.GetBid())) // TODO errors
		if err != nil {
			panic(err)
		}
	}

	// update bid and bidder
	auction.SetBid(bid)
	auction.SetBidder(bidder)
	// store updated auction
	k.setAuction(ctx, auction)

	return nil
}

// CloseAuction closes an auction and distributes funds to the seller and highest bidder.
// TODO because this is called by the end blocker, it has to be valid for the duration of the EndTime block. Should maybe move this to a begin blocker?
func (k Keeper) CloseAuction(ctx sdk.Context, auctionID ID) sdk.Error {

	// get the auction from the store
	auction, found := k.GetAuction(ctx, auctionID)
	if !found {
		return sdk.ErrInternal("auction doesn't exist")
	}
	// error if auction has not reached the end time
	if ctx.BlockHeight() < auction.GetEndTime() { // auctions close at the end of the block with blockheight == EndTime
		return sdk.ErrInternal(fmt.Sprintf("auction can't be closed as curent block height (%v) is under auction end time (%v)", ctx.BlockHeight(), auction.GetEndTime()))
	}

	if auction.GetBid().Denom == "uatom" && auction.GetBid().Amount.GT(sdk.ZeroInt()) {
		k.stakingKeeper.RepatriateFeeEarnings(ctx, auction.GetLot(), auction.GetBid())
	} else {
		k.stakingKeeper.RollOverFeesFromAuction(ctx, auction.GetLot())
	}
	// take the amount from auction.Bid and pass on to validator pools.

	// delete auction from store (and queue)
	k.deleteAuction(ctx, auctionID)

	return nil
}

// ---------- Store methods ----------
// Use these to add and remove auction from the store.

// getNextAuctionID gets the next available global AuctionID
func (k Keeper) getNextAuctionID(ctx sdk.Context) (ID, sdk.Error) { // TODO don't need error return here
	// get next ID from store

	store := ctx.KVStore(k.storeKey)
	bz := store.Get(k.getNextAuctionIDKey())
	if bz == nil {
		// if not found, set the id at 0
		bz = k.cdc.MustMarshalBinaryLengthPrefixed(ID(0))
		store.Set(k.getNextAuctionIDKey(), bz)
	}
	var auctionID ID
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &auctionID)
	return auctionID, nil
}

// incrementNextAuctionID increments the global ID in the store by 1
func (k Keeper) incrementNextAuctionID(ctx sdk.Context) sdk.Error {
	// get next ID from store
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(k.getNextAuctionIDKey())
	if bz == nil {
		panic("initial auctionID never set in genesis")
		//return 0, ErrInvalidGenesis(keeper.codespace, "InitialProposalID never set") // TODO is this needed? Why not just set it zero here?
	}
	var auctionID ID
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &auctionID)
	// increment the stored next ID
	bz = k.cdc.MustMarshalBinaryLengthPrefixed(auctionID + 1)
	store.Set(k.getNextAuctionIDKey(), bz)

	return nil
}

// setAuction puts the auction into the database and adds it to the queue
// it overwrites any pre-existing auction with same ID
func (k Keeper) setAuction(ctx sdk.Context, auction Auction) {
	// remove the auction from the queue if it is already in there
	existingAuction, found := k.GetAuction(ctx, auction.GetID())
	if found {
		k.RemoveFromList(ctx, existingAuction.GetEndTime(), existingAuction.GetID())
	}

	// store auction
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(auction)
	store.Set(k.getAuctionKey(auction.GetID()), bz)

	// add to the queue
	k.AddToList(ctx, auction.GetEndTime(), auction.GetID())
}

func (k Keeper) AddToList(ctx sdk.Context, endTime int64, id ID) {
	var refs AuctionRefs

	store := ctx.KVStore(k.storeKey)

	bz := store.Get(k.getLiveAuctionListKey())
	if len(bz) > 0 {
		k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &refs)
	}

	newRef := AuctionRef{id, endTime}
	refs.Refs = append(refs.Refs, newRef)
	fmt.Printf("REFS: %v", refs)
	bz = k.cdc.MustMarshalBinaryLengthPrefixed(refs)
	store.Set(k.getLiveAuctionListKey(), bz)
	return
}

func (k Keeper) RemoveFromList(ctx sdk.Context, endTime int64, id ID) {
	var refs AuctionRefs
	var newRefs []AuctionRef

	store := ctx.KVStore(k.storeKey)

	bz := store.Get(k.getLiveAuctionListKey())
	if len(bz) > 0 {
		k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &refs)
	}

	for _, ref := range refs.Refs {
		if ref.Id != id {
			newRefs = append(newRefs, ref)
		}
		refs.Refs = newRefs
		bz = k.cdc.MustMarshalBinaryLengthPrefixed(newRefs)
		store.Set(k.getLiveAuctionListKey(), bz)
	}
	return
}

func (k Keeper) ExpireAuctions(ctx sdk.Context, endTime int64) (result []AuctionRef) {
	var refs AuctionRefs
	var newRefs []AuctionRef

	store := ctx.KVStore(k.storeKey)

	bz := store.Get(k.getLiveAuctionListKey())
	if len(bz) > 0 {
		k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &refs)
	}

	for _, ref := range refs.Refs {
		if ref.EndTime > endTime {
			newRefs = append(newRefs, ref)
		} else {
			result = append(result, ref)
		}
	}
	refs.Refs = newRefs
	bz = k.cdc.MustMarshalBinaryLengthPrefixed(refs)
	store.Set(k.getLiveAuctionListKey(), bz)

	bz = store.Get(k.getPastAuctionListKey())
	if len(bz) > 0 {
		k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &refs)
	} else {
		refs = AuctionRefs{}
	}
	refs.Refs = append(refs.Refs, result...)
	bz = k.cdc.MustMarshalBinaryLengthPrefixed(refs)
	store.Set(k.getPastAuctionListKey(), bz)

	return
}

func (k Keeper) GetAuctions(ctx sdk.Context, key []byte) (result []Auction) {
	var refs AuctionRefs

	store := ctx.KVStore(k.storeKey)

	bz := store.Get(key)
	if len(bz) > 0 {
		k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &refs)
	}

	for _, ref := range refs.Refs {
		a, found := k.GetAuction(ctx, ref.Id)
		if !found {
			panic("Auction doesn't exist!")
		}
		result = append(result, a)
	}
	return
}

func (k Keeper) GetPastAuctions(ctx sdk.Context) (result []Auction) {
	return k.GetAuctions(ctx, k.getPastAuctionListKey())
}

func (k Keeper) GetLiveAuctions(ctx sdk.Context) (result []Auction) {
	return k.GetAuctions(ctx, k.getLiveAuctionListKey())
}

// GetAuction gets an auction from the store by auctionID
func (k Keeper) GetAuction(ctx sdk.Context, auctionID ID) (Auction, bool) {
	var auction Auction

	store := ctx.KVStore(k.storeKey)

	bz := store.Get(k.getAuctionKey(auctionID))
	if bz == nil {
		return auction, false // TODO what is the correct behavior when an auction is not found? gov module follows this pattern of returning a bool
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &auction)

	return auction, true
}

// deleteAuction removes an auction from the store without any validation
func (k Keeper) deleteAuction(ctx sdk.Context, auctionID ID) {
	// remove from queue
	// auction, found := k.GetAuction(ctx, auctionID)
	// if found {
	// 	k.RemoveFromList(ctx, auction.GetEndTime(), auctionID)
	// }
	//
	// // delete auction
	// store := ctx.KVStore(k.storeKey)
	// store.Delete(k.getAuctionKey(auctionID))
}

// ---------- Queue and key methods ----------
// These are lower level function used by the store methods above.

func (k Keeper) getNextAuctionIDKey() []byte {
	return []byte("nextAuctionID")
}
func (k Keeper) getAuctionKey(auctionID ID) []byte {
	return []byte(fmt.Sprintf("auctions:%d", auctionID))
}

func (k Keeper) getLiveAuctionListKey() []byte {
	return []byte("auctionLiveList")
}

func (k Keeper) getPastAuctionListKey() []byte {
	return []byte("auctionPastList")
}

// GetAuctionIterator returns an iterator over all auctions in the store
// implement iterator over live auctions?
func (k Keeper) GetAuctionIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte("auction"))
}

var queueKeyPrefix = []byte("queue")
var keyDelimiter = []byte(":")

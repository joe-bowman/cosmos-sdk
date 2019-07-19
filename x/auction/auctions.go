package auction

import (
	"fmt"
	"sort"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// MaxAuctionDuration max length of auction, in blocks\
	DefaultAuctionDuration int64 = 5  // 60s
	MaxAuctionDuration     int64 = 10 // 120s /*1 * 24 * 3600 / 5*/ // roughly 1 days, at 5s block time // 17280
	// BidExtensionDuration how long an auction gets extended when someone bids, in blocks
	BidExtensionDuration int64 = 3 // 30s /*1800 / 5*/ // roughly 30 mins, at 5s block time // 360
	FeeAuctionFrequency  int64 = 25
)

// Auction is an interface to several types of auction.
type Auction interface {
	GetID() ID
	SetID(ID)
	GetBids() []Bid
	PlaceBid(currentBlockHeight int64, bidder sdk.AccAddress, bid sdk.Coin) sdk.Error
	GetEndTime() int64 // auctions close at the end of the block with blockheight EndTime (ie bids placed in that block are valid)
	String() string
}

// BaseAuction type shared by all Auctions
type BaseAuction struct {
	ID         ID
	Initiator  sdk.ValAddress // Person who starts the auction. Giving away Lot (aka seller in a forward auction)
	Lot        sdk.Coin       // Amount of coins up being given by initiator (FA - amount for sale by seller, RA - cost of good by buyer (bid))
	Bids       []Bid          // Placed bids
	EndTime    int64          // Block height at which the auction closes. It closes at the end of this block
	MaxEndTime int64          // Maximum closing time. Auctions can close before this but never after.
}

type Bid struct {
	Bidder sdk.AccAddress // AccAddress of Bidder
	Bid    sdk.Coin       // Value of bid.
	Height int64          // Height that bid was placed, for ordering of same value bids.
}

// ID type for auction IDs
type ID uint64

// NewIDFromString generate new auction ID from a string
func NewIDFromString(s string) (ID, error) {
	n, err := strconv.ParseUint(s, 10, 64) // copied from how the gov module rest handler's parse proposal IDs
	if err != nil {
		return 0, err
	}
	return ID(n), nil
}

// GetID getter for auction ID
func (a BaseAuction) GetID() ID { return a.ID }

// SetID setter for auction ID
func (a *BaseAuction) SetID(id ID) { a.ID = id }

// GetEndTime getter for auction end time
func (a BaseAuction) GetEndTime() int64 { return a.EndTime }

func (a BaseAuction) GetBids() []Bid {
	// sort bid list
	sort.Slice(a.Bids[:], func(i, j int) bool {
		return a.Bids[i].Bid.Amount.GTE(a.Bids[j].Bid.Amount) && a.Bids[i].Height < a.Bids[j].Height
	})
	return a.Bids
}

func (a BaseAuction) String() string {
	return fmt.Sprintf(`Auction %d:
  Initiator:              %s
  Lot:               			%s
  Bids:            		    %v
  End Time:   						%d
  Max End Time:      			%d`,
		a.GetID(), a.Initiator, a.Lot,
		a.GetBids(), a.GetEndTime(),
		a.MaxEndTime,
	)
}

// ForwardAuction type for forward auctions
type ForwardAuction struct {
	BaseAuction
}

// NewForwardAuction creates a new forward auction
func NewForwardAuction(seller sdk.ValAddress, lot sdk.Coin, initialBid sdk.Coin, endTime int64) ForwardAuction {
	auction := ForwardAuction{BaseAuction{
		// no ID
		Initiator:  seller,
		Lot:        lot,
		Bids:       []Bid{}, // send the proceeds from the first bid back to the seller
		EndTime:    endTime,
		MaxEndTime: endTime,
	}}
	return auction
}

// PlaceBid implements Auction
func (a *ForwardAuction) PlaceBid(currentBlockHeight int64, bidder sdk.AccAddress, bid sdk.Coin) sdk.Error {
	// TODO check lot size matches lot?
	// check auction has not closed
	if currentBlockHeight > a.EndTime {
		return sdk.ErrInternal("auction has closed")
	}

	a.Bids = append(a.Bids, Bid{bidder, bid, currentBlockHeight})

	// increment timeout // TODO into keeper?
	a.EndTime = min(int64(currentBlockHeight+BidExtensionDuration), int64(a.MaxEndTime)) // TODO is there a better way to structure these types?

	return nil
}

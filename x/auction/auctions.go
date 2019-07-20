package auction

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// MaxAuctionDuration max length of auction, _in blocks_
	DefaultAuctionDuration int64 = 80  // 60s
	MaxAuctionDuration     int64 = 100 // 120s /*1 * 24 * 3600 / 5*/ // roughly 1 days, at 5s block time // 17280
	// BidExtensionDuration how long an auction gets extended when someone bids, in blocks
	BidExtensionDuration int64 = 10 // 30s /*1800 / 5*/ // roughly 30 mins, at 5s block time // 360
	FeeAuctionFrequency  int64 = 120
)

// Auction is an interface to several types of auction.
type Auction interface {
	GetID() ID
	SetID(ID)
	GetBid() sdk.Coin
	GetBidder() sdk.AccAddress
	GetLot() sdk.Coin
	SetBid(sdk.Coin)
	SetBidder(sdk.AccAddress)
	PlaceBid(currentBlockHeight int64, bidder sdk.AccAddress, bid sdk.Coin) sdk.Error
	GetEndTime() int64 // auctions close at the end of the block with blockheight EndTime (ie bids placed in that block are valid)
	String() string
}

// BaseAuction type shared by all Auctions
type BaseAuction struct {
	ID         ID             `json:"auction_id"`
	Lot        sdk.Coin       `json:"lot"`          // Amount of coins up being given by initiator (FA - amount for sale by seller, RA - cost of good by buyer (bid))
	Bid        sdk.Coin       `json:"bid"`          // Current bid
	Bidder     sdk.AccAddress `json:"bidder"`       // Current bidder
	EndTime    int64          `json:"end_time"`     // Block height at which the auction closes. It closes at the end of this block
	MaxEndTime int64          `json:"max_end_time"` // Maximum closing time. Auctions can close before this but never after.
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

// GetBid
func (a BaseAuction) GetBid() sdk.Coin { return a.Bid }

// SetBid
func (a *BaseAuction) SetBid(bid sdk.Coin) { a.Bid = bid }

// GetBidder
func (a BaseAuction) GetBidder() sdk.AccAddress { return a.Bidder }

// SetBidder
func (a *BaseAuction) SetBidder(bid sdk.AccAddress) { a.Bidder = bid }

// GetLot
func (a BaseAuction) GetLot() sdk.Coin { return a.Lot }

// GetEndTime getter for auction end time
func (a BaseAuction) GetEndTime() int64 { return a.EndTime }

func (a BaseAuction) String() string {
	return fmt.Sprintf(`Auction %d:
  Lot:               			%s
  Current Bid:         		%v
	Current Bidder:         %v
  End Time:   						%d
  Max End Time:      			%d`,
		a.GetID(), a.Lot,
		a.Bid, a.Bidder, a.GetEndTime(),
		a.MaxEndTime,
	)
}

// ForwardAuction type for forward auctions
type ForwardAuction struct {
	BaseAuction
}

// NewForwardAuction creates a new forward auction
func NewForwardAuction(lot sdk.Coin, initialBid sdk.Coin, endTime int64) ForwardAuction {
	auction := ForwardAuction{BaseAuction{
		// no ID
		Lot:        lot,
		Bid:        sdk.Coin{"uatom", sdk.ZeroInt()},
		Bidder:     nil,
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
		return sdk.ErrInternal("Auction has closed")
	}

	if bid.Amount.LTE(a.Bid.Amount) {
		return sdk.ErrInternal("Bid amount is less than or equal to current best bid")
	}

	// increment timeout // TODO into keeper?
	a.EndTime = min(int64(currentBlockHeight+BidExtensionDuration), int64(a.MaxEndTime)) // TODO is there a better way to structure these types?

	return nil
}

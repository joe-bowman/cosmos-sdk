package auction

import sdk "github.com/cosmos/cosmos-sdk/types"

// MsgPlaceBid is the message type used to place a bid on any type of auction.
type MsgPlaceBid struct {
	AuctionID ID             `json:"auction_id"`
	Bid       sdk.Coin       `json:"bid"`
	Bidder    sdk.AccAddress `json:"bidder"` // This can be a buyer (who increments bid), or a seller (who decrements lot) TODO rename to be clearer?
}

// NewMsgPlaceBid returns a new MsgPlaceBid.
func NewMsgPlaceBid(auctionID ID, bidder sdk.AccAddress, bid sdk.Coin) MsgPlaceBid {
	return MsgPlaceBid{
		AuctionID: auctionID,
		Bidder:    bidder,
		Bid:       bid,
	}
}

// Route return the message type used for routing the message.
func (msg MsgPlaceBid) Route() string { return ModuleName }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgPlaceBid) Type() string { return "place_bid" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgPlaceBid) ValidateBasic() sdk.Error {
	if msg.Bid.Amount.LTE(sdk.ZeroInt()) {
		return sdk.ErrInternal("invalid (negative) bid amount")
	}

	// TODO check coin denoms
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgPlaceBid) GetSignBytes() []byte {
	bz := moduleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgPlaceBid) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Bidder}
}

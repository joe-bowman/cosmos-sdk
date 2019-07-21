package dao

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Governance message types and routes
const (
	TypeMsgDepositDao        = "deposit_dao"
	TypeMsgVoteDao           = "vote_dao"
	TypeMsgSubmitProposalDao = "submit_proposal_dao"

	MaxDescriptionLength int = 5000
	MaxTitleLength       int = 140
	//MaxDenomLength // TODO: TBD
)

var _, _, _ sdk.Msg = MsgSubmitProposalDao{}, MsgDepositDao{}, MsgVoteDao{}

// MsgSubmitProposal
type MsgSubmitProposalDao struct {
	Title          string         `json:"title"`           //  Title of the proposal
	Description    string         `json:"description"`     //  Description of the proposal
	// TODO: add Dao Proposal spec with denom, tx, rebalancing
	Denom 			string
	ProposalType   ProposalKind   `json:"proposal_type"`   //  Type of proposal. Initial set {PlainTextProposal}
	Proposer       sdk.AccAddress `json:"proposer"`        //  Address of the proposer
	InitialDeposit sdk.Coins      `json:"initial_deposit"` //  Initial deposit paid by sender. Must be strictly positive.
	// TODO: add content
}

func NewMsgSubmitProposal(title, description string, denom string, proposalType ProposalKind, proposer sdk.AccAddress, initialDeposit sdk.Coins) MsgSubmitProposalDao {
	return MsgSubmitProposalDao{
		Title:          title,
		Description:    description,
		Denom:			denom,
		ProposalType:   proposalType,
		Proposer:       proposer,
		InitialDeposit: initialDeposit,
	}
}

//nolint
func (msg MsgSubmitProposalDao) Route() string { return RouterKey }
func (msg MsgSubmitProposalDao) Type() string  { return TypeMsgSubmitProposalDao }

// Implements Msg.
func (msg MsgSubmitProposalDao) ValidateBasic() sdk.Error {
	if len(msg.Title) == 0 {
		return ErrInvalidTitle(DefaultCodespace, "No title present in proposal")
	}
	if len(msg.Title) > MaxTitleLength {
		return ErrInvalidTitle(DefaultCodespace, fmt.Sprintf("Proposal title is longer than max length of %d", MaxTitleLength))
	}
	if len(msg.Description) == 0 {
		return ErrInvalidDescription(DefaultCodespace, "No description present in proposal")
	}
	if len(msg.Description) > MaxDescriptionLength {
		return ErrInvalidDescription(DefaultCodespace, fmt.Sprintf("Proposal description is longer than max length of %d", MaxDescriptionLength))
	}
	//if len(msg.Denom) > MaxDenom // TODO: denom max length
	if !validProposalType(msg.ProposalType) {
		return ErrInvalidProposalType(DefaultCodespace, msg.ProposalType)
	}
	if msg.Proposer.Empty() {
		return sdk.ErrInvalidAddress(msg.Proposer.String())
	}
	if !msg.InitialDeposit.IsValid() {
		return sdk.ErrInvalidCoins(msg.InitialDeposit.String())
	}
	if msg.InitialDeposit.IsAnyNegative() {
		return sdk.ErrInvalidCoins(msg.InitialDeposit.String())
	}
	return nil
}

func (msg MsgSubmitProposalDao) String() string {
	return fmt.Sprintf("MsgSubmitProposal{%s, %s, %s, %s, %v}", msg.Title, msg.Description, msg.Denom, msg.ProposalType, msg.InitialDeposit)
}

// Implements Msg.
func (msg MsgSubmitProposalDao) GetSignBytes() []byte {
	bz := msgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// Implements Msg.
func (msg MsgSubmitProposalDao) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Proposer}
}

// MsgDeposit
type MsgDepositDao struct {
	ProposalID uint64         `json:"proposal_id"` // ID of the proposal
	Depositor  sdk.AccAddress `json:"depositor"`   // Address of the depositor
	Amount     sdk.Coins      `json:"amount"`      // Coins to add to the proposal's deposit
}

func NewMsgDeposit(depositor sdk.AccAddress, proposalID uint64, amount sdk.Coins) MsgDepositDao {
	return MsgDepositDao{
		ProposalID: proposalID,
		Depositor:  depositor,
		Amount:     amount,
	}
}

// Implements Msg.
// nolint
func (msg MsgDepositDao) Route() string { return RouterKey }
func (msg MsgDepositDao) Type() string  { return TypeMsgDepositDao }

// Implements Msg.
func (msg MsgDepositDao) ValidateBasic() sdk.Error {
	if msg.Depositor.Empty() {
		return sdk.ErrInvalidAddress(msg.Depositor.String())
	}
	if !msg.Amount.IsValid() {
		return sdk.ErrInvalidCoins(msg.Amount.String())
	}
	if msg.Amount.IsAnyNegative() {
		return sdk.ErrInvalidCoins(msg.Amount.String())
	}
	if msg.ProposalID < 0 {
		return ErrUnknownProposal(DefaultCodespace, msg.ProposalID)
	}
	return nil
}

func (msg MsgDepositDao) String() string {
	return fmt.Sprintf("MsgDeposit{%s=>%v: %v}", msg.Depositor, msg.ProposalID, msg.Amount)
}

// Implements Msg.
func (msg MsgDepositDao) GetSignBytes() []byte {
	bz := msgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// Implements Msg.
func (msg MsgDepositDao) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Depositor}
}

// MsgVote
type MsgVoteDao struct {
	ProposalID uint64         `json:"proposal_id"` // ID of the proposal
	Voter      sdk.AccAddress `json:"voter"`       //  address of the voter
	Option     VoteOption     `json:"option"`      //  option from OptionSet chosen by the voter
	VoteAmount sdk.Coin       `json:"vote_amount"` // it will be staked(deposit)  need to bigger than balance,
													// need to same denom with proposal.denom
}

func NewMsgVoteDao(voter sdk.AccAddress, proposalID uint64, option VoteOption, voteAmount sdk.Coin) MsgVoteDao {
	return MsgVoteDao{
		ProposalID: proposalID,
		Voter:      voter,
		Option:     option,
		VoteAmount: voteAmount,
	}
}

// Implements Msg.
// nolint
func (msg MsgVoteDao) Route() string { return RouterKey }
func (msg MsgVoteDao) Type() string  { return TypeMsgVoteDao }

// Implements Msg.
func (msg MsgVoteDao) ValidateBasic() sdk.Error {
	if msg.Voter.Empty() {
		return sdk.ErrInvalidAddress(msg.Voter.String())
	}
	if msg.ProposalID < 0 {
		return ErrUnknownProposal(DefaultCodespace, msg.ProposalID)
	}
	if !validVoteOption(msg.Option) {
		return ErrInvalidVote(DefaultCodespace, msg.Option)
	}
	return nil
}

func (msg MsgVoteDao) String() string {
	return fmt.Sprintf("MsgVote{%v - %s}", msg.ProposalID, msg.Option)
}

// Implements Msg.
func (msg MsgVoteDao) GetSignBytes() []byte {
	bz := msgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// Implements Msg.
func (msg MsgVoteDao) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Voter}
}

package dao

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var msgCdc = codec.New()

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSubmitDao{}, "cosmos-sdk/MsgSubmitDao", nil)
	cdc.RegisterConcrete(MsgDepositDao{}, "cosmos-sdk/MsgDeposit", nil)
	cdc.RegisterConcrete(MsgVoteDao{}, "cosmos-sdk/MsgVote", nil)

	cdc.RegisterInterface((*ProposalContent)(nil), nil)
	cdc.RegisterConcrete(TextProposal{}, "dao/TextProposal", nil)
}

func init() {
	RegisterCodec(msgCdc)
}

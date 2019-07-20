package dao

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var msgCdc = codec.New()

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSubmitProposalDao{}, "cosmos-sdk/MsgSubmitProposalDao", nil)
	cdc.RegisterConcrete(MsgDepositDao{}, "cosmos-sdk/MsgDepositDao", nil)
	cdc.RegisterConcrete(MsgVoteDao{}, "cosmos-sdk/MsgVoteDao", nil)

	cdc.RegisterInterface((*ProposalContent)(nil), nil)
	cdc.RegisterConcrete(TextProposal{}, "dao/TextProposal", nil)
}

func init() {
	RegisterCodec(msgCdc)
}

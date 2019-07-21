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

// RegisterProposalTypeCodec registers an external proposal content type defined
// in another module for the internal ModuleCdc. This allows the MsgSubmitProposal
// to be correctly Amino encoded and decoded.
func RegisterProposalTypeCodec(o interface{}, name string) {
	msgCdc.RegisterConcrete(o, name, nil)
}

// TODO determine a good place to seal this codec
func init() {
	RegisterCodec(msgCdc)
}

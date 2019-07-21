package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/dao"
)

// module codec
var ModuleCdc = codec.New()

// RegisterCodec registers all the necessary types and interfaces for
// governance.
//func RegisterCodec(cdc *codec.Codec) {
//	cdc.RegisterInterface((*Content)(nil), nil)
//
//	cdc.RegisterConcrete(MsgSubmitProposal{}, "cosmos-sdk/MsgSubmitProposal", nil)
//	cdc.RegisterConcrete(MsgDeposit{}, "cosmos-sdk/MsgDeposit", nil)
//	cdc.RegisterConcrete(MsgVote{}, "cosmos-sdk/MsgVote", nil)
//
//	cdc.RegisterConcrete(TextProposal{}, "cosmos-sdk/TextProposal", nil)
//	cdc.RegisterConcrete(SoftwareUpgradeProposal{}, "cosmos-sdk/SoftwareUpgradeProposal", nil)
//}

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(dao.MsgSubmitProposalDao{}, "cosmos-sdk/MsgSubmitProposalDao", nil)
	cdc.RegisterConcrete(dao.MsgDepositDao{}, "cosmos-sdk/MsgDepositDao", nil)
	cdc.RegisterConcrete(dao.MsgVoteDao{}, "cosmos-sdk/MsgVoteDao", nil)

	cdc.RegisterInterface((*dao.ProposalContent)(nil), nil)
	cdc.RegisterConcrete(dao.TextProposal{}, "dao/TextProposal", nil)
}

// RegisterProposalTypeCodec registers an external proposal content type defined
// in another module for the internal ModuleCdc. This allows the MsgSubmitProposal
// to be correctly Amino encoded and decoded.
func RegisterProposalTypeCodec(o interface{}, name string) {
	ModuleCdc.RegisterConcrete(o, name, nil)
}

// TODO determine a good place to seal this codec
func init() {
	RegisterCodec(ModuleCdc)
}

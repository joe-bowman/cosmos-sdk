package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// SubModuleCdc defines the IBC tendermint client codec.
var SubModuleCdc *codec.Codec

// RegisterCodec registers the Tendermint types
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(ClientState{}, "ibc/client/wasm/ClientState", nil)
	cdc.RegisterConcrete(ConsensusState{}, "ibc/client/wasm/ConsensusState", nil)
	cdc.RegisterConcrete(Header{}, "ibc/client/wasm/Header", nil)
	cdc.RegisterConcrete(Evidence{}, "ibc/client/wasm/Evidence", nil)
	cdc.RegisterConcrete(MsgCreateWasmClient{}, "ibc/client/MsgCreateWasmClient", nil)
	cdc.RegisterConcrete(MsgUpdateWasmClient{}, "ibc/client/MsgUpdateWasmClient", nil)
	cdc.RegisterConcrete(MsgSubmitWasmClientMisbehaviour{}, "ibc/client/MsgSubmitWasmClientMisbehaviour", nil)
	cdc.RegisterConcrete(MsgStoreClientCode{}, "ibc/client/MsgStoreWasmClientCode", nil)
	cdc.RegisterConcrete(MsgWrappedData{}, "ibc/client/MsgWrappedData", nil)

	SetSubModuleCodec(cdc)
}

// SetSubModuleCodec sets the ibc tendermint client codec
func SetSubModuleCodec(cdc *codec.Codec) {
	SubModuleCdc = cdc
}

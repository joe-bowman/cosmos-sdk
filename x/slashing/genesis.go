package slashing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState - all slashing state that must be provided at genesis
type GenesisState struct {
	Params DefaultParams
}

// HubDefaultGenesisState - default GenesisState used by Cosmos Hub
func HubDefaultGenesisState() GenesisState {
	return GenesisState{
		Params: HubDefaultParams(),
	}
}

// TODO: use ConfigStore(see issue #1771)
// InitGenesis initialize default parameter
// takes the pointer to the keeper because default params is not store in KVStore
func InitGenesis(ctx sdk.Context, keeper *Keeper, data GenesisState) error {
	keeper.params.Defaults = &data.Params
	return nil
}

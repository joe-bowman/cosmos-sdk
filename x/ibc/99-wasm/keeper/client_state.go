package keeper

import (
	"errors"
	"fmt"
	tmmath "github.com/tendermint/tendermint/libs/math"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	ibcwasmtypes "github.com/cosmos/cosmos-sdk/x/ibc/99-wasm/types"
)

// InitializeFromMsg creates a tendermint client state from a CreateClientMsg
func (k *Keeper) InitializeFromMsg(ctx sdk.Context,
	msg ibcwasmtypes.MsgCreateWasmClient,
) (ibcwasmtypes.ClientState, error) {
	return k.Initialize(ctx, msg.GetClientID(), msg.TrustLevel, msg.TrustingPeriod, msg.UnbondingPeriod, msg.MaxClockDrift, msg.Header, msg.WasmId)
}

// Initialize creates a client state and validates its contents, checking that
// the provided consensus state is from the same client type.
func (k *Keeper) Initialize(
	ctx sdk.Context,
	id string, trustLevel tmmath.Fraction, trustingPeriod, ubdPeriod, maxClockDrift time.Duration,
	header ibcwasmtypes.Header, wasmId int,
) (ibcwasmtypes.ClientState, error) {
	if trustingPeriod >= ubdPeriod {
		return ibcwasmtypes.ClientState{}, errors.New("trusting period should be < unbonding period")
	}

	contractAddress, err := k.Instantiate(ctx, uint64(wasmId), ibcwasmtypes.ModuleAccount, []byte(""), fmt.Sprintf("wasm-client-%s-%d", id, wasmId))
	if err != nil {
		return ibcwasmtypes.ClientState{}, err
	}
	clientState := ibcwasmtypes.NewClientState(
		id, trustLevel, trustingPeriod, ubdPeriod, maxClockDrift, header, contractAddress,
	)
	return clientState, nil
}

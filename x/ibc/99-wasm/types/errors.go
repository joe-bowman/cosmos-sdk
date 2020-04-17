package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInvalidTrustingPeriod    = sdkerrors.Register(SubModuleName, 1, "invalid trusting period")
	ErrInvalidUnbondingPeriod   = sdkerrors.Register(SubModuleName, 2, "invalid unbonding period")
	ErrInvalidHeader            = sdkerrors.Register(SubModuleName, 3, "invalid header")
	ErrInvalidValidityPredicate = sdkerrors.Register(SubModuleName, 4, "invalid validity predicate")
	ErrCreateFailed             = sdkerrors.Register(SubModuleName, 5, "create wasm contract failed")

	// ErrAccountExists error for a contract account that already exists
	ErrAccountExists = sdkerrors.Register(SubModuleName, 6, "contract account already exists")

	// ErrInstantiateFailed error for rust instantiate contract failure
	ErrInstantiateFailed = sdkerrors.Register(SubModuleName, 7, "instantiate wasm contract failed")

	// ErrExecuteFailed error for rust execution contract failure
	ErrExecuteFailed = sdkerrors.Register(SubModuleName, 8, "execute wasm contract failed")

	// ErrGasLimit error for out of gas
	ErrGasLimit = sdkerrors.Register(SubModuleName, 9, "insufficient gas")

	// ErrInvalidGenesis error for invalid genesis file syntax
	ErrInvalidGenesis = sdkerrors.Register(SubModuleName, 10, "invalid genesis")

	// ErrNotFound error for an entry not found in the store
	ErrNotFound = sdkerrors.Register(SubModuleName, 11, "not found")

	// ErrQueryFailed error for rust smart query contract failure
	ErrQueryFailed      = sdkerrors.Register(SubModuleName, 12, "query wasm contract failed")
	ErrUnmarshalAddress = sdkerrors.Register(SubModuleName, 13, "unable to unmarshal address")

	ErrInvalidMaxClockDrift   = sdkerrors.Register(SubModuleName, 14, "invalid max clock drift")
	ErrTrustingPeriodExpired  = sdkerrors.Register(SubModuleName, 15, "time since latest trusted state has passed the trusting period")
	ErrUnbondingPeriodExpired = sdkerrors.Register(SubModuleName, 16, "time since latest trusted state has passed the unbonding period")
)

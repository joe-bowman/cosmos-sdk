package slashing

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/params"
)

// nolint
const (
	MaxEvidenceAgeKey           = "slashing/MaxEvidenceAge"
	SignedBlocksWindowKey       = "slashing/SignedBlocksWindow"
	MinSignedPerWindowKey       = "slashing/MinSignedPerWindow"
	DoubleSignUnbondDurationKey = "slashing/DoubleSignUnbondDuration"
	DowntimeUnbondDurationKey   = "slashing/DowntimeUnbondDuration"
	SlashFractionDoubleSignKey  = "slashing/SlashFractionDoubleSign"
	SlashFractionDowntimeKey    = "slashing/SlashFractionDowntime"
)

// DefaultParams - used for initializing default parameter for slashing at genesis
type DefaultParams struct {
	MaxEvidenceAge           int64
	SignedBlocksWindow       int64
	MinSignedPerWindow       sdk.Rat
	DoubleSignUnbondDuration int64
	DowntimeUnbondDuration   int64
	SlashFractionDoubleSign  sdk.Rat
	SlashFractionDowntime    sdk.Rat
}

// Default parameters used by Cosmos Hub
func HubDefaultParams() DefaultParams {
	return DefaultParams{
		// defaultMaxEvidenceAge = 60 * 60 * 24 * 7 * 3
		// TODO Temporarily set to 2 minutes for testnets.
		MaxEvidenceAge: 60 * 2,

		// TODO Temporarily set to five minutes for testnets
		DoubleSignUnbondDuration: 60 * 5,

		// TODO Temporarily set to 100 blocks for testnets
		SignedBlocksWindow: 100,

		// TODO Temporarily set to 10 minutes for testnets
		DowntimeUnbondDuration: 60 * 10,

		MinSignedPerWindow: sdk.NewRat(1, 2),

		SlashFractionDoubleSign: sdk.NewRat(1).Quo(sdk.NewRat(20)),

		SlashFractionDowntime: sdk.NewRat(1).Quo(sdk.NewRat(100)),
	}
}

// Wrapper for params.Getter with default parameter
type Params struct {
	Params   params.Getter
	Defaults *DefaultParams
}

// MaxEvidenceAge - Max age for evidence - 21 days (3 weeks)
// MaxEvidenceAge = 60 * 60 * 24 * 7 * 3
func (p Params) MaxEvidenceAge(ctx sdk.Context) int64 {
	return p.Params.GetInt64WithDefault(ctx, MaxEvidenceAgeKey, p.Defaults.MaxEvidenceAge)
}

// SignedBlocksWindow - sliding window for downtime slashing
func (p Params) SignedBlocksWindow(ctx sdk.Context) int64 {
	fmt.Printf("%+v\n", p.Defaults)
	return p.Params.GetInt64WithDefault(ctx, SignedBlocksWindowKey, p.Defaults.SignedBlocksWindow)
}

// Downtime slashing thershold - p.Defaults. 50% of the SignedBlocksWindow
func (p Params) MinSignedPerWindow(ctx sdk.Context) int64 {
	minSignedPerWindow := p.Params.GetRatWithDefault(ctx, MinSignedPerWindowKey, p.Defaults.MinSignedPerWindow)
	signedBlocksWindow := p.SignedBlocksWindow(ctx)
	return sdk.NewRat(signedBlocksWindow).Mul(minSignedPerWindow).RoundInt64()
}

// Double-sign unbond duration
func (p Params) DoubleSignUnbondDuration(ctx sdk.Context) int64 {
	return p.Params.GetInt64WithDefault(ctx, DoubleSignUnbondDurationKey, p.Defaults.DoubleSignUnbondDuration)
}

// Downtime unbond duration
func (p Params) DowntimeUnbondDuration(ctx sdk.Context) int64 {
	return p.Params.GetInt64WithDefault(ctx, DowntimeUnbondDurationKey, p.Defaults.DowntimeUnbondDuration)
}

// SlashFractionDoubleSign - currently p.Defaults. 5%
func (p Params) SlashFractionDoubleSign(ctx sdk.Context) sdk.Rat {
	return p.Params.GetRatWithDefault(ctx, SlashFractionDoubleSignKey, p.Defaults.SlashFractionDoubleSign)
}

// SlashFractionDowntime - currently p.Defaults. 1%
func (p Params) SlashFractionDowntime(ctx sdk.Context) sdk.Rat {
	return p.Params.GetRatWithDefault(ctx, SlashFractionDowntimeKey, p.Defaults.SlashFractionDowntime)
}

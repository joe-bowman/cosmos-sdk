package slashing

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func testEqualParams(t *testing.T, ctx sdk.Context, params DefaultParams, keeper Keeper) {
	require.Equal(t, params.MaxEvidenceAge, keeper.params.MaxEvidenceAge(ctx))
	require.Equal(t, params.SignedBlocksWindow, keeper.params.SignedBlocksWindow(ctx))
	require.Equal(t, sdk.NewRat(params.SignedBlocksWindow).Mul(params.MinSignedPerWindow).RoundInt64(), keeper.params.MinSignedPerWindow(ctx))
	require.Equal(t, params.DoubleSignUnbondDuration, keeper.params.DoubleSignUnbondDuration(ctx))
	require.Equal(t, params.DowntimeUnbondDuration, keeper.params.DowntimeUnbondDuration(ctx))

	require.Equal(t, params.SlashFractionDoubleSign, keeper.params.SlashFractionDoubleSign(ctx))
	require.Equal(t, params.SlashFractionDowntime, keeper.params.SlashFractionDowntime(ctx))

}

func TestGenesis(t *testing.T) {
	params := HubDefaultParams()

	ctx, _, _, setter, k := createTestInput(t, params)

	state := GenesisState{params}
	err := InitGenesis(ctx, &k, state)
	require.Nil(t, err)
	testEqualParams(t, ctx, params, k)

	params.MaxEvidenceAge = 1
	setter.SetInt64(ctx, MaxEvidenceAgeKey, params.MaxEvidenceAge)
	testEqualParams(t, ctx, params, k)

	params.SignedBlocksWindow = 1
	setter.SetInt64(ctx, SignedBlocksWindowKey, params.SignedBlocksWindow)
	testEqualParams(t, ctx, params, k)

	params.MinSignedPerWindow = sdk.OneRat()
	setter.SetRat(ctx, MinSignedPerWindowKey, params.MinSignedPerWindow)
	testEqualParams(t, ctx, params, k)

	params.DoubleSignUnbondDuration = 1
	setter.SetInt64(ctx, DoubleSignUnbondDurationKey, params.DoubleSignUnbondDuration)
	testEqualParams(t, ctx, params, k)

	params.DowntimeUnbondDuration = 1
	setter.SetInt64(ctx, DowntimeUnbondDurationKey, params.DowntimeUnbondDuration)
	testEqualParams(t, ctx, params, k)

	params.SlashFractionDoubleSign = sdk.OneRat()
	setter.SetRat(ctx, SlashFractionDoubleSignKey, params.SlashFractionDoubleSign)
	testEqualParams(t, ctx, params, k)

	params.SlashFractionDowntime = sdk.OneRat()
	setter.SetRat(ctx, SlashFractionDowntimeKey, params.SlashFractionDowntime)
	testEqualParams(t, ctx, params, k)
}

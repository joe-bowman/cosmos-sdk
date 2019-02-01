package app

import (
	"encoding/json"
	"fmt"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	stake "github.com/cosmos/cosmos-sdk/x/stake"
)

// export the state of gaia for a genesis file
func (app *GaiaApp) ExportAppStateAndValidators(forZeroHeight bool, kickValidators []sdk.ValAddress) (
	appState json.RawMessage, validators []tmtypes.GenesisValidator, err error) {

	// as if they could withdraw from the start of the next block
	ctx := app.NewContext(true, abci.Header{Height: app.LastBlockHeight()})

	app.prepareRemoveValidator(ctx, kickValidators)
	if forZeroHeight {
		app.prepForZeroHeightGenesis(ctx)
	}

	// iterate to get the accounts
	accounts := []GenesisAccount{}
	appendAccount := func(acc auth.Account) (stop bool) {
		account := NewGenesisAccountI(acc)
		accounts = append(accounts, account)
		return false
	}
	app.accountKeeper.IterateAccounts(ctx, appendAccount)

	genState := NewGenesisState(
		accounts,
		auth.ExportGenesis(ctx, app.feeCollectionKeeper),
		stake.ExportGenesis(ctx, app.stakeKeeper),
		mint.ExportGenesis(ctx, app.mintKeeper),
		distr.ExportGenesis(ctx, app.distrKeeper),
		gov.ExportGenesis(ctx, app.govKeeper),
		slashing.ExportGenesis(ctx, app.slashingKeeper),
	)
	appState, err = codec.MarshalJSONIndent(app.cdc, genState)
	if err != nil {
		return nil, nil, err
	}
	validators = stake.WriteValidators(ctx, app.stakeKeeper)
	return appState, validators, nil
}

// prepare for fresh start at zero height
func (app *GaiaApp) prepForZeroHeightGenesis(ctx sdk.Context) {

	/* Just to be safe, assert the invariants on current state. */
	app.assertRuntimeInvariantsOnContext(ctx)

	/* Handle fee distribution state. */

	// withdraw all delegator & validator rewards
	vdiIter := func(_ int64, valInfo distr.ValidatorDistInfo) (stop bool) {
		err := app.distrKeeper.WithdrawValidatorRewardsAll(ctx, valInfo.OperatorAddr)
		if err != nil {
			panic(err)
		}
		return false
	}
	app.distrKeeper.IterateValidatorDistInfos(ctx, vdiIter)

	ddiIter := func(_ int64, distInfo distr.DelegationDistInfo) (stop bool) {
		err := app.distrKeeper.WithdrawDelegationReward(
			ctx, distInfo.DelegatorAddr, distInfo.ValOperatorAddr)
		if err != nil {
			panic(err)
		}
		return false
	}
	app.distrKeeper.IterateDelegationDistInfos(ctx, ddiIter)

	app.assertRuntimeInvariantsOnContext(ctx)

	// set distribution info withdrawal heights to 0
	app.distrKeeper.IterateDelegationDistInfos(ctx, func(_ int64, delInfo distr.DelegationDistInfo) (stop bool) {
		delInfo.DelPoolWithdrawalHeight = 0
		app.distrKeeper.SetDelegationDistInfo(ctx, delInfo)
		return false
	})
	app.distrKeeper.IterateValidatorDistInfos(ctx, func(_ int64, valInfo distr.ValidatorDistInfo) (stop bool) {
		valInfo.FeePoolWithdrawalHeight = 0
		app.distrKeeper.SetValidatorDistInfo(ctx, valInfo)
		return false
	})

	// assert that the fee pool is empty
	feePool := app.distrKeeper.GetFeePool(ctx)
	if !feePool.TotalValAccum.Accum.IsZero() {
		panic("unexpected leftover validator accum")
	}
	bondDenom := app.stakeKeeper.GetParams(ctx).BondDenom
	if !feePool.ValPool.AmountOf(bondDenom).IsZero() {
		panic(fmt.Sprintf("unexpected leftover validator pool coins: %v",
			feePool.ValPool.AmountOf(bondDenom).String()))
	}

	// reset fee pool height, save fee pool
	feePool.TotalValAccum = distr.NewTotalAccum(0)
	app.distrKeeper.SetFeePool(ctx, feePool)

	/* Handle stake state. */

	// iterate through validators by power descending, reset bond height, update bond intra-tx counter
	store := ctx.KVStore(app.keyStake)
	iter := sdk.KVStoreReversePrefixIterator(store, stake.ValidatorsByPowerIndexKey)
	counter := int16(0)
	for ; iter.Valid(); iter.Next() {
		addr := sdk.ValAddress(iter.Value())
		validator, found := app.stakeKeeper.GetValidator(ctx, addr)
		if !found {
			panic("expected validator, not found")
		}
		validator.BondHeight = 0
		validator.UnbondingHeight = 0
		app.stakeKeeper.SetValidator(ctx, validator)
		counter++
	}
	iter.Close()

	/* Handle slashing state. */

	// we have to clear the slashing periods, since they reference heights
	app.slashingKeeper.DeleteValidatorSlashingPeriods(ctx)

	// reset start height on signing infos
	app.slashingKeeper.IterateValidatorSigningInfos(ctx, func(addr sdk.ConsAddress, info slashing.ValidatorSigningInfo) (stop bool) {
		info.StartHeight = 0
		app.slashingKeeper.SetValidatorSigningInfo(ctx, addr, info)
		return false
	})
}

// kick validators/accounts from state
func (app *GaiaApp) prepareRemoveValidator(ctx sdk.Context, kickValidators []sdk.ValAddress) {
	// For each validator
	// 1. remove validator delegations
	// 2. unbond all outgoing delegations
	// 3. clear account to CommunityPool
	// 4. Cleanup
        feePool := app.distrKeeper.GetFeePool(ctx)
	for _, removeValidator := range kickValidators {
		// 1. Unbond all delegations
		delegations := app.stakeKeeper.GetValidatorDelegations(ctx, removeValidator)

		for _, delegation := range delegations {
			_, err := app.stakeKeeper.BeginUnbonding(ctx, delegation.DelegatorAddr, delegation.ValidatorAddr, delegation.Shares)
			if err != nil {
				panic("")
			}
		}

		stake.EndBlocker(ctx, app.stakeKeeper)
		ctx = ctx.WithBlockTime(ctx.BlockHeader().Time.Add(app.stakeKeeper.UnbondingTime(ctx)).Add(time.Hour * 24 * 7))
		stake.EndBlocker(ctx, app.stakeKeeper)

		app.assertRuntimeInvariantsOnContext(ctx)

		// 2. Remove delegations
		validatorDelegations := app.stakeKeeper.GetDelegatorDelegations(ctx, sdk.AccAddress(removeValidator), 10000)
		for _, delegation := range validatorDelegations {
			_, err := app.stakeKeeper.BeginUnbonding(ctx, delegation.DelegatorAddr, delegation.ValidatorAddr, delegation.Shares)
			if err != nil {
				panic("")
			}
		}

		stake.EndBlocker(ctx, app.stakeKeeper)
		ctx = ctx.WithBlockTime(ctx.BlockHeader().Time.Add(app.stakeKeeper.UnbondingTime(ctx)).Add(time.Hour * 24 * 7))
		stake.EndBlocker(ctx, app.stakeKeeper)

		// 3. WIPE account
		account := app.accountKeeper.GetAccount(ctx, sdk.AccAddress(removeValidator))
		coins := account.GetCoins()
		feePool = app.distrKeeper.GetFeePool(ctx)
		feePool.ValidateGenesis()
		feePool.CommunityPool = feePool.CommunityPool.Plus(distr.NewDecCoins(coins))
		app.distrKeeper.SetFeePool(ctx, feePool)

		account.SetCoins(sdk.Coins{})
		app.accountKeeper.SetAccount(ctx, account)

		app.assertRuntimeInvariantsOnContext(ctx)

		// 4. Cleanup
		// Nothing yet
	}
}

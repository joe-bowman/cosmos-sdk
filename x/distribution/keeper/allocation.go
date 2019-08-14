package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// allocate fees handles distribution of the collected fees
func (k Keeper) AllocateTokens(ctx sdk.Context, sumPreviousPrecommitPower, totalPreviousPower int64,
	previousProposer sdk.ConsAddress, previousVotes []abci.VoteInfo) {

	logger := ctx.Logger().With("module", "x/distribution")

	// fetch and clear the collected fees for distribution, since this is
	// called in BeginBlock, collected fees will be from the previous block
	// (and distributed to the previous proposer)
	feesCollectedInt := k.feeCollectionKeeper.GetCollectedFees(ctx)
	feesCollected := sdk.NewDecCoins(feesCollectedInt)
	k.feeCollectionKeeper.ClearCollectedFees(ctx)

	// temporary workaround to keep CanWithdrawInvariant happy
	// general discussions here: https://github.com/cosmos/cosmos-sdk/issues/2906#issuecomment-441867634
	feePool := k.GetFeePool(ctx)
	if totalPreviousPower == 0 {
		feePool.CommunityPool = feePool.CommunityPool.Add(feesCollected)
		k.SetFeePool(ctx, feePool)
		return
	}

	// calculate fraction votes
	previousFractionVotes := sdk.NewDec(sumPreviousPrecommitPower).Quo(sdk.NewDec(totalPreviousPower))

	// calculate previous proposer reward
	baseProposerReward := k.GetBaseProposerReward(ctx)
	bonusProposerReward := k.GetBonusProposerReward(ctx)
	proposerMultiplier := baseProposerReward.Add(bonusProposerReward.MulTruncate(previousFractionVotes))
	proposerReward := feesCollected.MulDecTruncate(proposerMultiplier)

	// pay previous proposer
	remaining := feesCollected
	proposerValidator := k.stakingKeeper.ValidatorByConsAddr(ctx, previousProposer)

	if proposerValidator != nil {
		k.AllocateTokensToValidator(ctx, proposerValidator, proposerReward)
		remaining = remaining.Sub(proposerReward)
	} else {
		// previous proposer can be unknown if say, the unbonding period is 1 block, so
		// e.g. a validator undelegates at block X, it's removed entirely by
		// block X+1's endblock, then X+2 we need to refer to the previous
		// proposer for X+1, but we've forgotten about them.
		logger.Error(fmt.Sprintf(
			"WARNING: Attempt to allocate proposer rewards to unknown proposer %s. "+
				"This should happen only if the proposer unbonded completely within a single block, "+
				"which generally should not happen except in exceptional circumstances (or fuzz testing). "+
				"We recommend you investigate immediately.",
			previousProposer.String()))
	}

	// calculate fraction allocated to validators
	communityTax := k.GetCommunityTax(ctx)
	voteMultiplier := sdk.OneDec().Sub(proposerMultiplier).Sub(communityTax)

	// allocate tokens proportionally to voting power
	// TODO consider parallelizing later, ref https://github.com/cosmos/cosmos-sdk/pull/3099#discussion_r246276376
	for _, vote := range previousVotes {
		validator := k.stakingKeeper.ValidatorByConsAddr(ctx, vote.Validator.Address)

		// TODO consider microslashing for missing votes.
		// ref https://github.com/cosmos/cosmos-sdk/issues/2525#issuecomment-430838701
		powerFraction := sdk.NewDec(vote.Validator.Power).QuoTruncate(sdk.NewDec(totalPreviousPower))
		reward := feesCollected.MulDecTruncate(voteMultiplier).MulDecTruncate(powerFraction)
		k.AllocateTokensToValidator(ctx, validator, reward)
		remaining = remaining.Sub(reward)
	}

	// allocate community funding
	feePool.CommunityPool = feePool.CommunityPool.Add(remaining)
	k.SetFeePool(ctx, feePool)

}

// allocate tokens to a particular validator, splitting according to commission
func (k Keeper) AllocateTokensToValidator(ctx sdk.Context, val sdk.Validator, tokens sdk.DecCoins) {

	// handle commission properly, then bond everything to the requisite validator!
	// split tokens between validator and delegators according to commission

	// split tokens into bondDenomTokens and nonBondDenomTokens
	bondDenomTokens := tokens.Intersect(sdk.DecCoins{{k.stakingKeeper.GetParams(ctx).BondDenom, sdk.OneDec()}})
	nonBondDenomTokens := tokens.Except(sdk.DecCoins{{k.stakingKeeper.GetParams(ctx).BondDenom, sdk.OneDec()}})

	commission := bondDenomTokens.MulDec(val.GetCommission())
	// nonBondDenomTokens commission handled post-auction
	shared := bondDenomTokens.Sub(commission)

	// update current commission
	currentCommission := k.GetValidatorAccumulatedCommission(ctx, val.GetOperator())
	currentCommission = currentCommission.Add(commission)
	k.SetValidatorAccumulatedCommission(ctx, val.GetOperator(), currentCommission)

	if _, ok := val.(staking.Validator); ok {
		// huge massive caveat; at this point in time, we ignore any non baseDenom tokens.
		// TODO: don't do this, it sucks.
		k.stakingKeeper.AddValidatorTokens(ctx, staking.ValidatorFromSdkValidator(val), shared.AmountOf(k.stakingKeeper.GetParams(ctx).BondDenom).TruncateInt())
		k.stakingKeeper.AddFeesToAuctionPool(ctx, staking.ValidatorFromSdkValidator(val), nonBondDenomTokens)
	} else {
		fmt.Println("Well this is shit...")
	}

}

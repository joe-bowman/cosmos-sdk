package keeper

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"os"
)

// keeper of the staking store
type Keeper struct {
	storeKey            sdk.StoreKey
	cdc                 *codec.Codec
	paramSpace          params.Subspace
	bankKeeper          types.BankKeeper
	stakingKeeper       types.StakingKeeper
	feeCollectionKeeper types.FeeCollectionKeeper

	// codespace
	codespace sdk.CodespaceType
}

// create a new keeper
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, paramSpace params.Subspace, ck types.BankKeeper,
	sk types.StakingKeeper, fck types.FeeCollectionKeeper, codespace sdk.CodespaceType) Keeper {
	keeper := Keeper{
		storeKey:            key,
		cdc:                 cdc,
		paramSpace:          paramSpace.WithKeyTable(ParamKeyTable()),
		bankKeeper:          ck,
		stakingKeeper:       sk,
		feeCollectionKeeper: fck,
		codespace:           codespace,
	}
	return keeper
}

// set withdraw address
func (k Keeper) SetWithdrawAddr(ctx sdk.Context, delegatorAddr sdk.AccAddress, withdrawAddr sdk.AccAddress) sdk.Error {
	if !k.GetWithdrawAddrEnabled(ctx) {
		return types.ErrSetWithdrawAddrDisabled(k.codespace)
	}

	k.SetDelegatorWithdrawAddr(ctx, delegatorAddr, withdrawAddr)

	return nil
}

func (k Keeper) WithdrawDelegationRewards(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (sdk.Coins, sdk.Error) {
	return k.WithdrawDelegationRewardsWithSource(ctx, delAddr, valAddr, types.WithdrawSourceNoSource)
}

// withdraw rewards from a delegation
func (k Keeper) WithdrawDelegationRewardsWithSource(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, source int) (sdk.Coins, sdk.Error) {
	val := k.stakingKeeper.Validator(ctx, valAddr)
	if val == nil {
		return nil, types.ErrNoValidatorDistInfo(k.codespace)
	}

	del := k.stakingKeeper.Delegation(ctx, delAddr, valAddr)
	if del == nil {
		return nil, types.ErrNoDelegationDistInfo(k.codespace)
	}

	// withdraw rewards
	rewards, err := k.withdrawDelegationRewards(ctx, val, del, source)
	if err != nil {
		return nil, err
	}

	// reinitialize the delegation
	k.initializeDelegation(ctx, valAddr, delAddr)

	return rewards, nil
}

func (k Keeper) WithdrawValidatorCommission(ctx sdk.Context, valAddr sdk.ValAddress) (sdk.Coins, sdk.Error) {
	return k.WithdrawValidatorCommissionWithSource(ctx, valAddr, types.WithdrawSourceNoSource)
}

// withdraw validator commission
func (k Keeper) WithdrawValidatorCommissionWithSource(ctx sdk.Context, valAddr sdk.ValAddress, source int) (sdk.Coins, sdk.Error) {
	// fetch validator accumulated commission
	commission := k.GetValidatorAccumulatedCommission(ctx, valAddr)
	if commission.IsZero() {
		return nil, types.ErrNoValidatorCommission(k.codespace)
	}

	coins, remainder := commission.TruncateDecimal()
	k.SetValidatorAccumulatedCommission(ctx, valAddr, remainder) // leave remainder to withdraw later

	// update outstanding
	outstanding := k.GetValidatorOutstandingRewards(ctx, valAddr)
	k.SetValidatorOutstandingRewards(ctx, valAddr, outstanding.Sub(sdk.NewDecCoins(coins)))

	if !coins.IsZero() {
		accAddr := sdk.AccAddress(valAddr)
		withdrawAddr := k.GetDelegatorWithdrawAddr(ctx, accAddr)

		if _, _, err := k.bankKeeper.AddCoins(ctx, withdrawAddr, coins); err != nil {
			return nil, err
		}
	}

	if ctx.Value("ExtractDataMode") != nil {
		// for each coin type withdrawn, insert a row with 0 value.
		f, _ := os.OpenFile(fmt.Sprintf("./extract/unchecked/commission.%d.%s", ctx.BlockHeight(), ctx.ChainID()), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		defer f.Close()

		for _, coin := range commission {
			f.WriteString(fmt.Sprintf("%s,%s,%d,%d,%s,%s,%d\n", valAddr.String(), coin.Denom, uint64(coin.Amount.Int64()), uint64(ctx.BlockHeight()), ctx.BlockTime().Format("2006-01-02 15:04:05"), ctx.ChainID(), source))
			f.WriteString(fmt.Sprintf("%s,%s,%d,%d,%s,%s,%d\n", valAddr.String(), coin.Denom, 0, uint64(ctx.BlockHeight()), ctx.BlockTime().Format("2006-01-02 15:04:05"), ctx.ChainID(), types.WithdrawSourceZero))
		}
	}

	return coins, nil
}

package keeper

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/tendermint/tendermint/libs/log"
)

// Keeper of the distribution store
type Keeper struct {
	storeKey      sdk.StoreKey
	cdc           *codec.Codec
	paramSpace    params.Subspace
	stakingKeeper types.StakingKeeper
	supplyKeeper  types.SupplyKeeper

	codespace sdk.CodespaceType

	blacklistedAddrs map[string]bool

	feeCollectorName string // name of the FeeCollector ModuleAccount
}

// NewKeeper creates a new distribution Keeper instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, paramSpace params.Subspace,
	sk types.StakingKeeper, supplyKeeper types.SupplyKeeper, codespace sdk.CodespaceType,
	feeCollectorName string, blacklistedAddrs map[string]bool) Keeper {

	// ensure distribution module account is set
	if addr := supplyKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	return Keeper{
		storeKey:         key,
		cdc:              cdc,
		paramSpace:       paramSpace.WithKeyTable(ParamKeyTable()),
		stakingKeeper:    sk,
		supplyKeeper:     supplyKeeper,
		codespace:        codespace,
		feeCollectorName: feeCollectorName,
		blacklistedAddrs: blacklistedAddrs,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetWithdrawAddr sets a new address that will receive the rewards upon withdrawal
func (k Keeper) SetWithdrawAddr(ctx sdk.Context, delegatorAddr sdk.AccAddress, withdrawAddr sdk.AccAddress) sdk.Error {
	if k.blacklistedAddrs[withdrawAddr.String()] {
		return sdk.ErrUnauthorized(fmt.Sprintf("%s is blacklisted from receiving external funds", withdrawAddr))
	}

	if !k.GetWithdrawAddrEnabled(ctx) {
		return types.ErrSetWithdrawAddrDisabled(k.codespace)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSetWithdrawAddress,
			sdk.NewAttribute(types.AttributeKeyWithdrawAddress, withdrawAddr.String()),
		),
	)

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

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeWithdrawRewards,
			sdk.NewAttribute(sdk.AttributeKeyAmount, rewards.String()),
			sdk.NewAttribute(types.AttributeKeyValidator, valAddr.String()),
		),
	)

	// reinitialize the delegation
	k.initializeDelegation(ctx, valAddr, delAddr)
	return rewards, nil
}

// withdraw validator commission
func (k Keeper) WithdrawValidatorCommission(ctx sdk.Context, valAddr sdk.ValAddress) (sdk.Coins, sdk.Error) {
	return k.WithdrawValidatorCommissionWithSource(ctx, valAddr, types.WithdrawSourceNoSource)
}

func (k Keeper) WithdrawValidatorCommissionWithSource(ctx sdk.Context, valAddr sdk.ValAddress, source int) (sdk.Coins, sdk.Error) {
	// fetch validator accumulated commission
	accumCommission := k.GetValidatorAccumulatedCommission(ctx, valAddr)
	if accumCommission.IsZero() {
		return nil, types.ErrNoValidatorCommission(k.codespace)
	}

	commission, remainder := accumCommission.TruncateDecimal()
	k.SetValidatorAccumulatedCommission(ctx, valAddr, remainder) // leave remainder to withdraw later

	// update outstanding
	outstanding := k.GetValidatorOutstandingRewards(ctx, valAddr)
	k.SetValidatorOutstandingRewards(ctx, valAddr, outstanding.Sub(sdk.NewDecCoins(commission)))

	if !commission.IsZero() {
		accAddr := sdk.AccAddress(valAddr)
		withdrawAddr := k.GetDelegatorWithdrawAddr(ctx, accAddr)
		err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, withdrawAddr, commission)
		if err != nil {
			return nil, err
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeWithdrawCommission,
			sdk.NewAttribute(sdk.AttributeKeyAmount, commission.String()),
		),
	)

	if ctx.Value("ExtractDataMode") != nil {
		// for each coin type withdrawn, insert a row with 0 value.
		f2, _ := os.OpenFile(fmt.Sprintf("./extract/unchecked/commission.%d.%s", ctx.BlockHeight(), ctx.ChainID()), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		for _, coin := range commission {
			f2.WriteString(fmt.Sprintf("%s,%s,%d,%d,%s,%s,%d\n", valAddr.String(), coin.Denom, uint64(coin.Amount.Int64()), uint64(ctx.BlockHeight()), ctx.BlockTime().Format("2006-01-02 15:04:05"), ctx.ChainID(), source))
			f2.WriteString(fmt.Sprintf("%s,%s,%d,%d,%s,%s,%d\n", valAddr.String(), coin.Denom, 0, uint64(ctx.BlockHeight()), ctx.BlockTime().Format("2006-01-02 15:04:05"), ctx.ChainID(), types.WithdrawSourceZero))
		}
		f2.Close()
	}

	return commission, nil
}

// GetTotalRewards returns the total amount of fee distribution rewards held in the store
func (k Keeper) GetTotalRewards(ctx sdk.Context) (totalRewards sdk.DecCoins) {
	k.IterateValidatorOutstandingRewards(ctx,
		func(_ sdk.ValAddress, rewards types.ValidatorOutstandingRewards) (stop bool) {
			totalRewards = totalRewards.Add(rewards)
			return false
		},
	)
	return totalRewards
}

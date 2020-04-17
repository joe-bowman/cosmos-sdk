package keeper

import (
	"encoding/binary"
	"fmt"

	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	wasm "github.com/confio/go-cosmwasm"
	wasmTypes "github.com/confio/go-cosmwasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/x/ibc/99-wasm/types"
)

// GasMultiplier is how many cosmwasm gas points = 1 sdk gas point
// SDK reference costs can be found here: https://github.com/cosmos/cosmos-sdk/blob/02c6c9fafd58da88550ab4d7d494724a477c8a68/store/types/gas.go#L153-L164
// A write at ~3000 gas and ~200us = 10 gas per us (microsecond) cpu/io
// Rough timing have 88k gas at 90us, which is equal to 1k sdk gas... (one read)
const GasMultiplier = 100

// MaxGas for a contract is 900 million (enforced in rust)
const MaxGas = 900000000

// Keeper will have a reference to Wasmer with it's own data directory.
type Keeper struct {
	storeKey sdk.StoreKey
	cdc      *codec.Codec

	router sdk.Router

	wasmer wasm.Wasmer
	// queryGasLimit is the max wasm gas that can be spent on executing a query with a contract
	queryGasLimit uint64
}

// NewKeeper creates a new contract Keeper instance
func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey) Keeper {
	wasmConfig := types.DefaultWasmConfig()
	wasmer, err := wasm.NewWasmer(wasmConfig.HomeDir, wasmConfig.CacheSize)
	if err != nil {
		panic(err)
	}

	return Keeper{
		storeKey: storeKey,
		cdc:      cdc,
		wasmer:   *wasmer,
		//router:        router,
		queryGasLimit: wasmConfig.SmartQueryGasLimit,
	}
}

// Create uploads and compiles a WASM contract, returning a short identifier for the contract
func (k Keeper) Create(ctx sdk.Context, creator sdk.AccAddress, wasmCode []byte, source string, builder string) (codeID uint64, err error) {
	fmt.Println("DEBUG1")
	wasmCode, err = uncompress(wasmCode)
	if err != nil {
		return 0, sdkerrors.Wrap(types.ErrCreateFailed, err.Error())
	}
	fmt.Println("DEBUG2")
	codeHash, err := k.wasmer.Create(wasmCode)
	if err != nil {
		// return 0, sdkerrors.Wrap(err, "cosmwasm create")
		return 0, sdkerrors.Wrap(types.ErrCreateFailed, err.Error())
	}
	fmt.Println("DEBUG3")
	store := ctx.KVStore(k.storeKey)
	codeID = k.autoIncrementID(ctx, types.KeyLastCodeID)
	codeInfo := types.NewCodeInfo(codeHash, creator, source, builder)
	fmt.Println("DEBUG4")
	// 0x01 | codeID (uint64) -> ContractInfo
	store.Set(types.GetCodeKey(codeID), k.cdc.MustMarshalBinaryBare(codeInfo))

	return codeID, nil
}

// returns true when simulation mode used by gas=auto queries
func isSimulationMode(ctx sdk.Context) bool {
	return ctx.GasMeter().Limit() == 0 && ctx.BlockHeight() != 0
}

// Instantiate creates an instance of a WASM contract
func (k Keeper) Instantiate(ctx sdk.Context, codeID uint64, creator sdk.AccAddress, initMsg []byte, label string) (types.Address, error) {
	// create contract address
	contractAddress := k.generateContractAddress(ctx, codeID)
	ctx.Logger().Error("Debug 1", codeID, creator, initMsg)
	// get contact info
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetCodeKey(codeID))
	if bz == nil {
		return nil, sdkerrors.Wrap(types.ErrNotFound, "code")
	}
	ctx.Logger().Error("Debug 2")
	var codeInfo types.CodeInfo
	k.cdc.MustUnmarshalBinaryBare(bz, &codeInfo)
	ctx.Logger().Error("Debug 3", codeID, codeInfo)

	// prepare params for contract instantiate call
	params := types.NewParams(ctx, creator, contractAddress)
	ctx.Logger().Error("Debug 4", params)

	// create prefixed data store
	// 0x03 | contractAddress (sdk.AccAddress)
	prefixStoreKey := types.GetContractStorePrefixKey(contractAddress)
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), prefixStoreKey)

	// instantiate wasm contract
	gas := gasForContract(ctx)
	ctx.Logger().Error("Debug 5", prefixStoreKey)
	res, err := k.wasmer.Instantiate(codeInfo.CodeHash, params, initMsg, prefixStore, cosmwasmAPI, gas)
	if err != nil {
		ctx.Logger().Error("Got an error :(", err.Error())
		return contractAddress, sdkerrors.Wrap(types.ErrInstantiateFailed, err.Error())
		// return contractAddress, sdkerrors.Wrap(err, "cosmwasm instantiate")
	}
	consumeGas(ctx, res.GasUsed)

	// emit all events from this contract itself
	_, events := types.CosmosResult(*res, contractAddress)
	ctx.EventManager().EmitEvents(events)

	//err = k.dispatchMessages(ctx, contractAccount, res.Messages)
	//if err != nil {
	//	return nil, err
	//}

	// persist instance
	createdAt := types.NewCreatedAt(ctx)
	instance := types.NewContractInfo(codeID, creator, initMsg, label, createdAt)
	store.Set(types.GetContractAddressKey(contractAddress), k.cdc.MustMarshalBinaryBare(instance))

	return contractAddress, nil
}

// Execute executes the contract instance
func (k Keeper) Execute(ctx sdk.Context, contractAddress types.Address, caller sdk.AccAddress, msg []byte) (sdk.Result, error) {
	codeInfo, prefixStore, err := k.contractInstance(ctx, contractAddress)
	if err != nil {
		return sdk.Result{}, err
	}

	// add more funds
	// if !coins.IsZero() {
	// 	sdkerr := k.bankKeeper.SendCoins(ctx, caller, contractAddress, coins)
	// 	if sdkerr != nil {
	// 		return sdk.Result{}, sdkerr
	// 	}
	// }
	//contractAccount := k.accountKeeper.GetAccount(ctx, contractAddress)

	params := types.NewParams(ctx, caller, contractAddress)

	gas := gasForContract(ctx)
	res, execErr := k.wasmer.Execute(codeInfo.CodeHash, params, msg, prefixStore, cosmwasmAPI, gas)
	if execErr != nil {
		// TODO: wasmer doesn't return gas used on error. we should consume it (for error on metering failure)
		return sdk.Result{}, sdkerrors.Wrap(types.ErrExecuteFailed, execErr.Error())
	}
	consumeGas(ctx, res.GasUsed)

	// emit all events from this contract itself
	value, events := types.CosmosResult(*res, contractAddress)
	ctx.EventManager().EmitEvents(events)
	//value.Events = nil

	// TODO: capture events here as well
	//err = k.dispatchMessages(ctx, contractAccount, res.Messages)
	//if err != nil {
	//	return sdk.Result{}, err
	//}

	return value, nil
}

// QuerySmart queries the smart contract itself.
func (k Keeper) QuerySmart(ctx sdk.Context, contractAddr types.Address, req []byte) ([]byte, error) {
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(k.queryGasLimit))

	codeInfo, prefixStore, err := k.contractInstance(ctx, contractAddr)
	if err != nil {
		return nil, err
	}
	queryResult, gasUsed, qErr := k.wasmer.Query(codeInfo.CodeHash, req, prefixStore, cosmwasmAPI, gasForContract(ctx))
	if qErr != nil {
		return nil, sdkerrors.Wrap(types.ErrQueryFailed, qErr.Error())
	}
	consumeGas(ctx, gasUsed)
	return queryResult, nil
}

// QueryRaw returns the contract's state for give key. For a `nil` key a empty slice` result is returned.
func (k Keeper) QueryRaw(ctx sdk.Context, contractAddress types.Address, key []byte) []types.Model {
	result := make([]types.Model, 0)
	if key == nil {
		return result
	}
	prefixStoreKey := types.GetContractStorePrefixKey(contractAddress)
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), prefixStoreKey)

	if val := prefixStore.Get(key); val != nil {
		return append(result, types.Model{
			Key:   key,
			Value: val,
		})
	}
	return result
}

func (k Keeper) contractInstance(ctx sdk.Context, contractAddress types.Address) (types.CodeInfo, prefix.Store, error) {
	store := ctx.KVStore(k.storeKey)

	contractBz := store.Get(types.GetContractAddressKey(contractAddress))
	if contractBz == nil {
		return types.CodeInfo{}, prefix.Store{}, sdkerrors.Wrap(types.ErrNotFound, "contract")
	}
	var contract types.ContractInfo
	k.cdc.MustUnmarshalBinaryBare(contractBz, &contract)

	contractInfoBz := store.Get(types.GetCodeKey(contract.CodeID))
	if contractInfoBz == nil {
		return types.CodeInfo{}, prefix.Store{}, sdkerrors.Wrap(types.ErrNotFound, "contract info")
	}
	var codeInfo types.CodeInfo
	k.cdc.MustUnmarshalBinaryBare(contractInfoBz, &codeInfo)
	prefixStoreKey := types.GetContractStorePrefixKey(contractAddress)
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), prefixStoreKey)
	return codeInfo, prefixStore, nil
}

func (k Keeper) GetContractInfo(ctx sdk.Context, contractAddress types.Address) *types.ContractInfo {
	store := ctx.KVStore(k.storeKey)
	var contract types.ContractInfo
	contractBz := store.Get(types.GetContractAddressKey(contractAddress))
	if contractBz == nil {
		return nil
	}
	k.cdc.MustUnmarshalBinaryBare(contractBz, &contract)
	return &contract
}

func (k Keeper) setContractInfo(ctx sdk.Context, contractAddress types.Address, contract types.ContractInfo) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetContractAddressKey(contractAddress), k.cdc.MustMarshalBinaryBare(contract))
}

func (k Keeper) ListContractInfo(ctx sdk.Context, cb func(types.Address, types.ContractInfo) bool) {
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.ContractKeyPrefix)
	iter := prefixStore.Iterator(nil, nil)
	for ; iter.Valid(); iter.Next() {
		var contract types.ContractInfo
		k.cdc.MustUnmarshalBinaryBare(iter.Value(), &contract)
		// cb returns true to stop early
		if cb(iter.Key(), contract) {
			break
		}
	}
}

func (k Keeper) GetContractState(ctx sdk.Context, contractAddress types.Address) sdk.Iterator {
	prefixStoreKey := types.GetContractStorePrefixKey(contractAddress)
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), prefixStoreKey)
	return prefixStore.Iterator(nil, nil)
}

func (k Keeper) setContractState(ctx sdk.Context, contractAddress types.Address, models []types.Model) {
	prefixStoreKey := types.GetContractStorePrefixKey(contractAddress)
	prefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), prefixStoreKey)
	for _, model := range models {
		prefixStore.Set([]byte(model.Key), []byte(model.Value))
	}
}

func (k Keeper) GetCodeInfo(ctx sdk.Context, codeID uint64) *types.CodeInfo {
	store := ctx.KVStore(k.storeKey)
	var codeInfo types.CodeInfo
	codeInfoBz := store.Get(types.GetCodeKey(codeID))
	if codeInfoBz == nil {
		return nil
	}
	k.cdc.MustUnmarshalBinaryBare(codeInfoBz, &codeInfo)
	return &codeInfo
}

func (k Keeper) GetByteCode(ctx sdk.Context, codeID uint64) ([]byte, error) {
	store := ctx.KVStore(k.storeKey)
	var codeInfo types.CodeInfo
	codeInfoBz := store.Get(types.GetCodeKey(codeID))
	if codeInfoBz == nil {
		return nil, nil
	}
	k.cdc.MustUnmarshalBinaryBare(codeInfoBz, &codeInfo)
	return k.wasmer.GetCode(codeInfo.CodeHash)
}

func (k Keeper) dispatchMessages(ctx sdk.Context, contract authTypes.AccountI, msgs []wasmTypes.CosmosMsg) error {
	for _, msg := range msgs {
		if err := k.dispatchMessage(ctx, contract, msg); err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) dispatchMessage(ctx sdk.Context, contract authTypes.AccountI, msg wasmTypes.CosmosMsg) error {
	// maybe use this instead for the arg?
	contractAddr := contract.GetAddress()
	if msg.Send != nil {
		//return k.sendTokens(ctx, contractAddr, msg.Send.FromAddress, msg.Send.ToAddress, msg.Send.Amount)
	} else if msg.Contract != nil {
		targetAddr, stderr := types.AddressFromString(msg.Contract.ContractAddr)
		if stderr != nil {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, msg.Contract.ContractAddr)
		}
		//sentFunds, err := convertWasmCoinToSdkCoin(msg.Contract.Send)
		//if err != nil {
		//	return err
		//}
		_, err := k.Execute(ctx, targetAddr, contractAddr, msg.Contract.Msg)
		return err // may be nil
	} else if msg.Opaque != nil {
		msg, err := ParseOpaqueMsg(k.cdc, msg.Opaque)
		if err != nil {
			return err
		}
		return k.handleSdkMessage(ctx, contractAddr, msg)
	}
	// what is it?
	panic(fmt.Sprintf("Unknown CosmosMsg: %#v", msg))
}

// func convertCosmosSendMsg(from string, to string, coins []wasmTypes.Coin) (bank.MsgSend, error) {
// 	fromAddr, stderr := sdk.AccAddressFromBech32(from)
// 	if stderr != nil {
// 		return bank.MsgSend{}, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, from)
// 	}
// 	toAddr, stderr := sdk.AccAddressFromBech32(to)
// 	if stderr != nil {
// 		return bank.MsgSend{}, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, to)
// 	}
//
// 	toSend, err := convertWasmCoinToSdkCoin(coins)
// 	if err != nil {
// 		return bank.MsgSend{}, err
// 	}
// 	sendMsg := bank.MsgSend{
// 		FromAddress: fromAddr,
// 		ToAddress:   toAddr,
// 		Amount:      toSend,
// 	}
// 	return sendMsg, nil
// }

func convertWasmCoinToSdkCoin(coins []wasmTypes.Coin) (sdk.Coins, error) {
	var toSend sdk.Coins
	for _, coin := range coins {
		amount, ok := sdk.NewIntFromString(coin.Amount)
		if !ok {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, coin.Amount+coin.Denom)
		}
		c := sdk.Coin{
			Denom:  coin.Denom,
			Amount: amount,
		}
		toSend = append(toSend, c)
	}
	return toSend, nil
}

func (k Keeper) handleSdkMessage(ctx sdk.Context, contractAddr sdk.Address, msg sdk.Msg) error {
	// make sure this account can send it
	for _, acct := range msg.GetSigners() {
		if !acct.Equals(contractAddr) {
			return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "contract doesn't have permission")
		}
	}

	// find the handler and execute it
	h := k.router.Route(ctx, msg.Route())
	if h == nil {
		return sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, msg.Route())
	}
	//res, err := h(ctx, msg)
	//if err != nil {
	//	return err
	//}
	// redispatch all events, (type sdk.EventTypeMessage will be filtered out in the handler)
	//ctx.EventManager().EmitEvents(res.Events)

	return nil
}

func gasForContract(ctx sdk.Context) uint64 {
	meter := ctx.GasMeter()
	remaining := (meter.Limit() - meter.GasConsumed()) * GasMultiplier
	if remaining > MaxGas {
		return MaxGas
	}
	return remaining
}

func consumeGas(ctx sdk.Context, gas uint64) {
	consumed := gas / GasMultiplier
	ctx.GasMeter().ConsumeGas(consumed, "wasm contract")
}

// generates a contract address from codeID + instanceID
func (k Keeper) generateContractAddress(ctx sdk.Context, codeID uint64) types.Address {
	instanceID := k.autoIncrementID(ctx, types.KeyLastInstanceID)
	// NOTE: It is possible to get a duplicate address if either codeID or instanceID
	// overflow 32 bits. This is highly improbable, but something that could be refactored.
	contractID := codeID<<32 + instanceID
	return addrFromUint64(contractID)
}

func (k Keeper) GetNextCodeID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyLastCodeID)
	id := uint64(1)
	if bz != nil {
		id = binary.BigEndian.Uint64(bz)
	}
	return id
}

func (k Keeper) autoIncrementID(ctx sdk.Context, lastIDKey []byte) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(lastIDKey)
	id := uint64(1)
	if bz != nil {
		id = binary.BigEndian.Uint64(bz)
	}
	bz = sdk.Uint64ToBigEndian(id + 1)
	store.Set(lastIDKey, bz)
	return id
}

func addrFromUint64(id uint64) types.Address {
	addr := make([]byte, 20)
	addr[0] = 'C'
	binary.PutUvarint(addr[1:], id)
	return addr
}

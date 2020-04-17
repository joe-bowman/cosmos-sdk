package wasm

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/99-wasm/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/99-wasm/types"
)

const (
	AttributeKeyContract = "contract_address"
	AttributeKeyCodeID   = "code_id"
	AttributeSigner      = "signer"
)

// filterMessageEvents returns the same events with all of type == EventTypeMessage removed.
// this is so only our top-level message event comes through
func filterMessageEvents(manager *sdk.EventManager) sdk.Events {
	events := manager.Events()
	res := make([]sdk.Event, 0, len(events)+1)
	for _, e := range events {
		if e.Type != sdk.EventTypeMessage {
			res = append(res, e)
		}
	}
	return res
}

func HandleMsgWrappedData(ctx sdk.Context, k keeper.Keeper, msg types.MsgWrappedData) (*sdk.Result, error) {
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	result, err := k.Execute(ctx, msg.Target, types.ModuleAccount, msg.Data)
	if err != nil {
		return &sdk.Result{}, err
	}
	return &result, nil

}

func HandleMsgCreateClient(ctx sdk.Context, k keeper.Keeper, msg types.MsgCreateWasmClient) (*sdk.Result, error) {
	ctx.Logger().Info("Validating Create")
	err := msg.ValidateBasic()
	if err != nil {
		ctx.Logger().Info("Errored", err.Error())
		return nil, err
	}

	addr, err := k.Instantiate(ctx, uint64(msg.WasmId), types.ModuleAccount, []byte(fmt.Sprintf(`{"name": "%s", "message": "%s"}`, msg.ClientID, msg.Message)), "client")
	if err != nil {
		ctx.Logger().Info("Errored", err.Error())
		return nil, err
	}

	return &sdk.Result{
		Data:   addr,
		Events: nil,
	}, nil

}

func HandleMsgStoreCode(ctx sdk.Context, k keeper.Keeper, msg types.MsgStoreClientCode) (*sdk.Result, error) {
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	codeID, err := k.Create(ctx, msg.Sender, msg.WASMByteCode, msg.Source, msg.Builder)
	if err != nil {
		return nil, err
	}

	events := filterMessageEvents(ctx.EventManager())
	ourEvent := sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.SubModuleName),
		sdk.NewAttribute(AttributeSigner, msg.Sender.String()),
		sdk.NewAttribute(AttributeKeyCodeID, fmt.Sprintf("%d", codeID)),
	)

	return &sdk.Result{
		Data:   []byte(fmt.Sprintf("%d", codeID)),
		Events: append(events, ourEvent).ToABCIEvents(),
	}, nil
}

//func handleInstantiate(ctx sdk.Context, k Keeper, msg *MsgInstantiateContract) (*sdk.Result, error) {
//	contractAddr, err := k.Instantiate(ctx, msg.Code, msg.Sender, msg.InitMsg, msg.Label, msg.InitFunds)
//	if err != nil {
//		return nil, err
//	}
//
//	events := filterMessageEvents(ctx.EventManager())
//	ourEvent := sdk.NewEvent(
//		sdk.EventTypeMessage,
//		sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
//		sdk.NewAttribute(AttributeSigner, msg.Sender.String()),
//		sdk.NewAttribute(AttributeKeyCodeID, fmt.Sprintf("%d", msg.Code)),
//		sdk.NewAttribute(AttributeKeyContract, contractAddr.String()),
//	)
//
//	return &sdk.Result{
//		Data:   contractAddr,
//		Events: append(events, ourEvent),
//	}, nil
//}
//
//func handleExecute(ctx sdk.Context, k Keeper, msg *MsgExecuteContract) (*sdk.Result, error) {
//	res, err := k.Execute(ctx, msg.Contract, msg.Sender, msg.Msg, msg.SentFunds)
//	if err != nil {
//		return nil, err
//	}
//
//	events := filterMessageEvents(ctx.EventManager())
//	ourEvent := sdk.NewEvent(
//		sdk.EventTypeMessage,
//		sdk.NewAttribute(sdk.AttributeKeyModule, ModuleName),
//		sdk.NewAttribute(AttributeSigner, msg.Sender.String()),
//		sdk.NewAttribute(AttributeKeyContract, msg.Contract.String()),
//	)
//
//	res.Events = append(events, ourEvent)
//	return &res, nil
//}

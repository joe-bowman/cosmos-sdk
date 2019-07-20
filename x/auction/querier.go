package auction

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

const (
	QuerierRoute = ModuleName
	// QueryGetAuction command for getting the information about a particular auction
	QueryGetAuction = "getauctions"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryGetAuction:
			return queryAuctions(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown auction query endpoint")
		}
	}
}

func queryAuctions(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, keeper.GetLiveAuctions(ctx))
	if err2 != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

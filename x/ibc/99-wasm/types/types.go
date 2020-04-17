package types

import (
	"encoding/hex"
	"encoding/json"

	tmBytes "github.com/tendermint/tendermint/libs/bytes"

	wasmTypes "github.com/confio/go-cosmwasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const defaultLRUCacheSize = uint64(0)
const defaultQueryGasLimit = uint64(3000000)

// Model is a struct that holds a KV pair
type Model struct {
	// hex-encode key to read it better (this is often ascii)
	Key tmBytes.HexBytes `json:"key"`
	// base64-encode raw value
	Value []byte `json:"val"`
}

// CodeInfo is data for the uploaded contract WASM code
type CodeInfo struct {
	CodeHash []byte         `json:"code_hash"`
	Creator  sdk.AccAddress `json:"creator"`
	Source   string         `json:"source"`
	Builder  string         `json:"builder"`
}

// NewCodeInfo fills a new Contract struct
func NewCodeInfo(codeHash []byte, creator sdk.AccAddress, source string, builder string) CodeInfo {
	return CodeInfo{
		CodeHash: codeHash,
		Creator:  creator,
		Source:   source,
		Builder:  builder,
	}
}

type Address []byte

func (a Address) String() string {
	return "0x" + hex.EncodeToString(a)
}

// AddressFromString converts a string representation of an address into an address object.
func AddressFromString(addr string) (Address, error) {
	if addr[0:2] != "0x" {
		return Address{}, ErrUnmarshalAddress
	}
	return hex.DecodeString(addr[2:])
}

// ContractInfo stores a WASM contract instance
type ContractInfo struct {
	CodeID  uint64          `json:"code_id"`
	Creator sdk.AccAddress  `json:"creator"`
	Label   string          `json:"label"`
	InitMsg json.RawMessage `json:"init_msg,omitempty"`
	// never show this in query results, just use for sorting
	// (Note: when using json tag "-" amino refused to serialize it...)
	Created *CreatedAt `json:"created,omitempty"`
}

// CreatedAt can be used to sort contracts
type CreatedAt struct {
	// BlockHeight is the block the contract was created at
	BlockHeight int64
	// TxIndex is a monotonic counter within the block (actual transaction index, or gas consumed)
	TxIndex uint64
}

// LessThan can be used to sort
func (a *CreatedAt) LessThan(b *CreatedAt) bool {
	if a == nil {
		return true
	}
	if b == nil {
		return false
	}
	return a.BlockHeight < b.BlockHeight || (a.BlockHeight == b.BlockHeight && a.TxIndex < b.TxIndex)
}

// NewCreatedAt gets a timestamp from the context
func NewCreatedAt(ctx sdk.Context) *CreatedAt {
	// we must safely handle nil gas meters
	var index uint64
	meter := ctx.BlockGasMeter()
	if meter != nil {
		index = meter.GasConsumed()
	}
	return &CreatedAt{
		BlockHeight: ctx.BlockHeight(),
		TxIndex:     index,
	}
}

// NewContractInfo creates a new instance of a given WASM contract info
func NewContractInfo(codeID uint64, creator sdk.AccAddress, initMsg []byte, label string, createdAt *CreatedAt) ContractInfo {
	return ContractInfo{
		CodeID:  codeID,
		Creator: creator,
		InitMsg: initMsg,
		Label:   label,
		Created: createdAt,
	}
}

// NewParams initializes params for a contract instance
func NewParams(ctx sdk.Context, creator sdk.AccAddress, contract Address) wasmTypes.Env {
	return wasmTypes.Env{
		Block: wasmTypes.BlockInfo{
			Height:  ctx.BlockHeight(),
			Time:    ctx.BlockTime().Unix(),
			ChainID: ctx.ChainID(),
		},
		Message: wasmTypes.MessageInfo{
			Signer:    wasmTypes.CanonicalAddress(creator),
			SentFunds: []wasmTypes.Coin{},
		},
		Contract: wasmTypes.ContractInfo{
			Address: wasmTypes.CanonicalAddress(contract),
			Balance: []wasmTypes.Coin{},
		},
	}
}

const CustomEventType = "wasm"
const AttributeKeyContractAddr = "contract_address"

// CosmosResult converts from a Wasm Result type
func CosmosResult(wasmResult wasmTypes.Result, contractAddr Address) (sdk.Result, sdk.Events) {
	var events sdk.Events
	if len(wasmResult.Log) > 0 {
		// we always tag with the contract address issuing this event
		attrs := []sdk.Attribute{sdk.NewAttribute(AttributeKeyContractAddr, contractAddr.String())}
		for _, l := range wasmResult.Log {
			// and reserve the contract_address key for our use (not contract)
			if l.Key != AttributeKeyContractAddr {
				attr := sdk.NewAttribute(l.Key, l.Value)
				attrs = append(attrs, attr)
			}
		}
		events = append(events, sdk.NewEvent(CustomEventType, attrs...))
	}
	return sdk.Result{
		Data:   []byte(wasmResult.Data),
		Events: events.ToABCIEvents(),
	}, events
}

// WasmConfig is the extra config required for wasm
type WasmConfig struct {
	SmartQueryGasLimit uint64 `mapstructure:"query_gas_limit"`
	CacheSize          uint64 `mapstructure:"lru_size"`
	HomeDir            string `mapstructure:"home_dir"`
}

// DefaultWasmConfig returns the default settings for WasmConfig
func DefaultWasmConfig() WasmConfig {
	return WasmConfig{
		SmartQueryGasLimit: defaultQueryGasLimit,
		CacheSize:          defaultLRUCacheSize,
		HomeDir:            "./wasm",
	}
}

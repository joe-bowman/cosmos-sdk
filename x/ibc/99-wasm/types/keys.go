package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
)

const (
	// ModuleName is the name of the contract module
	SubModuleName = "wasm"

	// StoreKey is the string store representation
	StoreKey = SubModuleName

	// TStoreKey is the string transient store representation
	TStoreKey = "transient_" + SubModuleName

	// QuerierRoute is the querier route for the staking module
	QuerierRoute = SubModuleName

	// RouterKey is the msg router key for the staking module
	RouterKey = SubModuleName
)

// nolint
var (
	KeyLastCodeID     = []byte("lastCodeId")
	KeyLastInstanceID = []byte("lastContractId")

	CodeKeyPrefix       = []byte{0x01}
	ContractKeyPrefix   = []byte{0x02}
	ContractStorePrefix = []byte{0x03}

	ModuleAccount = sdk.AccAddress(crypto.AddressHash([]byte("ibc-wasm")))
)

// GetCodeKey constructs the key for retreiving the ID for the WASM code
func GetCodeKey(contractID uint64) []byte {
	contractIDBz := sdk.Uint64ToBigEndian(contractID)
	return append(CodeKeyPrefix, contractIDBz...)
}

// GetContractAddressKey returns the key for the WASM contract instance
func GetContractAddressKey(addr Address) []byte {
	return append(ContractKeyPrefix, addr...)
}

// GetContractStorePrefixKey returns the store prefix for the WASM contract instance
func GetContractStorePrefixKey(addr Address) []byte {
	return append(ContractStorePrefix, addr...)
}

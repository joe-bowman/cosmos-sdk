package crypto

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	tmed25519 "github.com/tendermint/tendermint/crypto/ed25519"
	tmsecp256k1 "github.com/tendermint/tendermint/crypto/secp256k1"
)


func PubkeyFromTmPubkey(pubkey tmcrypto.PubKey) (valPubKey tmcrypto.PubKey, err error) {

	switch pubkey.(type) {
	case tmed25519.PubKey:
		valPubKey, err = ed25519.FromTmEd25519(pubkey)
		if err != nil {
			return nil, err
		}
		break
	case tmsecp256k1.PubKey:
		valPubKey, err = secp256k1.FromTmSecp256k1(pubkey)
		if err != nil {
			return nil, err
		}
		break
	default:
		return nil, fmt.Errorf("unexpected pubkey type, got %T", pubkey)
	}
	return valPubKey, nil
}
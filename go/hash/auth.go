package hash

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/eth/ethsecp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

// RecoverAddr recovers the sender address from msg and signature
func RecoverAddr(msg []byte, sig []byte) (sdk.AccAddress, ethsecp256k1.PubKey, error) {
	pubKeyByte, err := secp256k1.RecoverPubkey(msg, sig)
	if err != nil {
		return nil, ethsecp256k1.PubKey{}, err
	}
	pubKey, _ := ethcrypto.UnmarshalPubkey(pubKeyByte)
	pk := ethsecp256k1.PubKey{
		Key: ethcrypto.CompressPubkey(pubKey),
	}
	recoverAcc := sdk.AccAddress(pk.Address().Bytes())
	return recoverAcc, pk, nil
}

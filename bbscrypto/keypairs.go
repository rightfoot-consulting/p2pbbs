package bbscrypto

import (
	"os"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/mr-tron/base58/base58"
)

func SavePrivateKey(filepath string, privateKey crypto.PrivKey) (err error) {
	privKeyBytes, err := crypto.MarshalPrivateKey(privateKey)
	if err != nil {
		return
	}
	encoded := base58.Encode(privKeyBytes)
	err = os.WriteFile(filepath, []byte(encoded), 0600)
	return
}

func LoadPrivateKey(filepath string) (privateKey crypto.PrivKey, err error) {
	encoded, err := os.ReadFile(filepath)
	if err != nil {
		return
	}
	privKeyBytes, err := base58.Decode(string(encoded))
	if err != nil {
		return
	}
	privateKey, err = crypto.UnmarshalPrivateKey(privKeyBytes)
	return
}

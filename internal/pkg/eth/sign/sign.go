package sign

import (
	"log"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func Sign(nonce string, privateKeyStr string) (string, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		log.Fatal(err)
	}

	data := []byte(nonce)
	hash := crypto.Keccak256Hash(data)

	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return "", err
	}

	return hexutil.Encode(signature), nil
}

func ParseSign(nonce string, signatureHex string) (string, error) {
	signature, err := hexutil.Decode(signatureHex)
	data := []byte(nonce)
	hash := crypto.Keccak256Hash(data)

	sigPublicKeyECDSA, err := crypto.SigToPub(hash.Bytes(), signature)
	if err != nil {
		return "", err
	}

	address := crypto.PubkeyToAddress(*sigPublicKeyECDSA)

	return address.String(), nil
}

func ParsePersonal(signatureHex string, msg []byte) (common.Address, error) {
	sig, err := hexutil.Decode(signatureHex)
	if err != nil {
		return common.Address{}, err
	}

	msg = accounts.TextHash(msg)
	if len(sig) >= crypto.RecoveryIDOffset && (sig[crypto.RecoveryIDOffset] == 27 || sig[crypto.RecoveryIDOffset] == 28) {
		sig[crypto.RecoveryIDOffset] -= 27 // Transform yellow paper V from 27/28 to 0/1
	}

	recovered, err := crypto.SigToPub(msg, sig)
	if err != nil {
		return common.Address{}, err
	}

	recoveredAddr := crypto.PubkeyToAddress(*recovered)

	return recoveredAddr, nil
}

func Verify(from, signatureHex string, msg string) bool {
	recoveredAddr, err := ParsePersonal(signatureHex, []byte(msg))
	if err != nil {
		return false
	}

	return from == recoveredAddr.Hex()
}

package crypto

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/pem"
	"errors"
	"github.com/btcsuite/btcd/btcec"
	"github.com/tendermint/tendermint/crypto"
	"io/ioutil"
)

const (
	privateKeyStart = 7
	privateKeyEnd   = 39
	publicKeyStart  = 23
)

func LoadKeys(privKeyFile, pubKeyFile string) (privKey []byte, pubKey []byte) {
	privKeyPem, _ := ioutil.ReadFile(privKeyFile)
	pubKeyPem, _ := ioutil.ReadFile(pubKeyFile)
	privKeyBlock, _ := pem.Decode(privKeyPem)
	pubKeyBlock, _ := pem.Decode(pubKeyPem)
	privKey = privKeyBlock.Bytes[privateKeyStart:privateKeyEnd]
	pubKey = pubKeyBlock.Bytes[publicKeyStart:]
	return
}

func LoadSignature(signatureFile string) (signature []byte) {
	signature, _ = ioutil.ReadFile(signatureFile)
	return
}

func LoadTmKeys(privTmKey crypto.PrivKey, pubTmKey crypto.PubKey) (privKey []byte, pubKey []byte) {
	privKey = privTmKey.Bytes()[5:]
	pubKey = pubTmKey.Bytes()[5:]
	return
}

func Sign(privKey, message []byte) []byte {
	hash := sha256.Sum256(message)
	key, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKey)
	sign, _ := key.Sign(hash[:])
	return sign.Serialize()
}

func Verify(pubKey, message []byte, signature []byte) bool {
	hash := sha256.Sum256(message)
	key, err := btcec.ParsePubKey(pubKey, btcec.S256())
	if err != nil {
		return false
	}
	sign, err := btcec.ParseSignature(signature, btcec.S256())
	if err != nil {
		return false
	}
	return sign.Verify(hash[:], key)
}

func CheckPubKey(pubKey []byte) error {
	_, err := btcec.ParsePubKey(pubKey, btcec.S256())
	return err
}

func SignED(privKey, message []byte) []byte {
	return ed25519.Sign(privKey, message)
}

func VerifyED(pubKey, message []byte, signature []byte) bool {
	return ed25519.Verify(pubKey, message, signature)
}

func CheckEDPubKey(pubKey []byte) error {
	if len(pubKey) != ed25519.PublicKeySize {
		return errors.New("invalid public ed25519 key length")
	}
	return nil
}

package main

import (
	"crypto/sha256"
	"encoding/pem"
	"github.com/btcsuite/btcd/btcec"
	ecies "github.com/ecies/go"
	"io/ioutil"
)

const (
	privateKeyStart = 7
	privateKeyEnd   = 39
	publicKeyStart  = 23
)

func LoadKeys(privKeyFile string, pubKeyFile string) (privKey []byte, pubKey []byte) {
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

func Sign(privKey []byte, message []byte) (signature []byte) {
	hash := sha256.Sum256(message)
	key, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKey)
	sign, _ := key.Sign(hash[:])
	return sign.Serialize()
}

func Verify(pubKey []byte, message []byte, signature []byte) (signed bool) {
	hash := sha256.Sum256(message)
	key, _ := btcec.ParsePubKey(pubKey, btcec.S256())
	sign, _ := btcec.ParseSignature(signature, btcec.S256())
	return sign.Verify(hash[:], key)
}

func Encrypt(pubKey []byte, message []byte) (encrypted []byte) {
	key, _ := btcec.ParsePubKey(pubKey, btcec.S256())
	encrypted, _ = btcec.Encrypt(key, message)
	return
}

func Encrypt2(pubKey []byte, message []byte) (encrypted []byte) {
	key, _ := ecies.NewPublicKeyFromBytes(pubKey)
	encrypted, _ = ecies.Encrypt(key, message)
	return
}

func Decrypt(privKey []byte, encrypted []byte) (message []byte) {
	key, _ := btcec.PrivKeyFromBytes(btcec.S256(), privKey)
	message, _ = btcec.Decrypt(key, encrypted)
	return
}

func Decrypt2(privKey []byte, decrypted []byte) (message []byte) {
	key := ecies.NewPrivateKeyFromBytes(privKey)
	decrypted, _ = ecies.Decrypt(key, message)
	return
}

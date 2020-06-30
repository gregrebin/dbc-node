package main

import (
	"bytes"
	"io/ioutil"
	"os/exec"
	"testing"
)

const (
	privKeyFile   = testDirectory + "privkey.pem"
	pubKeyFile    = testDirectory + "pubkey.pem"
	signatureFile = testDirectory + "signature.sing"
	messageFile   = testDirectory + "file.txt"
)

var (
	privKey []byte
	pubKey []byte
)

func init() {
	_ = exec.Command("./key.sh", ecParamFile, privKeyFile, pubKeyFile).Run()
	_ = ioutil.WriteFile(messageFile, []byte("Some message inside a file"), 0644)
	_ = exec.Command("./sign.sh", privKeyFile, signatureFile, messageFile).Run()
	privKey, pubKey = loadKeys(privKeyFile, pubKeyFile)
}

func TestSignature(t *testing.T) {
	message := []byte("Some message to be signed")
	signature := sign(privKey, message)
	signed := verify(pubKey, message, signature)
	if !signed {
		t.Fail()
	} // TODO: check some wrong signature
}

func TestOpenSslSignature(t *testing.T) {
	message, _ := ioutil.ReadFile(messageFile)
	signature := loadSignature(signatureFile)
	signed := verify(pubKey, message, signature)
	if !signed {
		t.Fail()
	} // TODO: check some wrong signature
}

func TestEncryption(t *testing.T) {
	message := []byte("Some message to be encrypted and decrypted")
	encrypted := encrypt(pubKey, message)
	decrypted := decrypt(privKey, encrypted)
	difference := bytes.Compare(message, decrypted)
	if difference != 0 {
		t.Fail()
	} // TODO: check some wrong encryption
}

package tests

import (
	"bytes"
	"dbc-node/crypto"
	"io/ioutil"
	"os/exec"
	"testing"
)

const (
	privKeyFile   = testDirectory + "privkey.pem"
	pubKeyFile    = testDirectory + "pubkey.pem"
	signatureFile = testDirectory + "Signature.sing"
	messageFile   = testDirectory + "file.txt"
)

var (
	privKey []byte
	pubKey  []byte
)

func init() {
	_ = exec.Command("./key.sh", ecParamFile, privKeyFile, pubKeyFile).Run()
	_ = ioutil.WriteFile(messageFile, []byte("Some message inside a file"), 0644)
	_ = exec.Command("./sign.sh", privKeyFile, signatureFile, messageFile).Run()
	privKey, pubKey = crypto.LoadKeys(privKeyFile, pubKeyFile)
}

func TestSignature(t *testing.T) {
	message := []byte("Some message to be signed")
	signature := crypto.Sign(privKey, message)
	signed := crypto.Verify(pubKey, message, signature)
	if !signed {
		t.Fail()
	}
}

func TestOpenSslSignature(t *testing.T) {
	message, _ := ioutil.ReadFile(messageFile)
	signature := crypto.LoadSignature(signatureFile)
	signed := crypto.Verify(pubKey, message, signature)
	if !signed {
		t.Fail()
	}
}

func TestEncryption(t *testing.T) {
	message := []byte("Some message to be encrypted and decrypted")
	encrypted := crypto.Encrypt(pubKey, message)
	decrypted := crypto.Decrypt(privKey, encrypted)
	difference := bytes.Compare(message, decrypted)
	if difference != 0 {
		t.Fail()
	}
}

package tests

import (
	"dbc-node/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
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

func TestSignatureED(t *testing.T) {
	message := []byte("Some message to be signed")
	tmKey := ed25519.GenPrivKey()
	privKeyEd, pubKeyEd := crypto.LoadTmKeys(tmKey, tmKey.PubKey())
	signature := crypto.SignED(privKeyEd, message)
	signed := crypto.VerifyED(pubKeyEd, message, signature)
	if !signed {
		t.Fail()
	}
}

func TestTmSignatureED(t *testing.T) {
	message := []byte("Some message to be signed")
	tmKey := ed25519.GenPrivKey()
	signature, _ := tmKey.Sign(message)
	_, pubKeyEd := crypto.LoadTmKeys(tmKey, tmKey.PubKey())
	signed := crypto.VerifyED(pubKeyEd, message, signature)
	if !signed {
		t.Fail()
	}
}

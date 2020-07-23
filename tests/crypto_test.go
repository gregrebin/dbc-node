package tests

import (
	"dbc-node/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"io/ioutil"
	"testing"
)

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

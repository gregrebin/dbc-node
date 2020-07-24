package tests

import (
	"crypto/sha256"
	"dbc-node/crypto"
	"dbc-node/modules"
	"encoding/hex"
	lorem "github.com/drhodes/golorem"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"io/ioutil"
	"os/exec"
)

const (
	// General
	testDirectory = "testdata/"
	ecParamFile   = testDirectory + "ecparam.pem"
	// Crypto
	privKeyFile   = testDirectory + "privkey.pem"
	pubKeyFile    = testDirectory + "pubkey.pem"
	signatureFile = testDirectory + "Signature.sing"
	messageFile   = testDirectory + "file.txt"
	// Modules
	requirerPrivKeyFile  = testDirectory + "requirerPrivKey.pem"
	requirerPubKeyFile   = testDirectory + "requirerPubKey.pem"
	validatorPrivKeyFile = testDirectory + "validatorPrivKey.pem"
	validatorPubKeyFile  = testDirectory + "validatorPubKey.pem"
	providerPrivKeyFile  = testDirectory + "providerPrivKey.pem"
	providerPubKeyFile   = testDirectory + "providerPubKey.pem"
	acceptorPrivKeyFile  = testDirectory + "acceptorPrivKey.pem"
	acceptorPubKeyFile   = testDirectory + "acceptorPubKey.pem"
)

var (
	// Crypto
	privKey []byte
	pubKey  []byte
	// Modules
	requirerPrivKey  []byte
	requirerPubKey   []byte
	validatorPrivKey []byte
	validatorPubKey  []byte
	providerPrivKey  []byte
	providerPubKey   []byte
	acceptorPrivKey  []byte
	acceptorPubKey   []byte
	// Dataset
	zpks      []zpk
	zpkToData []int
	// Balance
	tokenDistribution map[string]int64
	initialUsers      map[string]int64
	initialStake      int64
	stakePrivKey      []byte
	stakePubKey       []byte
	initialValidators map[string]int64
)

func init() {
	// Crypto
	_ = exec.Command("./key.sh", ecParamFile, privKeyFile, pubKeyFile).Run()
	_ = ioutil.WriteFile(messageFile, []byte("Some message inside a file"), 0644)
	_ = exec.Command("./sign.sh", privKeyFile, signatureFile, messageFile).Run()
	privKey, pubKey = crypto.LoadKeys(privKeyFile, pubKeyFile)
	// Modules
	_ = exec.Command("./key.sh", ecParamFile, requirerPrivKeyFile, requirerPubKeyFile).Run()
	requirerPrivKey, requirerPubKey = crypto.LoadKeys(requirerPrivKeyFile, requirerPubKeyFile)
	_ = exec.Command("./key.sh", ecParamFile, validatorPrivKeyFile, validatorPubKeyFile).Run()
	validatorPrivKey, validatorPubKey = crypto.LoadKeys(validatorPrivKeyFile, validatorPubKeyFile)
	_ = exec.Command("./key.sh", ecParamFile, providerPrivKeyFile, providerPubKeyFile).Run()
	providerPrivKey, providerPubKey = crypto.LoadKeys(providerPrivKeyFile, providerPubKeyFile)
	_ = exec.Command("./key.sh", ecParamFile, acceptorPrivKeyFile, acceptorPubKeyFile).Run()
	acceptorPrivKey, acceptorPubKey = crypto.LoadKeys(acceptorPrivKeyFile, acceptorPubKeyFile)
	// Dataset
	for i := 0; i < 10; i++ {
		secret := []byte(lorem.Sentence(0, 5))
		proof := sha256.Sum256(secret)
		info := sha256.Sum256(proof[:])
		zpks = append(zpks, zpk{secret: secret, proof: proof[:], info: info[:]})
	}
	zpkToData = []int{0, 0, 0, 0, 2, 2, 2, 3, 3, 3}
	// Balance
	initialUsers = make(map[string]int64, 4)
	tokenDistribution = map[string]int64{
		"Requirer":  modules.ToSats(25),
		"Validator": modules.ToSats(5),
		"Provider":  modules.ToSats(10),
		"Acceptor":  modules.ToSats(15),
	}
	initialUsers[hex.EncodeToString(requirerPubKey)] = tokenDistribution["Requirer"]
	initialUsers[hex.EncodeToString(validatorPubKey)] = tokenDistribution["Validator"]
	initialUsers[hex.EncodeToString(providerPubKey)] = tokenDistribution["Provider"]
	initialUsers[hex.EncodeToString(acceptorPubKey)] = tokenDistribution["Acceptor"]
	initialStake = 30
	tmPrivKey := ed25519.GenPrivKey()
	stakePrivKey, stakePubKey = crypto.LoadTmKeys(tmPrivKey, tmPrivKey.PubKey())
	initialValidators = make(map[string]int64, 1)
	initialValidators[hex.EncodeToString(stakePubKey)] = modules.ToSats(initialStake)
}

type zpk struct {
	secret []byte
	proof  []byte
	info   []byte
}

package main

import (
	"bytes"
	"crypto/sha256"
	"github.com/drhodes/golorem"
	"os/exec"
	"reflect"
	"testing"
)

const (
	requirerPrivKeyFile  = testDirectory + "requirerPrivKey.pem"
	requirerPubKeyFile   = testDirectory + "requirerPubKey.pem"
	validatorPrivKeyFile = testDirectory + "validatorPrivKey.pem"
	validatorPubKeyFile  = testDirectory + "validatorPubKey.pem"
	providerPrivKeyFile  = testDirectory + "providerPrivKey.pem"
	providerPubKeyFile   = testDirectory + "providerPubKey.pem"
	acceptorPrivKeyFile  = testDirectory + "acceptorPrivKey.pem"
	acceptorPubKeyFile   = testDirectory + "acceptorPubKey.pem"
)

type zpk struct {
	secret []byte
	proof  []byte
	info   []byte
}

var (
	requirerPrivKey  []byte
	requirerPubKey   []byte
	validatorPrivKey []byte
	validatorPubKey  []byte
	providerPrivKey  []byte
	providerPubKey   []byte
	acceptorPrivKey  []byte
	acceptorPubKey   []byte
	zpks             []zpk
	zpkToData        []int
)

// ------------------------------------------------------------------------------------------------------------------- //
// INITIALIZATION

func init() {
	initKeys()
	initZpk()
}

func initKeys() {
	_ = exec.Command("./key.sh", ecParamFile, requirerPrivKeyFile, requirerPubKeyFile).Run()
	requirerPrivKey, requirerPubKey = loadKeys(requirerPrivKeyFile, requirerPubKeyFile)
	_ = exec.Command("./key.sh", ecParamFile, validatorPrivKeyFile, validatorPubKeyFile).Run()
	validatorPrivKey, validatorPubKey = loadKeys(validatorPrivKeyFile, validatorPubKeyFile)
	_ = exec.Command("./key.sh", ecParamFile, providerPrivKeyFile, providerPubKeyFile).Run()
	providerPrivKey, providerPubKey = loadKeys(providerPrivKeyFile, providerPubKeyFile)
	_ = exec.Command("./key.sh", ecParamFile, acceptorPrivKeyFile, acceptorPubKeyFile).Run()
	acceptorPrivKey, acceptorPubKey = loadKeys(acceptorPrivKeyFile, acceptorPubKeyFile)
}

func initZpk() {
	for i := 0; i < 10; i++ {
		secret := []byte(lorem.Sentence(0, 5))
		proof := sha256.Sum256(secret)
		info := sha256.Sum256(proof[:])
		zpks = append(zpks, zpk{secret: secret, proof: proof[:], info: info[:]})
	}
	zpkToData = []int{0, 0, 0, 0, 2, 2, 2, 3, 3, 3}
}

// ------------------------------------------------------------------------------------------------------------------- //
// EMPTY STATE

func TestEmptyState(t *testing.T) {
	state := NewState()
	checkNil(state.dataList, "Data list", t)
	validHash := sha256.Sum256(nil)
	checkHash(state.stateHash, validHash[:], "Empty state", t)
	checkHash(state.hash(), validHash[:], "Hash function return", t)
	checkHash(state.stateHash, validHash[:], "Empty state after hash function", t)

}

// ------------------------------------------------------------------------------------------------------------------- //
// NEW DATA

func TestAddData(t *testing.T) {
	state := mockState(false, false, false)
	checkValidData(state, t)
	checkValidData(state, t)
	checkValidData(state, t)
}

func checkValidData(state *State, t *testing.T) {
	initialLength := len(state.dataList)
	otherDataHash, _ := dataHash(state, initialLength)

	description := makeDescription()
	state.addData(description)
	data := state.dataList[initialLength]

	checkLength(state.dataList, initialLength + 1, "Data list", t)
	checkNil(data.versionList, "Version list", t)
	compareDescription(data.description, description, t)
	dataHash := sha256.Sum256(description.hash())
	checkHash(data.dataHash, dataHash[:], "Data", t)
	stateHash := sha256.Sum256(append(otherDataHash, dataHash[:]...))
	checkHash(state.stateHash, stateHash[:], "State", t)
}

func makeDescription() Description {
	providerInfo := []byte(lorem.Sentence(10, 20))
	dataInfo := []byte(lorem.Sentence(10, 20))
	signature := sign(requirerPrivKey, append(providerInfo, dataInfo...))
	description := Description{
		providerInfo:      providerInfo,
		dataInfo:          dataInfo,
		trustedValidators: [][]byte{validatorPubKey},
		acceptor:          acceptorPubKey,
		requirer:          requirerPubKey,
		signature:         signature,
	}
	return description
}

func compareDescription(desc1, desc2 Description, t *testing.T) {
	if bytes.Compare(desc1.providerInfo, desc2.providerInfo) != 0 {
		t.Errorf("Corrupted provider info")
	}
	if bytes.Compare(desc1.dataInfo, desc2.dataInfo) != 0 {
		t.Errorf("Corrupted data info")
	}
	if bytes.Compare(desc1.trustedValidators[0], desc2.trustedValidators[0]) != 0 {
		t.Errorf("Corrupted trusted validator")
	}
	if bytes.Compare(desc1.acceptor, desc2.acceptor) != 0 {
		t.Errorf("Corrupted acceptor")
	}
	if bytes.Compare(desc1.requirer, desc2.requirer) != 0 {
		t.Errorf("Corrupted requirer")
	}
	if bytes.Compare(desc1.signature, desc2.signature) != 0 {
		t.Errorf("Corrupted signature")
	}
}

// ------------------------------------------------------------------------------------------------------------------- //
// NEW VALIDATION

func TestAddValidation(t *testing.T) {
	state := mockState(true, false, false)

	checkValidation(state, 0, zpks[0], t)
	checkValidation(state, 2, zpks[1], t)
	checkValidation(state, 1, zpks[2], t)
	checkValidation(state, 2, zpks[3], t)
	checkValidation(state, 2, zpks[4], t)
	checkValidation(state, 2, zpks[5], t)

	// TODO: check invalid validation (wrong signature, etc.)
}

func checkValidation(state *State, dataIndex int, zpk zpk, t *testing.T) {
	dataLength, versionLength := dataLength(state, dataIndex)
	dataHashL, dataHashR := dataHash(state, dataIndex)
	otherVersionHash, _ := versionHash(state, dataIndex, versionLength) // since we add a new version, we will have only versions at the left

	validation := makeValidation(zpk)
	state.addValidation(validation, dataIndex)
	data := state.dataList[dataIndex]
	version := state.dataList[dataIndex].versionList[versionLength]

	checkLength(state.dataList, dataLength, "Data list", t)
	checkLength(state.dataList[dataIndex].versionList, versionLength + 1, "Version list", t)
	checkEmpty(&version.acceptedPayload, "Accepted payload", t)
	checkEmpty(&version.payload, "Payload", t)
	compareValidation(version.validation, validation, t)
	emptyAcceptedPayload := AcceptedPayload{}
	emptyPayload := Payload{}
	versionHash := sha256.Sum256(append(append(emptyAcceptedPayload.hash(), emptyPayload.hash()...), validation.hash()...))
	checkHash(version.versionHash, versionHash[:], "Version", t)
	dataHash := sha256.Sum256(append(append(data.description.hash(), otherVersionHash...), versionHash[:]...))
	checkHash(data.dataHash, dataHash[:], "Data", t)
	stateHash := sha256.Sum256(append(append(dataHashL, dataHash[:]...), dataHashR...))
	checkHash(state.stateHash, stateHash[:], "State", t)
}

func makeValidation(zpk zpk) Validation {
	validationInfo := zpk.info
	signature := sign(validatorPrivKey, validationInfo[:])
	return Validation{
		info:          validationInfo[:],
		validatorAddr: validatorPubKey,
		signature:     signature,
	}
}

func compareValidation(val1, val2 Validation, t *testing.T) {
	if bytes.Compare(val1.info, val2.info) != 0 {
		t.Errorf("Corrupted info")
	}
	if bytes.Compare(val1.validatorAddr, val2.validatorAddr) != 0 {
		t.Errorf("Corrupted validator adddress")
	}
	if bytes.Compare(val1.signature, val2.signature) != 0 {
		t.Errorf("Corrupted signature")
	}
}

// ------------------------------------------------------------------------------------------------------------------- //
// NEW PAYLOAD

func TestAddPayload(t *testing.T) {
	state := mockState(true, true, false)
	var versionIndex int
	var lastDataIndex int
	for zpkIndex, dataIndex := range zpkToData {
		if lastDataIndex != dataIndex {
			versionIndex = 0
			lastDataIndex = dataIndex
		}
		checkPayload(state, dataIndex, versionIndex, zpks[zpkIndex], t)
		versionIndex++
	}
}

func checkPayload(state *State, dataIndex, versionIndex int, zpk zpk, t *testing.T) {
	dataLength, versionLength := dataLength(state, dataIndex)
	dataHashL, dataHashR := dataHash(state, dataIndex)
	versionHashL, versionHashR := versionHash(state, dataIndex, versionIndex)
	initialVersion := state.dataList[dataIndex].versionList[versionIndex]

	payload := makePayload(zpk)
	state.addPayload(payload, dataIndex, versionIndex)
	data := state.dataList[dataIndex]
	version := state.dataList[dataIndex].versionList[versionIndex]

	checkLength(state.dataList, dataLength, "Data list", t)
	checkLength(state.dataList[dataIndex].versionList, versionLength, "Version list", t)
	checkEmpty(&version.acceptedPayload, "Accepted payload", t)
	comparePayload(version.payload, payload, t)
	compareValidation(version.validation, initialVersion.validation, t)
	emptyAcceptedPayload := AcceptedPayload{}
	versionHash := sha256.Sum256(append(append(emptyAcceptedPayload.hash(), payload.hash()...), initialVersion.validation.hash()...))
	checkHash(version.versionHash, versionHash[:], "Version", t)
	dataHash := sha256.Sum256(append(append(append(data.description.hash(), versionHashL...), versionHash[:]...), versionHashR...))
	checkHash(data.dataHash, dataHash[:], "Data", t)
	stateHash := sha256.Sum256(append(append(dataHashL, dataHash[:]...), dataHashR...))
	checkHash(state.stateHash, stateHash[:], "State", t)
}

func makePayload(zpk zpk) Payload {
	data := []byte(lorem.Sentence(10, 50))
	signature := sign(providerPrivKey, append(data, zpk.proof...))
	return Payload{
		data:         data,
		proof:        zpk.proof,
		providerAddr: providerPubKey,
		signature:    signature,
	}
}

func comparePayload(payload1, payload2 Payload, t *testing.T) {
	if bytes.Compare(payload1.data, payload2.data) != 0 {
		t.Errorf("Corrupted data")
	}
	if bytes.Compare(payload1.proof, payload2.proof) != 0 {
		t.Errorf("Corrupted proof")
	}
	if bytes.Compare(payload1.providerAddr, payload2.providerAddr) != 0 {
		t.Errorf("Corrupted provider address")
	}
	if bytes.Compare(payload1.signature, payload2.signature) != 0 {
		t.Errorf("Corrupted signature")
	}
}

// ------------------------------------------------------------------------------------------------------------------- //
// ACCEPT PAYLOAD

func TestAcceptPayload(t *testing.T) {
	state := mockState(true, true, true)
	for dataIndex, data := range state.dataList {
		for versionIndex := range data.versionList {
			checkAcceptedPayload(state, dataIndex, versionIndex, t)
		}
	}
}

func checkAcceptedPayload(state *State, dataIndex, versionIndex int, t *testing.T)  {
	dataLength, versionLength := dataLength(state, dataIndex)
	dataHashL, dataHashR := dataHash(state, dataIndex)
	versionHashL, versionHashR := versionHash(state, dataIndex, versionIndex)
	initialVersion := state.dataList[dataIndex].versionList[versionIndex]

	acceptedPayload := makeAcceptedPayload()
	state.acceptPayload(acceptedPayload, dataIndex, versionIndex)
	data := state.dataList[dataIndex]
	version := state.dataList[dataIndex].versionList[versionIndex]

	checkLength(state.dataList, dataLength, "Data list", t)
	checkLength(state.dataList[dataIndex].versionList, versionLength, "Version list", t)
	compareAcceptedPayload(version.acceptedPayload, acceptedPayload, t)
	comparePayload(version.payload, initialVersion.payload, t)
	compareValidation(version.validation, initialVersion.validation, t)
	versionHash := sha256.Sum256(append(append(acceptedPayload.hash(), initialVersion.payload.hash()...), initialVersion.validation.hash()...))
	checkHash(version.versionHash, versionHash[:], "Version", t)
	dataHash := sha256.Sum256(append(append(append(data.description.hash(), versionHashL...), versionHash[:]...), versionHashR...))
	checkHash(data.dataHash, dataHash[:], "Data", t)
	stateHash := sha256.Sum256(append(append(dataHashL, dataHash[:]...), dataHashR...))
	checkHash(state.stateHash, stateHash[:], "State", t)
}

func makeAcceptedPayload() AcceptedPayload {
	data := []byte(lorem.Sentence(10, 50))
	signature := sign(acceptorPrivKey, data)
	return AcceptedPayload{
		data:         data,
		acceptorAddr: acceptorPubKey,
		signature:    signature,
	}
}

func compareAcceptedPayload(acceptedPayload1, acceptedPayload2 AcceptedPayload, t *testing.T)  {
	if bytes.Compare(acceptedPayload1.data, acceptedPayload2.data) != 0 {
		t.Errorf("Corrupted data")
	}
	if bytes.Compare(acceptedPayload1.acceptorAddr, acceptedPayload2.acceptorAddr) != 0 {
		t.Errorf("Corrupted acceptor address")
	}
	if bytes.Compare(acceptedPayload1.signature, acceptedPayload2.signature) != 0 {
		t.Errorf("Corrupted signature")
	}
}

// ------------------------------------------------------------------------------------------------------------------- //
// TESTING UTILITIES

func mockState(mockData, mockValidation, mockPayload bool) *State {
	if !mockData {
		mockValidation, mockPayload = false, false
	} else if !mockValidation {
		mockPayload = false
	}
	state := NewState()
	var versionIndex int
	if mockData {
		for zpkIndex, dataIndex := range zpkToData {
			for len(state.dataList) <= dataIndex {
				state.addData(makeDescription())
				versionIndex = 0
			}
			if mockValidation {
				state.addValidation(makeValidation(zpks[zpkIndex]), dataIndex)
			}
			if mockPayload {
				state.addPayload(makePayload(zpks[zpkIndex]), dataIndex, versionIndex)
				versionIndex++
			}
		}
	}
	return state
}

func dataLength(state *State, dataIndex int) (dataListLength, versionListLength int) {
	dataListLength = len(state.dataList)
	versionListLength = len(state.dataList[dataIndex].versionList)
	return
}

func dataHash(state *State, dataIndex int) (dataHashL, dataHashR []byte) {
	for i, data := range state.dataList {
		if i < dataIndex {
			dataHashL = append(dataHashL, data.dataHash...)
		} else if i > dataIndex {
			dataHashR = append(dataHashR, data.dataHash...)
		}
	}
	return
}

func versionHash(state *State, dataIndex, versionIndex int) (versionHashL, versionHashR []byte) {
	for i, version := range state.dataList[dataIndex].versionList {
		if i < versionIndex {
			versionHashL = append(versionHashL, version.versionHash...)
		} else if i > versionIndex {
			versionHashR = append(versionHashR, version.versionHash...)
		}
	}
	return
}

func checkLength(list interface{}, validLength int, descriptor string, t *testing.T)  {
	length := reflect.ValueOf(list).Len()
	if length != validLength {
		t.Errorf(descriptor + ": invalid length")
	}
}

func checkEmpty(element empty, descriptor string, t *testing.T) {
	if !element.isEmpty() {
		t.Errorf(descriptor + " invalid: not empty")
	}
}

func checkNil(element interface{}, descriptor string, t *testing.T) {
	if !reflect.ValueOf(element).IsNil() {
		t.Errorf(descriptor + " invalid: not nil")
	}
}

func checkHash(hash, validHash []byte, descriptor string, t *testing.T) {
	if bytes.Compare(hash, validHash) != 0 {
		t.Errorf(descriptor + ": invalid hash")
	}
}
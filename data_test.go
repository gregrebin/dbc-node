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
	requirerPrivKey, requirerPubKey = LoadKeys(requirerPrivKeyFile, requirerPubKeyFile)
	_ = exec.Command("./key.sh", ecParamFile, validatorPrivKeyFile, validatorPubKeyFile).Run()
	validatorPrivKey, validatorPubKey = LoadKeys(validatorPrivKeyFile, validatorPubKeyFile)
	_ = exec.Command("./key.sh", ecParamFile, providerPrivKeyFile, providerPubKeyFile).Run()
	providerPrivKey, providerPubKey = LoadKeys(providerPrivKeyFile, providerPubKeyFile)
	_ = exec.Command("./key.sh", ecParamFile, acceptorPrivKeyFile, acceptorPubKeyFile).Run()
	acceptorPrivKey, acceptorPubKey = LoadKeys(acceptorPrivKeyFile, acceptorPubKeyFile)
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
	checkNil(state.DataList, "Data list", t)
	validHash := sha256.Sum256(nil)
	checkHash(state.StateHash, validHash[:], "Empty state", t)
	checkHash(state.Hash(), validHash[:], "Hash function return", t)
	checkHash(state.StateHash, validHash[:], "Empty state after hash function", t)

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
	initialLength := len(state.DataList)
	otherDataHash, _ := dataHash(state, initialLength)

	description := mockDescription()
	state.AddData(description)
	data := state.DataList[initialLength]

	checkLength(state.DataList, initialLength+1, "Data list", t)
	checkNil(data.VersionList, "Version list", t)
	compareDescription(data.Description, description, t)
	dataHash := sha256.Sum256(description.Hash())
	checkHash(data.Hash(), dataHash[:], "Data", t)
	stateHash := sha256.Sum256(append(otherDataHash, dataHash[:]...))
	checkHash(state.StateHash, stateHash[:], "State", t)
}

func mockDescription() *Description {
	providerInfo := []byte(lorem.Sentence(10, 20))
	dataInfo := []byte(lorem.Sentence(10, 20))
	signature := Sign(requirerPrivKey, append(providerInfo, dataInfo...))
	description := Description{
		ProviderInfo:      providerInfo,
		DataInfo:          dataInfo,
		TrustedValidators: [][]byte{validatorPubKey},
		Acceptor:          acceptorPubKey,
		Requirer:          requirerPubKey,
		Signature:         signature,
	}
	return &description
}

func compareDescription(desc1, desc2 *Description, t *testing.T) {
	if bytes.Compare(desc1.ProviderInfo, desc2.ProviderInfo) != 0 {
		t.Errorf("Corrupted provider info")
	}
	if bytes.Compare(desc1.DataInfo, desc2.DataInfo) != 0 {
		t.Errorf("Corrupted data info")
	}
	if bytes.Compare(desc1.TrustedValidators[0], desc2.TrustedValidators[0]) != 0 {
		t.Errorf("Corrupted trusted validator")
	}
	if bytes.Compare(desc1.Acceptor, desc2.Acceptor) != 0 {
		t.Errorf("Corrupted Acceptor")
	}
	if bytes.Compare(desc1.Requirer, desc2.Requirer) != 0 {
		t.Errorf("Corrupted Requirer")
	}
	if bytes.Compare(desc1.Signature, desc2.Signature) != 0 {
		t.Errorf("Corrupted Signature")
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
}

func checkValidation(state *State, dataIndex int, zpk zpk, t *testing.T) {
	dataLength, versionLength := dataLength(state, dataIndex)
	dataHashL, dataHashR := dataHash(state, dataIndex)
	otherVersionHash, _ := versionHash(state, dataIndex, versionLength) // since we add a new version, we will have only versions at the left

	validation := mockValidation(zpk)
	state.AddValidation(validation, dataIndex)
	data := state.DataList[dataIndex]
	version := state.DataList[dataIndex].VersionList[versionLength]

	checkLength(state.DataList, dataLength, "Data list", t)
	checkLength(state.DataList[dataIndex].VersionList, versionLength+1, "Version list", t)
	checkEmpty(version.AcceptedPayload, "Accepted payload", t)
	checkEmpty(version.Payload, "Payload", t)
	compareValidation(version.Validation, validation, t)
	emptyAcceptedPayload := AcceptedPayload{}
	emptyPayload := Payload{}
	versionHash := sha256.Sum256(append(append(emptyAcceptedPayload.Hash(), emptyPayload.Hash()...), validation.Hash()...))
	checkHash(version.Hash(), versionHash[:], "Version", t)
	dataHash := sha256.Sum256(append(append(data.Description.Hash(), otherVersionHash...), versionHash[:]...))
	checkHash(data.Hash(), dataHash[:], "Data", t)
	stateHash := sha256.Sum256(append(append(dataHashL, dataHash[:]...), dataHashR...))
	checkHash(state.StateHash, stateHash[:], "State", t)
}

func mockValidation(zpk zpk) *Validation {
	validationInfo := zpk.info
	signature := Sign(validatorPrivKey, validationInfo[:])
	validation := Validation{
		Info:          validationInfo[:],
		ValidatorAddr: validatorPubKey,
		Signature:     signature,
	}
	return &validation
}

func compareValidation(val1, val2 *Validation, t *testing.T) {
	if bytes.Compare(val1.Info, val2.Info) != 0 {
		t.Errorf("Corrupted info")
	}
	if bytes.Compare(val1.ValidatorAddr, val2.ValidatorAddr) != 0 {
		t.Errorf("Corrupted validator adddress")
	}
	if bytes.Compare(val1.Signature, val2.Signature) != 0 {
		t.Errorf("Corrupted Signature")
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
	initialVersion := state.DataList[dataIndex].VersionList[versionIndex]

	payload := mockPayload(zpk)
	state.AddPayload(payload, dataIndex, versionIndex)
	data := state.DataList[dataIndex]
	version := state.DataList[dataIndex].VersionList[versionIndex]

	checkLength(state.DataList, dataLength, "Data list", t)
	checkLength(state.DataList[dataIndex].VersionList, versionLength, "Version list", t)
	checkEmpty(version.AcceptedPayload, "Accepted payload", t)
	comparePayload(version.Payload, payload, t)
	compareValidation(version.Validation, initialVersion.Validation, t)
	emptyAcceptedPayload := AcceptedPayload{}
	versionHash := sha256.Sum256(append(append(emptyAcceptedPayload.Hash(), payload.Hash()...), initialVersion.Validation.Hash()...))
	checkHash(version.Hash(), versionHash[:], "Version", t)
	dataHash := sha256.Sum256(append(append(append(data.Description.Hash(), versionHashL...), versionHash[:]...), versionHashR...))
	checkHash(data.Hash(), dataHash[:], "Data", t)
	stateHash := sha256.Sum256(append(append(dataHashL, dataHash[:]...), dataHashR...))
	checkHash(state.StateHash, stateHash[:], "State", t)
}

func mockPayload(zpk zpk) *Payload {
	data := []byte(lorem.Sentence(10, 50))
	signature := Sign(providerPrivKey, append(data, zpk.proof...))
	payload := Payload{
		Data:         data,
		Proof:        zpk.proof,
		ProviderAddr: providerPubKey,
		Signature:    signature,
	}
	return &payload
}

func comparePayload(payload1, payload2 *Payload, t *testing.T) {
	if bytes.Compare(payload1.Data, payload2.Data) != 0 {
		t.Errorf("Corrupted data")
	}
	if bytes.Compare(payload1.Proof, payload2.Proof) != 0 {
		t.Errorf("Corrupted proof")
	}
	if bytes.Compare(payload1.ProviderAddr, payload2.ProviderAddr) != 0 {
		t.Errorf("Corrupted provider address")
	}
	if bytes.Compare(payload1.Signature, payload2.Signature) != 0 {
		t.Errorf("Corrupted Signature")
	}
}

// ------------------------------------------------------------------------------------------------------------------- //
// ACCEPT PAYLOAD

func TestAcceptPayload(t *testing.T) {
	state := mockState(true, true, true)
	for dataIndex, data := range state.DataList {
		for versionIndex := range data.VersionList {
			checkAcceptedPayload(state, dataIndex, versionIndex, t)
		}
	}
}

func checkAcceptedPayload(state *State, dataIndex, versionIndex int, t *testing.T) {
	dataLength, versionLength := dataLength(state, dataIndex)
	dataHashL, dataHashR := dataHash(state, dataIndex)
	versionHashL, versionHashR := versionHash(state, dataIndex, versionIndex)
	initialVersion := state.DataList[dataIndex].VersionList[versionIndex]

	acceptedPayload := mockAcceptedPayload()
	state.AcceptPayload(acceptedPayload, dataIndex, versionIndex)
	data := state.DataList[dataIndex]
	version := state.DataList[dataIndex].VersionList[versionIndex]

	checkLength(state.DataList, dataLength, "Data list", t)
	checkLength(state.DataList[dataIndex].VersionList, versionLength, "Version list", t)
	compareAcceptedPayload(version.AcceptedPayload, acceptedPayload, t)
	comparePayload(version.Payload, initialVersion.Payload, t)
	compareValidation(version.Validation, initialVersion.Validation, t)
	versionHash := sha256.Sum256(append(append(acceptedPayload.Hash(), initialVersion.Payload.Hash()...), initialVersion.Validation.Hash()...))
	checkHash(version.Hash(), versionHash[:], "Version", t)
	dataHash := sha256.Sum256(append(append(append(data.Description.Hash(), versionHashL...), versionHash[:]...), versionHashR...))
	checkHash(data.Hash(), dataHash[:], "Data", t)
	stateHash := sha256.Sum256(append(append(dataHashL, dataHash[:]...), dataHashR...))
	checkHash(state.StateHash, stateHash[:], "State", t)
}

func mockAcceptedPayload() *AcceptedPayload {
	data := []byte(lorem.Sentence(10, 50))
	signature := Sign(acceptorPrivKey, data)
	acceptedPayload := AcceptedPayload{
		Data:         data,
		AcceptorAddr: acceptorPubKey,
		Signature:    signature,
	}
	return &acceptedPayload
}

func compareAcceptedPayload(acceptedPayload1, acceptedPayload2 *AcceptedPayload, t *testing.T) {
	if bytes.Compare(acceptedPayload1.Data, acceptedPayload2.Data) != 0 {
		t.Errorf("Corrupted data")
	}
	if bytes.Compare(acceptedPayload1.AcceptorAddr, acceptedPayload2.AcceptorAddr) != 0 {
		t.Errorf("Corrupted Acceptor address")
	}
	if bytes.Compare(acceptedPayload1.Signature, acceptedPayload2.Signature) != 0 {
		t.Errorf("Corrupted Signature")
	}
}

// ------------------------------------------------------------------------------------------------------------------- //
// TESTING UTILITIES

func mockState(data, validation, payload bool) *State {
	if !data {
		validation, payload = false, false
	} else if !validation {
		payload = false
	}
	state := NewState()
	var versionIndex int
	if data {
		for zpkIndex, dataIndex := range zpkToData {
			for len(state.DataList) <= dataIndex {
				state.AddData(mockDescription())
				versionIndex = 0
			}
			if validation {
				state.AddValidation(mockValidation(zpks[zpkIndex]), dataIndex)
			}
			if payload {
				state.AddPayload(mockPayload(zpks[zpkIndex]), dataIndex, versionIndex)
				versionIndex++
			}
		}
	}
	return state
}

func dataLength(state *State, dataIndex int) (dataListLength, versionListLength int) {
	dataListLength = len(state.DataList)
	versionListLength = len(state.DataList[dataIndex].VersionList)
	return
}

func dataHash(state *State, dataIndex int) (dataHashL, dataHashR []byte) {
	for i, data := range state.DataList {
		if i < dataIndex {
			dataHashL = append(dataHashL, data.Hash()...)
		} else if i > dataIndex {
			dataHashR = append(dataHashR, data.Hash()...)
		}
	}
	return
}

func versionHash(state *State, dataIndex, versionIndex int) (versionHashL, versionHashR []byte) {
	for i, version := range state.DataList[dataIndex].VersionList {
		if i < versionIndex {
			versionHashL = append(versionHashL, version.Hash()...)
		} else if i > versionIndex {
			versionHashR = append(versionHashR, version.Hash()...)
		}
	}
	return
}

func checkLength(list interface{}, validLength int, descriptor string, t *testing.T) {
	length := reflect.ValueOf(list).Len()
	if length != validLength {
		t.Errorf(descriptor + ": invalid length")
	}
}

func checkEmpty(element Empty, descriptor string, t *testing.T) {
	if !element.IsEmpty() {
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

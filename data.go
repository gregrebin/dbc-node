package main

/*
State contains a list of Data and the hash
Data contains Description and a list of versions and the hash
Each version contains Validation, Payload and acceptedPayload
A validation is just 3 arrays of bytes

Each new block a new state gets generated
*/

import (
	"bytes"
	"crypto/sha256"
	"github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/node"
	tmdb "github.com/tendermint/tm-db"
)

type merkleNode interface {
	hash() []byte
}
type empty interface {
	isEmpty() bool
}

type State struct {
	stateHash []byte
	dataList  []Data
}
func (state *State) hash() []byte {
	var sum []byte
	for i := range state.dataList {
		sum = append(sum, state.dataList[i].hash()...)
	}
	hash := sha256.Sum256(sum)
	state.stateHash = hash[:]
	return state.stateHash
}

/*
	Represent data of some type with a single owner
	Contains a general description of itself defined at creation by the owner
	and many versions that can be added later by people decided by the owner
	Can be hashed by adding hashes of description and every version, and hashing the result
*/
type Data struct {
	dataHash    []byte
	description Description
	versionList []Version
}
func (data *Data) hash() []byte {
	sum := data.description.hash()
	for i := range data.versionList {
		sum = append(sum, data.versionList[i].hash()...)
	}
	hash := sha256.Sum256(sum)
	data.dataHash = hash[:]
	return data.dataHash
}

/*	Describes the type of data with dataInfo and the expected data provider with providerInfo, both generic, anything will be accepted.
	Defines data owner with requirer, must be a valid secp256k1 public key. The signature must be a valid signature
	of (description.providerInfo + description.dataInfo) for the given key.
	Defines a list of trusted validators, each must be a secp256k1 public key, they are expected to
	define valid data providers conforming to providerInfo, they could be a government entity, some other trusted entity,
	the owner himself or can be left blank if any data from any provider is accepted.
	Defines an acceptor, must be a secp256k1 public key, he is responsible for manually checking the data and
	confirming its conformance to the data requested in dataInfo
	Contains only arrays of bytes. Can be hashed by adding hashes of every field and hashing the result. */
type Description struct {
	providerInfo      []byte
	dataInfo          []byte
	trustedValidators [][]byte
	acceptor          []byte
	requirer          []byte
	signature         []byte
}
func (description *Description) hash() []byte {
	sum := append(description.providerInfo, description.dataInfo...)
	for _, validator := range description.trustedValidators {
		sum = append(sum, validator...)
	}
	sum = append(sum, description.acceptor...)
	sum = append(sum, description.requirer...)
	sum = append(sum, description.signature...)
	hash := sha256.Sum256(sum)
	return hash[:]
}


/*	A version of data ... */
type Version struct {
	versionHash     []byte
	acceptedPayload AcceptedPayload
	payload         Payload
	validation      Validation
}
func (version *Version) hash() []byte {
	sum := append(version.acceptedPayload.hash(), version.payload.hash()...)
	sum = append(sum, version.validation.hash()...)
	hash := sha256.Sum256(sum)
	version.versionHash = hash[:]
	return version.versionHash
}

/*	Contains the actual data encrypted with requirer secp256k1 public key. The data is considered provided and verified.
	The acceptor address must be the secp256k1 public key provided (description.acceptor).
	The signature must be a valid signature of (acceptedPayload.data) for the given key
	Contains only arrays of bytes. Can be hashed by adding hashes of every field and hashing the result.
	Can be empty / uninitialized. */
type AcceptedPayload struct {
	data         []byte // encrypted with requirer, when decrypted by requirer should be encrypted with acceptorAddr to check if it's the same as in payload
	acceptorAddr []byte // public key representing acceptor address, should be the same as in the description
	signature    []byte // confirming acceptorAddr
}
func (acceptedPayload *AcceptedPayload) hash() []byte {
	sum := append(acceptedPayload.data, acceptedPayload.acceptorAddr...)
	sum = append(sum, acceptedPayload.signature...)
	hash :=  sha256.Sum256(sum)
	return hash[:]
}
func (acceptedPayload *AcceptedPayload) isEmpty() bool {
	return acceptedPayload.data == nil && acceptedPayload.acceptorAddr == nil && acceptedPayload.signature == nil
}

/*	Contains the actual data encrypted with acceptor secp256k1 public key. The data is considered provided, but not verified.
	The zero knowledge proof must be an arbitrary info or seed known to both validator and provider hashed n-1 times,
	if (hash(payload.proof) != validation.info) then the payload wont be accepted!
	The provider address can be any secp256k1 public key.
	The signature must be a valid signature of (payload.data + payload.proof) for the given key.
	Contains only arrays of bytes. Can be hashed by adding hashes of every field and hashing the result.
	Can be empty / uninitialized. */
type Payload struct {
	data         []byte
	proof        []byte
	providerAddr []byte
	signature    []byte
}
func (payload *Payload) hash() []byte {
	sum := append(payload.data, payload.proof...)
	sum = append(sum, payload.providerAddr...)
	sum = append(sum, payload.signature...)
	hash := sha256.Sum256(sum)
	return hash[:]
}
func (payload *Payload) isEmpty() bool {
	return payload.data == nil && payload.proof == nil && payload.providerAddr == nil && payload.signature == nil
}

/*	An arbitrary info or seed known to both validator and provider hashed n times, could be an official ID number,
	is needed for zero knowledge proof of validation identity.
	The validator address must be one of secp256k1 public keys provided in (description.trustedValidators).
	The signature must be a valid signature of (validation.info) for the given key.
	Contains only arrays of bytes. Can be hashed by adding hashes of every field and hashing the result. */
type Validation struct {
	info          []byte
	validatorAddr []byte
	signature     []byte
}
func (validation *Validation) hash() []byte {
	sum := append(validation.info, validation.validatorAddr...)
	sum = append(sum, validation.signature...)
	hash := sha256.Sum256(sum)
	return hash[:]
}

func NewState() *State { // called every new block
	state := &State{}
	state.hash()
	return state
}

func (state *State) addData(description Description) { // called at requireTx
	id := append(description.providerInfo, description.dataInfo...)
	isSigned := verify(description.requirer, id, description.signature)
	if isSigned {
		data := Data{description: description}
		state.dataList = append(state.dataList, data)
		state.hash()
	}
}

func (state *State) addValidation(validation Validation, dataIndex int) { // called at validateTx
	isTrusted := false
	for _, trusted := range state.dataList[dataIndex].description.trustedValidators {
		if bytes.Compare(validation.validatorAddr, trusted) == 0 {
			isTrusted = true
		}
	}
	isSigned := verify(validation.validatorAddr, validation.info, validation.signature)
	if isTrusted && isSigned {
		version := Version{validation: validation}
		state.dataList[dataIndex].versionList = append(state.dataList[dataIndex].versionList, version)
		state.hash()
	}
}

func (state *State) addPayload(payload Payload, dataIndex int, versionIndex int) { //called at provideTx
	// TODO: check if a payload already exists
	isProved := false
	proof := sha256.Sum256(payload.proof)
	info := state.dataList[dataIndex].versionList[versionIndex].validation.info
	if bytes.Compare(proof[:], info) == 0 {isProved = true}
	isSigned := verify(payload.providerAddr, append(payload.data, payload.proof...), payload.signature)
	if isProved && isSigned {
		state.dataList[dataIndex].versionList[versionIndex].payload = payload
		state.hash()
	}
}

func (state *State) acceptPayload(acceptedPayload AcceptedPayload, dataIndex int, versionIndex int) { //called at acceptTx
	// TODO: check if payload acceptance already exists
	isAcceptor := bytes.Compare(acceptedPayload.acceptorAddr, state.dataList[dataIndex].description.acceptor) == 0
	isSigned := verify(acceptedPayload.acceptorAddr, acceptedPayload.data, acceptedPayload.signature)
	if isAcceptor && isSigned {
		state.dataList[dataIndex].versionList[versionIndex].acceptedPayload = acceptedPayload
		state.hash()
	}
}



func NewPersistentData(config *config.Config) tmdb.DB {
	database, _ := node.DefaultDBProvider(
		&node.DBContext{
			ID:     "dbc-node",
			Config: config,
		})
	return database
}

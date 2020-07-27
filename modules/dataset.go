package modules

/*
Dataset contains a list of Data and the hash
Data contains Description and a list of versions and the hash
Each version contains Validation, Payload and acceptedPayload
A validation is just 3 arrays of bytes

Each new block a new dataset gets generated
*/

import (
	"bytes"
	"crypto/sha256"
	"dbc-node/crypto"
)

type Empty interface {
	IsEmpty() bool
}

// ------------------------------------------------------------------------------------------------------------------- //
// DATASET

type Dataset struct {
	DataList []Data
}

func NewDataset(old *Dataset) *Dataset { // called every new block
	dataset := &Dataset{}
	for _, oldData := range old.DataList {
		data := Data{
			Description: oldData.Description,
		}
		for _, oldVersion := range oldData.VersionList {
			version := Version{
				AcceptedPayload: oldVersion.AcceptedPayload,
				Payload:         oldVersion.Payload,
				Validation:      oldVersion.Validation,
			}
			data.VersionList = append(data.VersionList, version)
		}
		dataset.DataList = append(dataset.DataList, data)
	}
	return dataset
}

func (dataset *Dataset) Hash() []byte {
	var sum []byte
	if dataset == nil {
		return sum
	}
	for i := range dataset.DataList {
		sum = append(sum, dataset.DataList[i].Hash()...)
	}
	hash := sha256.Sum256(sum)
	return hash[:]
}
func (dataset *Dataset) AddData(description *Description) { // called at requireTx
	id := append(description.ProviderInfo, description.DataInfo...)
	isSigned := crypto.Verify(description.Requirer, id, description.Signature)
	if isSigned {
		data := Data{Description: description}
		dataset.DataList = append(dataset.DataList, data)
		dataset.Hash()
	}
}
func (dataset *Dataset) AddValidation(validation *Validation, dataIndex int) { // called at validateTx
	isTrusted := false
	for _, trusted := range dataset.DataList[dataIndex].Description.TrustedValidators {
		if bytes.Compare(validation.ValidatorAddr, trusted) == 0 {
			isTrusted = true
		}
	}
	isSigned := crypto.Verify(validation.ValidatorAddr, validation.Info, validation.Signature)
	if isTrusted && isSigned {
		version := Version{Validation: validation, Payload: &Payload{}, AcceptedPayload: &AcceptedPayload{}}
		dataset.DataList[dataIndex].VersionList = append(dataset.DataList[dataIndex].VersionList, version)
		dataset.Hash()
	}
}
func (dataset *Dataset) AddPayload(payload *Payload, dataIndex int, versionIndex int) { //called at provideTx
	isProved := false
	proof := sha256.Sum256(payload.Proof)
	info := dataset.DataList[dataIndex].VersionList[versionIndex].Validation.Info
	if bytes.Compare(proof[:], info) == 0 {
		isProved = true
	}
	isSigned := crypto.Verify(payload.ProviderAddr, append(payload.Data, payload.Proof...), payload.Signature)
	isEmpty := dataset.DataList[dataIndex].VersionList[versionIndex].Payload.IsEmpty()
	if isProved && isSigned && isEmpty {
		dataset.DataList[dataIndex].VersionList[versionIndex].Payload = payload
		dataset.Hash()
	}
}
func (dataset *Dataset) AcceptPayload(acceptedPayload *AcceptedPayload, dataIndex int, versionIndex int) { //called at acceptTx
	isAcceptor := bytes.Compare(acceptedPayload.AcceptorAddr, dataset.DataList[dataIndex].Description.Acceptor) == 0
	isSigned := crypto.Verify(acceptedPayload.AcceptorAddr, acceptedPayload.Data, acceptedPayload.Signature)
	isEmpty := dataset.DataList[dataIndex].VersionList[versionIndex].AcceptedPayload.IsEmpty()
	if isAcceptor && isSigned && isEmpty {
		dataset.DataList[dataIndex].VersionList[versionIndex].AcceptedPayload = acceptedPayload
		dataset.Hash()
	}
}

// ------------------------------------------------------------------------------------------------------------------- //
// DATA

/*
	Represent data of some type with a single owner
	Contains a general description of itself defined at creation by the owner
	and many versions that can be added later by people decided by the owner
	Can be hashed by adding hashes of description and every version, and hashing the result
*/
type Data struct {
	Description *Description
	VersionList []Version
}

func (data *Data) Hash() []byte {
	sum := data.Description.Hash()
	for i := range data.VersionList {
		sum = append(sum, data.VersionList[i].Hash()...)
	}
	hash := sha256.Sum256(sum)
	return hash[:]
}

// ------------------------------------------------------------------------------------------------------------------- //
// DESCRIPTION

/*	Describes the type of data with DataInfo and the expected data provider with ProviderInfo, both generic, anything will be accepted.
	Defines data owner with Requirer, must be a valid secp256k1 public key. The Signature must be a valid Signature
	of (description.ProviderInfo + description.DataInfo) for the given key.
	Defines a list of trusted validators, each must be a secp256k1 public key, they are expected to
	define valid data providers conforming to ProviderInfo, they could be a government entity, some other trusted entity,
	the owner himself or can be left blank if any data from any provider is accepted.
	Defines an Acceptor, must be a secp256k1 public key, he is responsible for manually checking the data and
	confirming its conformance to the data requested in DataInfo
	Contains only arrays of bytes. Can be hashed by adding hashes of every field and hashing the result. */
type Description struct {
	ProviderInfo      []byte
	DataInfo          []byte
	TrustedValidators [][]byte
	Acceptor          []byte
	Requirer          []byte
	Signature         []byte
}

func (description *Description) Hash() []byte {
	sum := append(description.ProviderInfo, description.DataInfo...)
	for _, validator := range description.TrustedValidators {
		sum = append(sum, validator...)
	}
	sum = append(sum, description.Acceptor...)
	sum = append(sum, description.Requirer...)
	sum = append(sum, description.Signature...)
	hash := sha256.Sum256(sum)
	return hash[:]
}

// ------------------------------------------------------------------------------------------------------------------- //
// VERSION

/*	A version of data ... */
type Version struct {
	AcceptedPayload *AcceptedPayload
	Payload         *Payload
	Validation      *Validation
}

func (version *Version) Hash() []byte {
	sum := append(version.AcceptedPayload.Hash(), version.Payload.Hash()...)
	sum = append(sum, version.Validation.Hash()...)
	hash := sha256.Sum256(sum)
	return hash[:]
}

// ------------------------------------------------------------------------------------------------------------------- //
// ACCEPTED PAYLOAD

/*	Contains the actual data encrypted with Requirer secp256k1 public key. The data is considered provided and verified.
	The Acceptor address must be the secp256k1 public key provided (description.Acceptor).
	The Signature must be a valid Signature of (acceptedPayload.data) for the given key
	Contains only arrays of bytes. Can be hashed by adding hashes of every field and hashing the result.
	Can be empty / uninitialized. */
type AcceptedPayload struct {
	Data         []byte // encrypted with Requirer, when decrypted by Requirer should be encrypted with acceptorAddr to check if it's the same as in payload
	AcceptorAddr []byte // public key representing Acceptor address, should be the same as in the description
	Signature    []byte // confirming acceptorAddr
}

func (acceptedPayload *AcceptedPayload) Hash() []byte {
	sum := append(acceptedPayload.Data, acceptedPayload.AcceptorAddr...)
	sum = append(sum, acceptedPayload.Signature...)
	hash := sha256.Sum256(sum)
	return hash[:]
}
func (acceptedPayload *AcceptedPayload) IsEmpty() bool {
	return acceptedPayload.Data == nil && acceptedPayload.AcceptorAddr == nil && acceptedPayload.Signature == nil
}

// ------------------------------------------------------------------------------------------------------------------- //
// PAYLOAD

/*	Contains the actual data encrypted with Acceptor secp256k1 public key. The data is considered provided, but not verified.
	The zero knowledge proof must be an arbitrary info or seed known to both validator and provider hashed n-1 times,
	if (hash(payload.proof) != validation.info) then the payload wont be accepted!
	The provider address can be any secp256k1 public key.
	The Signature must be a valid Signature of (payload.data + payload.proof) for the given key.
	Contains only arrays of bytes. Can be hashed by adding hashes of every field and hashing the result.
	Can be empty / uninitialized. */
type Payload struct {
	Data         []byte
	Proof        []byte
	ProviderAddr []byte
	Signature    []byte
}

func (payload *Payload) Hash() []byte {
	sum := append(payload.Data, payload.Proof...)
	sum = append(sum, payload.ProviderAddr...)
	sum = append(sum, payload.Signature...)
	hash := sha256.Sum256(sum)
	return hash[:]
}
func (payload *Payload) IsEmpty() bool {
	return payload.Data == nil && payload.Proof == nil && payload.ProviderAddr == nil && payload.Signature == nil
}

// ------------------------------------------------------------------------------------------------------------------- //
// VALIDATION

/*	An arbitrary info or seed known to both validator and provider hashed n times, could be an official ID number,
	is needed for zero knowledge proof of validation identity.
	The validator address must be one of secp256k1 public keys provided in (description.TrustedValidators).
	The Signature must be a valid Signature of (validation.info) for the given key.
	Contains only arrays of bytes. Can be hashed by adding hashes of every field and hashing the result. */
type Validation struct {
	Info          []byte
	ValidatorAddr []byte
	Signature     []byte
}

func (validation *Validation) Hash() []byte {
	sum := append(validation.Info, validation.ValidatorAddr...)
	sum = append(sum, validation.Signature...)
	hash := sha256.Sum256(sum)
	return hash[:]
}
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
	"errors"
)

type Empty interface {
	IsEmpty() bool
}

// ------------------------------------------------------------------------------------------------------------------- //
// DATASET

type Dataset struct {
	DataList []Data
	balance  *Balance
}

func NewDataset(old *Dataset, balance *Balance) *Dataset { // called every new block
	dataset := &Dataset{balance: balance}
	for _, oldData := range old.DataList {
		data := Data{
			Description: oldData.Description,
			Reward:      oldData.Reward,
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

func (dataset *Dataset) AddData(description *Description) error { // called at requireTx
	if err := description.check(); err != nil {
		return err
	}
	if !description.isSigned() {
		return errors.New("invalid description signature")
	}
	err, index := dataset.balance.AddReward(description.reward())
	if err != nil {
		return err
	}
	data := Data{Description: description, Reward: index}
	dataset.DataList = append(dataset.DataList, data)
	dataset.Hash()
	return nil
}

func (dataset *Dataset) AddValidation(validation *Validation, dataIndex int) error { // called at validateTx
	if err := validation.check(); err != nil {
		return err
	}
	if !validation.isSigned() {
		return errors.New("invalid validation signature")
	}
	data := &dataset.DataList[dataIndex]
	if !data.isValidator(validation) {
		return errors.New("validator not approved")
	}
	if !data.inRange() {
		return errors.New("reached max versions limit")
	}
	version := Version{Validation: validation, Payload: &Payload{}, AcceptedPayload: &AcceptedPayload{}}
	data.VersionList = append(data.VersionList, version)
	dataset.Hash()
	return nil
}

func (dataset *Dataset) AddPayload(payload *Payload, dataIndex int, versionIndex int) error { //called at provideTx
	if err := payload.check(); err != nil {
		return err
	}
	if !payload.isSigned() {
		return errors.New("invalid payload signature")
	}
	version := &dataset.DataList[dataIndex].VersionList[versionIndex]
	if !version.prove(payload) {
		return errors.New("invalid payload proof")
	}
	if !version.Payload.IsEmpty() {
		return errors.New("payload already exists")
	}
	version.Payload = payload
	dataset.Hash()
	return nil
}

func (dataset *Dataset) AcceptPayload(acceptedPayload *AcceptedPayload, dataIndex int, versionIndex int) error { //called at acceptTx
	if err := acceptedPayload.check(); err != nil {
		return err
	}
	if !acceptedPayload.isSigned() {
		return errors.New("invalid accepted payload signature")
	}
	data := &dataset.DataList[dataIndex]
	if !data.isAcceptor(acceptedPayload) {
		return errors.New("acceptor not approved")
	}
	version := &data.VersionList[versionIndex]
	if !version.AcceptedPayload.IsEmpty() {
		return errors.New("accepted payload already exists")
	}
	if err := dataset.balance.ConfirmReward(version.rewardConfirm(), data.Reward); err != nil {
		return err
	}
	version.AcceptedPayload = acceptedPayload
	dataset.Hash()
	return nil
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
	Reward      int
}

func (data *Data) Hash() []byte {
	sum := data.Description.Hash()
	for i := range data.VersionList {
		sum = append(sum, data.VersionList[i].Hash()...)
	}
	hash := sha256.Sum256(sum)
	return hash[:]
}

func (data *Data) isValidator(validation *Validation) bool {
	return bytes.Compare(validation.ValidatorAddr, data.Description.Validator) == 0
}

func (data *Data) isAcceptor(acceptedPayload *AcceptedPayload) bool {
	return bytes.Compare(acceptedPayload.AcceptorAddr, data.Description.Acceptor) == 0
}

func (data *Data) inRange() bool {
	return int64(len(data.VersionList)) < data.Description.MaxVersions
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
	Contains only arrays of bytes (amounts don't count). Can be hashed by adding hashes of every field and hashing the result. */
type Description struct {
	ProviderInfo    []byte
	DataInfo        []byte
	Validator       []byte
	Acceptor        []byte
	Requirer        []byte
	ValidatorAmount int64
	ProviderAmount  int64
	AcceptorAmount  int64
	MaxVersions     int64
	Signature       []byte
}

func (description *Description) Hash() []byte {
	sum := append(description.ProviderInfo, description.DataInfo...)
	sum = append(sum, description.Validator...)
	sum = append(sum, description.Acceptor...)
	sum = append(sum, description.Requirer...)
	sum = append(sum, description.Signature...)
	hash := sha256.Sum256(sum)
	return hash[:]
}

func (description Description) check() error {
	if err := crypto.CheckPubKey(description.Requirer); err != nil {
		return err
	} else if err := crypto.CheckPubKey(description.Validator); err != nil {
		return err
	} else if err := crypto.CheckPubKey(description.Acceptor); err != nil {
		return nil
	} else if description.ValidatorAmount < 0 {
		return errors.New("negative validator amount")
	} else if description.ProviderAmount < 0 {
		return errors.New("negative provider amount")
	} else if description.AcceptorAmount < 0 {
		return errors.New("negative acceptor amount")
	} else {
		return nil
	}
}

func (description *Description) isSigned() bool {
	id := append(description.ProviderInfo, description.DataInfo...)
	return crypto.Verify(description.Requirer, id, description.Signature)
}

func (description *Description) reward() Reward {
	return Reward{
		Info: &RewardInfo{
			Requirer:        description.Requirer,
			Validator:       description.Validator,
			Acceptor:        description.Acceptor,
			ValidatorAmount: description.ValidatorAmount,
			ProviderAmount:  description.ProviderAmount,
			AcceptorAmount:  description.AcceptorAmount,
			MaxConfirms:     description.MaxVersions,
		},
		State: RewardOpen,
	}
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

func (version *Version) prove(payload *Payload) bool {
	proof := sha256.Sum256(payload.Proof)
	info := version.Validation.Info
	return bytes.Compare(proof[:], info) == 0
}

func (version *Version) rewardConfirm() *RewardConfirm {
	return &RewardConfirm{
		Provider: version.Payload.ProviderAddr,
	}
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

func (acceptedPayload *AcceptedPayload) check() error {
	return crypto.CheckPubKey(acceptedPayload.AcceptorAddr)
}

func (acceptedPayload *AcceptedPayload) isSigned() bool {
	return crypto.Verify(acceptedPayload.AcceptorAddr, acceptedPayload.Data, acceptedPayload.Signature)
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

func (payload *Payload) check() error {
	return crypto.CheckPubKey(payload.ProviderAddr)
}

func (payload *Payload) isSigned() bool {
	return crypto.Verify(payload.ProviderAddr, append(payload.Data, payload.Proof...), payload.Signature)
}

// ------------------------------------------------------------------------------------------------------------------- //
// VALIDATION

/*	An arbitrary info or seed known to both validator and provider hashed n times, could be an official ID number,
	is needed for zero knowledge proof of validation identity.
	The validator address must be one of secp256k1 public keys provided in (description.Validator).
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

func (validation *Validation) check() error {
	return crypto.CheckPubKey(validation.ValidatorAddr)
}

func (validation *Validation) isSigned() bool {
	return crypto.Verify(validation.ValidatorAddr, validation.Info, validation.Signature)
}

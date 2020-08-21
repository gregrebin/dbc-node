package modules

import (
	"crypto/sha256"
	"dbc-node/crypto"
	"encoding/hex"
	"errors"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/types"
	"strconv"
)

const (
	CoinName   = "DBCC"
	MinUnit    = "SATS"
	DbccSats   = 100000000
	SatsSupply = types.MaxTotalVotingPower
	DbccSupply = SatsSupply / DbccSats
	TxFee      = 7700000
)

func ToSats(dbcc int64) int64 { return dbcc * DbccSats }

// ------------------------------------------------------------------------------------------------------------------- //
// BALANCE

type Balance struct {
	Users      map[string]int64
	Validators map[string]int64
	ValChanges map[string]int64
	ValAddr    map[[20]byte][32]byte
	Transfers  []*Transfer
	Stakes     []*Stake
	Rewards    []Reward
	Fees       []*Fee
}

func NewBalance(oldBalance *Balance) *Balance {
	balance := &Balance{
		Users:      make(map[string]int64),
		Validators: make(map[string]int64),
		ValChanges: make(map[string]int64),
		ValAddr:    make(map[[20]byte][32]byte),
	}
	for user, value := range oldBalance.Users {
		balance.Users[user] = value
	}
	for validator, value := range oldBalance.Validators {
		balance.Validators[validator] = value
	}
	for validator, _ := range oldBalance.Validators {
		valBytes, _ := hex.DecodeString(validator)
		balance.registerValAddr(valBytes)
	}
	for _, transfer := range oldBalance.Transfers {
		balance.Transfers = append(balance.Transfers, transfer)
	}
	for _, stake := range oldBalance.Stakes {
		balance.Stakes = append(balance.Stakes, stake)
	}
	for _, oldReward := range oldBalance.Rewards {
		reward := Reward{
			Info:  oldReward.Info,
			State: oldReward.State,
		}
		for _, confirm := range oldReward.Confirms {
			reward.Confirms = append(reward.Confirms, confirm)
		}
		balance.Rewards = append(balance.Rewards, reward)
	}
	for _, fee := range oldBalance.Fees {
		balance.Fees = append(balance.Fees, fee)
	}
	return balance
}

func (balance *Balance) Hash() []byte {
	var sum []byte
	if balance == nil {
		return sum
	}
	for _, transfer := range balance.Transfers {
		sum = append(sum, transfer.Hash()...)
	}
	for _, stake := range balance.Stakes {
		sum = append(sum, stake.Hash()...)
	}
	for _, reward := range balance.Rewards {
		sum = append(sum, reward.Hash()...)
	}
	for _, fee := range balance.Fees {
		sum = append(sum, fee.Hash()...)
	}
	hash := sha256.Sum256(sum)
	return hash[:]
}

func (balance *Balance) AddTransfer(transfer *Transfer) error {
	if err := transfer.check(); err != nil {
		return err
	}
	if !transfer.isSigned() {
		return errors.New("invalid transfer signature")
	}
	if !balance.hasBalance(transfer.Sender, transfer.Amount) {
		return errors.New("insufficient balance")
	}
	balance.Transfers = append(balance.Transfers, transfer)
	sender := hex.EncodeToString(transfer.Sender)
	balance.Users[sender] -= transfer.Amount
	receiver := hex.EncodeToString(transfer.Receiver)
	balance.Users[receiver] += transfer.Amount
	return nil
}

func (balance *Balance) AddStake(stake *Stake) error {
	if err := stake.check(); err != nil {
		return err
	}
	if isSigned := stake.isSigned(); !isSigned {
		return errors.New("invalid stake signature")
	}
	if stake.Amount >= 0 && !balance.hasBalance(stake.User, stake.Amount) {
		return errors.New("insufficient balance")
	}
	if stake.Amount < 0 && !balance.hasStake(stake.Validator, -stake.Amount) {
		return errors.New("insufficient stake")
	}
	balance.Stakes = append(balance.Stakes, stake)
	user := hex.EncodeToString(stake.User)
	balance.Users[user] -= stake.Amount
	validator := hex.EncodeToString(stake.Validator)
	balance.Validators[validator] += stake.Amount
	balance.ValChanges[validator] += stake.Amount
	balance.registerValAddr(stake.Validator)
	return nil
}

func (balance *Balance) AddReward(reward Reward) (error, int) {
	if !balance.hasBalance(reward.Info.Requirer, reward.totalAmount()) {
		return errors.New("insufficient balance"), 0
	}
	balance.Rewards = append(balance.Rewards, reward)
	requirer := hex.EncodeToString(reward.Info.Requirer)
	balance.Users[requirer] -= reward.totalAmount()
	return nil, len(balance.Rewards) - 1
}

func (balance *Balance) ConfirmReward(confirm *RewardConfirm, index int) error {
	reward := &balance.Rewards[index]
	if !reward.inRange() {
		return errors.New("reached max confirms limit or reward is closed")
	}
	balance.Rewards[index].Confirms = append(balance.Rewards[index].Confirms, confirm)
	validator := hex.EncodeToString(reward.Info.Validator)
	balance.Users[validator] += reward.Info.ValidatorAmount
	provider := hex.EncodeToString(confirm.Provider)
	balance.Users[provider] += reward.Info.ProviderAmount
	acceptor := hex.EncodeToString(reward.Info.Acceptor)
	balance.Users[acceptor] += reward.Info.AcceptorAmount
	return nil
}

func (balance *Balance) CloseReward(index int) error {
	reward := &balance.Rewards[index]
	if reward.State == RewardClosed {
		return errors.New("reward closed")
	}
	requirer := hex.EncodeToString(reward.Info.Requirer)
	balance.Users[requirer] += reward.onCloseReturn()
	reward.State = RewardClosed
	return nil
}

func (balance *Balance) AddFee(fee *Fee) error {
	if !balance.hasBalance(fee.User, TxFee) {
		return errors.New("insufficient balance")
	}
	balance.Fees = append(balance.Fees, fee)
	user := hex.EncodeToString(fee.User)
	balance.Users[user] -= TxFee
	validator := hex.EncodeToString(balance.searchValAddr(fee.ValAddr))
	balance.Validators[validator] += TxFee
	balance.ValChanges[validator] += TxFee
	return nil
}

func (balance *Balance) hasBalance(user []byte, amount int64) bool {
	return balance.Users[hex.EncodeToString(user)] >= amount
}

func (balance *Balance) hasStake(validator []byte, amount int64) bool {
	return balance.Validators[hex.EncodeToString(validator)] >= amount
}

func (balance *Balance) registerValAddr(validator []byte) {
	var pubKey ed25519.PubKeyEd25519
	copy(pubKey[:], validator)
	var address [20]byte
	copy(address[:], pubKey.Address())
	balance.ValAddr[address] = pubKey
}

func (balance *Balance) searchValAddr(address []byte) []byte {
	var valAddr [20]byte
	copy(valAddr[:], address)
	validator := balance.ValAddr[valAddr]
	return validator[:]
}

// ------------------------------------------------------------------------------------------------------------------- //
// TRANSFER

type Transfer struct {
	Sender    []byte
	Receiver  []byte
	Amount    int64
	Time      int64
	Signature []byte
}

func (transfer *Transfer) Hash() []byte {
	sum := append(transfer.Sender, transfer.Receiver...)
	sum = append(sum, []byte(strconv.FormatInt(transfer.Amount, 10))...)
	sum = append(sum, []byte(strconv.FormatInt(transfer.Time, 10))...)
	sum = append(sum, transfer.Signature...)
	hash := sha256.Sum256(sum)
	return hash[:]
}

func (transfer *Transfer) check() error {
	if err := crypto.CheckPubKey(transfer.Sender); err != nil {
		return err
	} else if err := crypto.CheckPubKey(transfer.Receiver); err != nil {
		return err
	} else if transfer.Amount < 0 {
		return errors.New("negative transfer amount")
	} else {
		return nil
	}
}

func (transfer *Transfer) isSigned() bool {
	id := append(transfer.Sender, transfer.Receiver...)
	id = append(id, []byte(strconv.FormatInt(transfer.Amount, 10))...)
	id = append(id, []byte(strconv.FormatInt(transfer.Time, 10))...)
	return crypto.Verify(transfer.Sender, id, transfer.Signature)
}

// ------------------------------------------------------------------------------------------------------------------- //
// STAKE

type Stake struct {
	User      []byte
	Validator []byte
	Amount    int64
	Time      int64
	Signature []byte
}

func (stake *Stake) Hash() []byte {
	sum := append(stake.User, stake.Validator...)
	sum = append(sum, []byte(strconv.FormatInt(stake.Amount, 10))...)
	sum = append(sum, []byte(strconv.FormatInt(stake.Time, 10))...)
	sum = append(sum, stake.Signature...)
	hash := sha256.Sum256(sum)
	return hash[:]
}

func (stake *Stake) check() error {
	if err := crypto.CheckPubKey(stake.User); err != nil {
		return err
	} else if err := crypto.CheckEDPubKey(stake.Validator); err != nil {
		return err
	} else {
		return nil
	}
}

func (stake *Stake) isSigned() bool {
	id := append(stake.User, stake.Validator...)
	id = append(id, []byte(strconv.FormatInt(stake.Amount, 10))...)
	id = append(id, []byte(strconv.FormatInt(stake.Time, 10))...)
	if stake.Amount >= 0 {
		return crypto.Verify(stake.User, id, stake.Signature)
	} else {
		return crypto.VerifyED(stake.Validator, id, stake.Signature)
	}
}

// ------------------------------------------------------------------------------------------------------------------- //
// REWARD

type Reward struct {
	Info     *RewardInfo
	Confirms []*RewardConfirm
	State    RewardState
}

type RewardInfo struct {
	Requirer        []byte
	Validator       []byte
	Acceptor        []byte
	ValidatorAmount int64
	ProviderAmount  int64
	AcceptorAmount  int64
	MaxConfirms     int64
}

type RewardConfirm struct {
	Provider []byte
}

type RewardState int8

const RewardOpen RewardState = 0
const RewardClosed RewardState = 1

func (reward *Reward) Hash() []byte {
	sum := append(reward.Info.Requirer, reward.Info.Validator...)
	sum = append(sum, reward.Info.Acceptor...)
	sum = append(sum, []byte(strconv.FormatInt(reward.Info.ValidatorAmount, 10))...)
	sum = append(sum, []byte(strconv.FormatInt(reward.Info.ProviderAmount, 10))...)
	sum = append(sum, []byte(strconv.FormatInt(reward.Info.AcceptorAmount, 10))...)
	sum = append(sum, []byte(strconv.FormatInt(reward.Info.MaxConfirms, 10))...)
	for _, confirm := range reward.Confirms {
		sum = append(sum, confirm.Provider...)
	}
	hash := sha256.Sum256(sum)
	return hash[:]
}

func (reward *Reward) totalAmount() int64 {
	return (reward.Info.ValidatorAmount + reward.Info.ProviderAmount + reward.Info.AcceptorAmount) * reward.Info.MaxConfirms
}

func (reward *Reward) inRange() bool {
	return int64(len(reward.Confirms)) < reward.Info.MaxConfirms && reward.State == RewardOpen
}

func (reward *Reward) onCloseReturn() int64 {
	paid := (reward.Info.ValidatorAmount + reward.Info.ProviderAmount + reward.Info.AcceptorAmount) * reward.Info.MaxConfirms
	due := (reward.Info.ValidatorAmount + reward.Info.ProviderAmount + reward.Info.AcceptorAmount) * int64(len(reward.Confirms))
	return paid - due
}

// ------------------------------------------------------------------------------------------------------------------- //
// FEE

type Fee struct {
	User    []byte
	ValAddr []byte
	TxHash  []byte
}

func (fee *Fee) Hash() []byte {
	sum := append(fee.User, fee.ValAddr...)
	sum = append(sum, fee.TxHash...)
	hash := sha256.Sum256(sum)
	return hash[:]
}

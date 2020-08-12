package modules

import (
	"crypto/sha256"
	"dbc-node/crypto"
	"encoding/hex"
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

type Balance struct {
	Users      map[string]int64
	Validators map[string]int64
	Transfers  []*Transfer
	Stakes     []*Stake
	Rewards    []Reward
	Fees       []*Fee
}

func NewBalance(oldBalance *Balance) *Balance {
	balance := &Balance{
		Users:      make(map[string]int64),
		Validators: make(map[string]int64),
	}
	for user, value := range oldBalance.Users {
		balance.Users[user] = value
	}
	for validator, value := range oldBalance.Validators {
		balance.Validators[validator] = value
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
	for fee, value := range oldBalance.Fees {
		balance.Fees[fee] = value
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

func (balance *Balance) AddTransfer(transfer *Transfer) {
	id := append(transfer.Sender, transfer.Receiver...)
	id = append(id, []byte(strconv.FormatInt(transfer.Amount, 10))...)
	id = append(id, []byte(strconv.FormatInt(transfer.Time, 10))...)
	isSigned := crypto.Verify(transfer.Sender, id, transfer.Signature)
	sender := hex.EncodeToString(transfer.Sender)
	hasBalance := balance.Users[sender] >= transfer.Amount
	if isSigned && hasBalance {
		balance.Transfers = append(balance.Transfers, transfer)
		receiver := hex.EncodeToString(transfer.Receiver)
		balance.Users[sender] -= transfer.Amount
		balance.Users[receiver] += transfer.Amount
	}
}

func (balance *Balance) AddStake(stake *Stake) {
	id := append(stake.User, stake.Validator...)
	id = append(id, []byte(strconv.FormatInt(stake.Amount, 10))...)
	id = append(id, []byte(strconv.FormatInt(stake.Time, 10))...)
	isSigned := false
	hasBalance := false
	user := hex.EncodeToString(stake.User)
	validator := hex.EncodeToString(stake.Validator)
	if stake.Amount > 0 {
		isSigned = crypto.Verify(stake.User, id, stake.Signature)
		hasBalance = balance.Users[user] >= stake.Amount
	} else if stake.Amount < 0 {
		isSigned = crypto.VerifyED(stake.Validator, id, stake.Signature)
		hasBalance = balance.Validators[validator] >= -stake.Amount
	}
	if isSigned && hasBalance {
		balance.Stakes = append(balance.Stakes, stake)
		balance.Users[user] -= stake.Amount
		balance.Validators[validator] += stake.Amount
	}
}

func (balance *Balance) AddReward(reward Reward) (success bool, index int) {
	requirer := hex.EncodeToString(reward.Info.Requirer)
	totalAmount := (reward.Info.ValidatorAmount + reward.Info.ProviderAmount + reward.Info.AcceptorAmount) * reward.Info.MaxConfirms
	hasBalance := balance.Users[requirer] >= totalAmount
	if hasBalance {
		balance.Rewards = append(balance.Rewards, reward)
		balance.Users[requirer] -= totalAmount
		return true, len(balance.Rewards) - 1
	} else {
		return false, -1
	}
}

func (balance *Balance) ConfirmReward(confirm *RewardConfirm, index int) {
	reward := &balance.Rewards[index]
	if int64(len(reward.Confirms)) < reward.Info.MaxConfirms && reward.State == RewardOpen {
		balance.Rewards[index].Confirms = append(balance.Rewards[index].Confirms, confirm)
		reward := balance.Rewards[index]
		validator := hex.EncodeToString(reward.Info.Validator)
		provider := hex.EncodeToString(confirm.Provider)
		acceptor := hex.EncodeToString(reward.Info.Acceptor)
		balance.Users[validator] += reward.Info.ValidatorAmount
		balance.Users[provider] += reward.Info.ProviderAmount
		balance.Users[acceptor] += reward.Info.AcceptorAmount
	}
}

func (balance *Balance) CloseReward(index int) {
	reward := &balance.Rewards[index]
	if reward.State == RewardOpen {
		reward.State = RewardClosed
		paid := (reward.Info.ValidatorAmount + reward.Info.ProviderAmount + reward.Info.AcceptorAmount) * reward.Info.MaxConfirms
		due := (reward.Info.ValidatorAmount + reward.Info.ProviderAmount + reward.Info.AcceptorAmount) * int64(len(reward.Confirms))
		change := paid - due
		requirer := hex.EncodeToString(reward.Info.Requirer)
		balance.Users[requirer] += change
	}
}

func (balance *Balance) AddFee(fee *Fee) {
	user := hex.EncodeToString(fee.User)
	hasBalance := balance.Users[user] >= TxFee
	if hasBalance {
		balance.Fees = append(balance.Fees, fee)
		validator := hex.EncodeToString(fee.Validator)
		balance.Users[user] -= TxFee
		balance.Validators[validator] += TxFee
	}
}

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

func (promise *Reward) Hash() []byte {
	sum := append(promise.Info.Requirer, promise.Info.Validator...)
	sum = append(sum, promise.Info.Acceptor...)
	sum = append(sum, []byte(strconv.FormatInt(promise.Info.ValidatorAmount, 10))...)
	sum = append(sum, []byte(strconv.FormatInt(promise.Info.ProviderAmount, 10))...)
	sum = append(sum, []byte(strconv.FormatInt(promise.Info.AcceptorAmount, 10))...)
	sum = append(sum, []byte(strconv.FormatInt(promise.Info.MaxConfirms, 10))...)
	for _, confirm := range promise.Confirms {
		sum = append(sum, confirm.Provider...)
	}
	hash := sha256.Sum256(sum)
	return hash[:]
}

type Fee struct {
	User      []byte
	Validator []byte
	TxHash    []byte
}

func (fee *Fee) Hash() []byte {
	sum := append(fee.User, fee.Validator...)
	sum = append(sum, fee.TxHash...)
	hash := sha256.Sum256(sum)
	return hash[:]
}

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
	Users       map[string]int64
	Validators  map[string]int64
	Transfers   []*Transfer
	Stakes      []*Stake
	Rewards     map[[2]int]*Reward
	ConfRewards map[[2]int]bool
	Fees        []*Fee
}

func NewBalance(oldBalance *Balance) *Balance {
	balance := &Balance{
		Users:       make(map[string]int64),
		Validators:  make(map[string]int64),
		Rewards:     make(map[[2]int]*Reward),
		ConfRewards: make(map[[2]int]bool),
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
	for reward, value := range oldBalance.Rewards {
		balance.Rewards[reward] = value
	}
	for conf, value := range oldBalance.ConfRewards {
		balance.ConfRewards[conf] = value
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
	for _, confirmed := range balance.ConfRewards {
		if confirmed {
			sum = append(sum, '\xFF')
		} else {
			sum = append(sum, '\x00')
		}
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

func (balance *Balance) AddReward(reward *Reward, dataIndex, versionIndex int) bool {
	dtRequirer := hex.EncodeToString(reward.Requirer)
	totalAmount := reward.ValidatorAmount + reward.ProviderAmount + reward.AcceptorAmount
	hasBalance := balance.Users[dtRequirer] >= totalAmount
	if hasBalance {
		balance.Rewards[[2]int{dataIndex, versionIndex}] = reward
		dtValidator := hex.EncodeToString(reward.Validator)
		dtProvider := hex.EncodeToString(reward.Provider)
		dtAcceptor := hex.EncodeToString(reward.Acceptor)
		balance.Users[dtRequirer] -= totalAmount
		balance.Users[dtValidator] += reward.ValidatorAmount
		balance.Users[dtProvider] += reward.ProviderAmount
		balance.Users[dtAcceptor] += reward.AcceptorAmount
		return true
	} else {
		return false
	}
}

func (balance *Balance) ConfirmReward(dataIndex, versionIndex int) {
	balance.ConfRewards[[2]int{dataIndex, versionIndex}] = true
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
	Requirer        []byte
	Validator       []byte
	Provider        []byte
	Acceptor        []byte
	ValidatorAmount int64
	ProviderAmount  int64
	AcceptorAmount  int64
}

func (reward *Reward) Hash() []byte {
	sum := append(reward.Requirer, reward.Provider...)
	sum = append(sum, reward.Provider...)
	sum = append(sum, reward.Acceptor...)
	sum = append(sum, []byte(strconv.FormatInt(reward.ValidatorAmount, 10))...)
	sum = append(sum, []byte(strconv.FormatInt(reward.ProviderAmount, 10))...)
	sum = append(sum, []byte(strconv.FormatInt(reward.AcceptorAmount, 10))...)
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

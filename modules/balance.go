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

type Balance struct {
	Users       map[string]int64
	Validators  map[string]int64
	Transfers   []*Transfer
	Stakes      []*Stake
	Rewards     map[[2]int]*Reward
	ConfRewards map[[2]int]bool
	Fees        []*Fee
}

func NewBalance(balance *Balance) *Balance {
	return &Balance{
		Users:       balance.Users,
		Validators:  balance.Validators,
		Transfers:   balance.Transfers,
		Stakes:      balance.Stakes,
		Rewards:     balance.Rewards,
		ConfRewards: balance.ConfRewards,
		Fees:        balance.Fees,
	}
}

func (balance *Balance) hash() []byte {
	var sum []byte
	for _, transfer := range balance.Transfers {
		sum = append(sum, transfer.hash()...)
	}
	for _, stake := range balance.Stakes {
		sum = append(sum, stake.hash()...)
	}
	hash := sha256.Sum256(sum)
	return hash[:]
}

func (balance *Balance) AddTransfer(transfer *Transfer) {
	id := append(transfer.sender, transfer.receiver...)
	id = append(id, []byte(strconv.FormatInt(transfer.amount, 10))...)
	id = append(id, []byte(strconv.FormatInt(transfer.time, 10))...)
	isSigned := crypto.Verify(transfer.sender, id, transfer.signature)
	sender := hex.EncodeToString(transfer.sender)
	hasBalance := balance.Users[sender] >= transfer.amount
	if isSigned && hasBalance {
		balance.Transfers = append(balance.Transfers, transfer)
		receiver := hex.EncodeToString(transfer.receiver)
		balance.Users[sender] -= transfer.amount
		balance.Users[receiver] += transfer.amount
	}
}

func (balance *Balance) AddStake(stake *Stake) {
	id := append(stake.user, stake.validator...)
	id = append(id, []byte(strconv.FormatInt(stake.amount, 10))...)
	id = append(id, []byte(strconv.FormatInt(stake.time, 10))...)
	isSigned := false
	hasBalance := false
	user := hex.EncodeToString(stake.user)
	validator := hex.EncodeToString(stake.validator)
	if stake.amount > 0 {
		isSigned = crypto.Verify(stake.user, id, stake.signature)
		hasBalance = balance.Users[user] >= stake.amount
	} else if stake.amount < 0 {
		isSigned = crypto.VerifyED(stake.validator, id, stake.signature)
		hasBalance = balance.Validators[validator] >= -stake.amount
	}
	if isSigned && hasBalance {
		balance.Stakes = append(balance.Stakes, stake)
		balance.Users[user] -= stake.amount
		balance.Validators[validator] += stake.amount
	}
}

func (balance *Balance) AddReward(reward *Reward, dataIndex, versionIndex int) {
	dtRequirer := hex.EncodeToString(reward.dtRequirer)
	totalAmount := reward.dtValidatorAmount + reward.dtProviderAmount + reward.dtAcceptorAmount
	hasBalance := balance.Users[dtRequirer] >= totalAmount
	if hasBalance {
		balance.Rewards[[2]int{dataIndex, versionIndex}] = reward
		dtValidator := hex.EncodeToString(reward.dtValidator)
		dtProvider := hex.EncodeToString(reward.dtProvider)
		dtAcceptor := hex.EncodeToString(reward.dtAcceptor)
		balance.Users[dtRequirer] -= totalAmount
		balance.Users[dtValidator] += reward.dtValidatorAmount
		balance.Users[dtProvider] += reward.dtProviderAmount
		balance.Users[dtAcceptor] += reward.dtAcceptorAmount
	}
}

func (balance *Balance) ConfirmReward(dataIndex, versionIndex int) {
	balance.ConfRewards[[2]int{dataIndex, versionIndex}] = true
}

func (balance *Balance) AddFee(fee *Fee) {
	user := hex.EncodeToString(fee.user)
	hasBalance := balance.Users[user] >= TxFee
	if hasBalance {
		balance.Fees = append(balance.Fees, fee)
		validator := hex.EncodeToString(fee.validator)
		balance.Users[user] -= TxFee
		balance.Validators[validator] += TxFee
	}
}

type Transfer struct {
	sender    []byte
	receiver  []byte
	amount    int64
	time      int64
	signature []byte
}

func (transfer *Transfer) hash() []byte {
	sum := append(transfer.sender, transfer.receiver...)
	sum = append(sum, []byte(strconv.FormatInt(transfer.amount, 10))...)
	sum = append(sum, []byte(strconv.FormatInt(transfer.time, 10))...)
	sum = append(sum, transfer.signature...)
	hash := sha256.Sum256(sum)
	return hash[:]
}

type Stake struct {
	user      []byte
	validator []byte
	amount    int64
	time      int64
	signature []byte
}

func (stake *Stake) hash() []byte {
	sum := append(stake.user, stake.validator...)
	sum = append(sum, []byte(strconv.FormatInt(stake.amount, 10))...)
	sum = append(sum, []byte(strconv.FormatInt(stake.time, 10))...)
	sum = append(sum, stake.signature...)
	hash := sha256.Sum256(sum)
	return hash[:]
}

type Reward struct {
	dtRequirer        []byte
	dtValidator       []byte
	dtProvider        []byte
	dtAcceptor        []byte
	dtValidatorAmount int64
	dtProviderAmount  int64
	dtAcceptorAmount  int64
}

func (reward *Reward) hash() []byte {
	sum := append(reward.dtRequirer, reward.dtProvider...)
	sum = append(sum, reward.dtProvider...)
	sum = append(sum, reward.dtAcceptor...)
	sum = append(sum, []byte(strconv.FormatInt(reward.dtValidatorAmount, 10))...)
	sum = append(sum, []byte(strconv.FormatInt(reward.dtProviderAmount, 10))...)
	sum = append(sum, []byte(strconv.FormatInt(reward.dtAcceptorAmount, 10))...)
	hash := sha256.Sum256(sum)
	return hash[:]
}

type Fee struct {
	user      []byte
	validator []byte
	txHash    []byte
}

func (fee *Fee) hash() []byte {
	sum := append(fee.user, fee.validator...)
	sum = append(sum, fee.txHash...)
	hash := sha256.Sum256(sum)
	return hash[:]
}

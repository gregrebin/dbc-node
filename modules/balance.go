package modules

import (
	"crypto/sha256"
	"dbc-node/crypto"
	"encoding/hex"
	"strconv"
)

type Balance struct {
	Users      map[string]int64
	Validators map[string]int64
	GasPrice   map[string]int64
	Transfers  []*Transfer
	Stakes     []*Stake
}

func NewBalance(balance *Balance) *Balance {
	return &Balance{
		Users:      balance.Users,
		Validators: balance.Validators,
		Transfers:  balance.Transfers,
		Stakes:     balance.Stakes,
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

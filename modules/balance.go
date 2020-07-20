package modules

import (
	"dbc-node/crypto"
	"encoding/hex"
	"strconv"
)

type Balance struct {
	Users      map[string]int64
	Validators map[string]int64
	Transfers  []Transfer
	Stakes     []Stake
}

func NewBalance(balance *Balance) *Balance {
	return &Balance{
		Users:      balance.Users,
		Validators: balance.Validators,
	}
}

func (balance *Balance) hash() []byte {
	return nil
}

func (balance *Balance) AddTransfer(transfer Transfer) {
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

func (balance *Balance) AddStake(stake Stake) {
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

type Stake struct {
	user      []byte
	validator []byte
	amount    int64
	time      int64
	signature []byte
}

package tests

import (
	"bytes"
	"crypto/sha256"
	"dbc-node/crypto"
	"dbc-node/modules"
	"encoding/hex"
	"strconv"
	"testing"
	"time"
)

func initBalance() *modules.Balance {
	return modules.NewBalance(&modules.Balance{
		Users:      initialUsers,
		Validators: initialValidators,
	})
}

func TestBalance(t *testing.T) {
	balance := initBalance()
	if balance.Users[hex.EncodeToString(requirerPubKey)] != tokenDistribution["Requirer"] &&
		balance.Users[hex.EncodeToString(validatorPubKey)] != tokenDistribution["Validator"] &&
		balance.Users[hex.EncodeToString(providerPubKey)] != tokenDistribution["Provider"] &&
		balance.Users[hex.EncodeToString(acceptorPubKey)] != tokenDistribution["Acceptor"] {
		t.Errorf("Failed initializing balance users token distribution")
	}
	if balance.Transfers != nil && balance.Stakes != nil && balance.Rewards != nil && balance.Fees != nil {
		t.Errorf("Failed initializing balance transactions list")
	}
	validHash := sha256.Sum256(nil)
	hash := balance.Hash()
	if bytes.Compare(hash, validHash[:]) != 0 {
		t.Errorf("Failed initializing balance hash")
	}
}

func TestAddTransfer(t *testing.T) {
	balance := initBalance()
	sender := acceptorPubKey
	senderKey := acceptorPrivKey
	receiver := requirerPubKey
	amount := modules.ToSats(2)
	transfer := mockTransfer(sender, senderKey, receiver, amount)
	balance.AddTransfer(transfer)
	if len(balance.Transfers) != 1 {
		t.Errorf("Failed to register transfer")
	}
	if balance.Users[hex.EncodeToString(sender)] != (initialUsers[hex.EncodeToString(sender)] - amount) {
		t.Errorf("Failed to subtract transfer ammount")
	}
	if balance.Users[hex.EncodeToString(receiver)] != (initialUsers[hex.EncodeToString(receiver)] + amount) {
		t.Errorf("Failder to add transfer ammount")
	}
	validHash := sha256.Sum256(transfer.Hash())
	if bytes.Compare(balance.Hash(), validHash[:]) != 0 {
		t.Errorf("Incorrect hash after transfer")
	}
}

func mockTransfer(sender, senderKey, receiver []byte, amount int64) *modules.Transfer {
	time := time.Now().Unix()
	id := append(sender, receiver...)
	id = append(id, strconv.FormatInt(amount, 10)...)
	id = append(id, strconv.FormatInt(time, 10)...)
	return &modules.Transfer{
		Sender:    sender,
		Receiver:  receiver,
		Amount:    amount,
		Time:      time,
		Signature: crypto.Sign(senderKey, id),
	}
}

func TestAddStake(t *testing.T) {
	balance := initBalance()
	user := providerPubKey
	userKey := providerPrivKey
	validator := stakePubKey
	validatorKey := stakePrivKey
	stakeAmount := modules.ToSats(3)
	stake := mockStake(user, userKey, validator, validatorKey, stakeAmount)
	balance.AddStake(stake)
	if len(balance.Stakes) != 1 {
		t.Errorf("Failed to register stake")
	}
	if balance.Users[hex.EncodeToString(providerPubKey)] != (initialUsers[hex.EncodeToString(providerPubKey)] - stakeAmount) {
		t.Errorf("Failed to substract stake amount")
	}
	if balance.Validators[hex.EncodeToString(stakePubKey)] != (initialValidators[hex.EncodeToString(stakePubKey)] + stakeAmount) {
		t.Errorf("Failed to add stake amount")
	}
	validHash := sha256.Sum256(stake.Hash())
	if bytes.Compare(balance.Hash(), validHash[:]) != 0 {
		t.Errorf("Incorrect hash after stake")
	}
	unstakeAmount := modules.ToSats(-5)
	unstake := mockStake(user, userKey, validator, validatorKey, unstakeAmount)
	balance.AddStake(unstake)
	if len(balance.Stakes) != 2 {
		t.Errorf("Failed to register unstake")
	}
	if balance.Users[hex.EncodeToString(providerPubKey)] != (initialUsers[hex.EncodeToString(providerPubKey)] - stakeAmount - unstakeAmount) {
		t.Errorf("Failed to substract unstake amount")
	}
	if balance.Validators[hex.EncodeToString(stakePubKey)] != (initialValidators[hex.EncodeToString(stakePubKey)] + stakeAmount + unstakeAmount) {
		t.Errorf("Failed to add unstake amount")
	}
	validHash = sha256.Sum256(append(stake.Hash(), unstake.Hash()...))
	if bytes.Compare(balance.Hash(), validHash[:]) != 0 {
		t.Errorf("Incorrect hash after stake")
	}
}

func mockStake(user, userKey, validator, validatorKey []byte, amount int64) *modules.Stake {
	time := time.Now().Unix()
	id := append(user, validator...)
	id = append(id, strconv.FormatInt(amount, 10)...)
	id = append(id, strconv.FormatInt(time, 10)...)
	var signature []byte
	if amount >= 0 {
		signature = crypto.Sign(userKey, id)
	} else {
		signature = crypto.SignED(validatorKey, id)
	}
	return &modules.Stake{
		User:      user,
		Validator: validator,
		Amount:    amount,
		Time:      time,
		Signature: signature,
	}
}

//func TestAddReward(t *testing.T) {
//	balance := initBalance()
//	reward := &modules.Reward{
//		Requirer:        requirerPubKey,
//		Validator:       validatorPubKey,
//		Provider:        providerPubKey,
//		Acceptor:        acceptorPubKey,
//		ValidatorAmount: modules.ToSats(1),
//		ProviderAmount:  modules.ToSats(5),
//		AcceptorAmount:  modules.ToSats(2),
//	}
//	balance.AddReward(reward, 1, 3)
//	if len(balance.Rewards) != 1 {
//		t.Errorf("Failed to register reward")
//	}
//	if balance.Rewards[[2]int{1, 3}] == nil {
//		t.Errorf("Failed to find the reward")
//	}
//	totalAmount := reward.ValidatorAmount + reward.ProviderAmount + reward.AcceptorAmount
//	if balance.Users[hex.EncodeToString(requirerPubKey)] != (initialUsers[hex.EncodeToString(requirerPubKey)] - totalAmount) {
//		t.Errorf("Failed to substract reward amount from requirer")
//	}
//	if balance.Users[hex.EncodeToString(validatorPubKey)] != (initialUsers[hex.EncodeToString(validatorPubKey)] + reward.ValidatorAmount) {
//		t.Errorf("Failed to add reward amount to validator")
//	}
//	if balance.Users[hex.EncodeToString(providerPubKey)] != (initialUsers[hex.EncodeToString(providerPubKey)] + reward.ProviderAmount) {
//		t.Errorf("Failed to add reward amount to provider")
//	}
//	if balance.Users[hex.EncodeToString(acceptorPubKey)] != (initialUsers[hex.EncodeToString(acceptorPubKey)] + reward.AcceptorAmount) {
//		t.Errorf("Failed to add reward amount to acceptor")
//	}
//	validHash := sha256.Sum256(reward.Hash())
//	if bytes.Compare(balance.Hash(), validHash[:]) != 0 {
//		t.Errorf("Incorrect hash after reward")
//	}
//}

func TestAddFee(t *testing.T) {
	balance := initBalance()
	hash := sha256.Sum256([]byte("Some transaction bytes"))
	fee := &modules.Fee{
		User:      requirerPubKey,
		Validator: stakePubKey,
		TxHash:    hash[:],
	}
	balance.AddFee(fee)
	if len(balance.Fees) != 1 {
		t.Errorf("Failed to register fee")
	}
	if balance.Users[hex.EncodeToString(requirerPubKey)] != (initialUsers[hex.EncodeToString(requirerPubKey)] - modules.TxFee) {
		t.Errorf("Failed to substract fee amount")
	}
	if balance.Validators[hex.EncodeToString(stakePubKey)] != (initialValidators[hex.EncodeToString(stakePubKey)] + modules.TxFee) {
		t.Errorf("Failed to add fee amount")
	}
	validHash := sha256.Sum256(fee.Hash())
	if bytes.Compare(balance.Hash(), validHash[:]) != 0 {
		t.Errorf("Incorrect hash after stake")
	}
}

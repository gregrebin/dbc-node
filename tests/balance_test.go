package tests

import (
	"bytes"
	"crypto/sha256"
	"dbc-node/crypto"
	"dbc-node/modules"
	"encoding/hex"
	"reflect"
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

func TestAddReward(t *testing.T) {
	balance := initBalance()

	reward := mockReward()
	_, rewardIndex := balance.AddReward(reward)
	if !rewardAdded(balance, reward, rewardIndex) {
		t.Errorf("Failed to add reward")
	}

	for i := int64(1); i < reward.Info.MaxConfirms; i++ {
		confirm := mockConfirm()
		balance.ConfirmReward(confirm, rewardIndex)
		if !rewardConfirmed(balance, reward, rewardIndex, i) {
			t.Errorf("Failed to confirm reward at iteration " + strconv.FormatInt(i, 10))
		}
	}

	confirm := mockConfirm()
	balance.ConfirmReward(confirm, rewardIndex)
	if rewardConfirmed(balance, reward, rewardIndex, reward.Info.MaxConfirms+1) {
		t.Errorf("Failed to stop confirming rewards after surpassing max number")
	}

	reward = mockReward()
	_, rewardIndex = balance.AddReward(reward)
	confirm = mockConfirm()
	balance.ConfirmReward(confirm, rewardIndex)
	balance.CloseReward(rewardIndex)
	if rewardConfirmed(balance, reward, rewardIndex, 2) {
		t.Errorf("Reward confirmed after closing")
	}

	//validHash := sha256.Sum256(sum)
	//if bytes.Compare(balance.Hash(), validHash[:]) != 0 {
	//	t.Errorf("Incorrect hash after reward")
	//}
}

func mockReward() modules.Reward {
	return modules.Reward{
		Info: &modules.RewardInfo{
			Requirer:        requirerPubKey,
			Validator:       validatorPubKey,
			Acceptor:        acceptorPubKey,
			ValidatorAmount: modules.ToSats(2),
			ProviderAmount:  modules.ToSats(5),
			AcceptorAmount:  modules.ToSats(3),
			MaxConfirms:     3,
		},
	}
}

func mockConfirm() *modules.RewardConfirm {
	return &modules.RewardConfirm{
		Provider: providerPubKey,
	}
}

func rewardAdded(balance *modules.Balance, reward modules.Reward, rewardIndex int) bool {
	cost := (reward.Info.ValidatorAmount + reward.Info.ProviderAmount + reward.Info.AcceptorAmount) * reward.Info.MaxConfirms
	return reflect.DeepEqual(balance.Rewards[rewardIndex], reward) &&
		balance.Users[hex.EncodeToString(requirerPubKey)] == (initialUsers[hex.EncodeToString(requirerPubKey)]-cost) &&
		balance.Users[hex.EncodeToString(validatorPubKey)] == initialUsers[hex.EncodeToString(validatorPubKey)] &&
		balance.Users[hex.EncodeToString(providerPubKey)] == initialUsers[hex.EncodeToString(providerPubKey)] &&
		balance.Users[hex.EncodeToString(acceptorPubKey)] == initialUsers[hex.EncodeToString(acceptorPubKey)]
}

func rewardConfirmed(balance *modules.Balance, reward modules.Reward, rewardIndex int, count int64) bool {
	return balance.Users[hex.EncodeToString(validatorPubKey)] == (initialUsers[hex.EncodeToString(validatorPubKey)]+(reward.Info.ValidatorAmount*count)) &&
		balance.Users[hex.EncodeToString(providerPubKey)] == (initialUsers[hex.EncodeToString(providerPubKey)]+(reward.Info.ProviderAmount*count)) &&
		balance.Users[hex.EncodeToString(acceptorPubKey)] == (initialUsers[hex.EncodeToString(acceptorPubKey)]+(reward.Info.AcceptorAmount*count)) &&
		len(balance.Rewards[rewardIndex].Confirms) == int(count)
}

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

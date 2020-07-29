package tests

import (
	"dbc-node/app"
	"dbc-node/messages"
	"dbc-node/modules"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"github.com/tendermint/tendermint/abci/types"
	"testing"
)

func TestApp(t *testing.T) {
	dbc := app.NewDataBlockChain(genUsers, genValidators)
	_ = dbc.Info(mockRequestInfo())

	checkTx(t, dbc, messages.TxAddData, 1)
	checkTx(t, dbc, messages.TxAddData, 2)
	checkTx(t, dbc, messages.TxAddData, 3)
	checkTx(t, dbc, messages.TxTransfer, 1)
	checkTx(t, dbc, messages.TxTransfer, 2)
	checkTx(t, dbc, messages.TxStake, 1)
}

func checkTx(t *testing.T, dbc *app.DataBlockChain, txType messages.TransactionType, txCount int) {
	switch txType {

	case messages.TxAddData:
		_ = dbc.DeliverTx(mockRequestDeliverTx(messages.TxAddData))
		if len(dbc.New.Dataset.DataList) != txCount {
			t.Errorf("Transaction not added")
		}
		_ = dbc.Commit()
		if len(dbc.New.Dataset.DataList) != txCount {
			t.Errorf("Transaction not retained")
		}
		_ = dbc.Query(mockRequestQuery())

	case messages.TxTransfer:
		_ = dbc.DeliverTx(mockRequestDeliverTx(messages.TxTransfer))
		if len(dbc.New.Balance.Transfers) != txCount {
			t.Errorf("Transaction not added")
		}
		if dbc.New.Balance.Users[hex.EncodeToString(requirerPubKey)] != (genUsers[hex.EncodeToString(requirerPubKey)] - modules.ToSats(2*int64(txCount))) {
			t.Errorf("Transfer amount not substracted")
		}
		if dbc.New.Balance.Users[hex.EncodeToString(acceptorPubKey)] != (genUsers[hex.EncodeToString(acceptorPubKey)] + modules.ToSats(2*int64(txCount))) {
			t.Errorf("Transfer amount not substracted")
		}
		_ = dbc.Commit()
		if len(dbc.New.Balance.Transfers) != txCount {
			t.Errorf("Transaction not retained")
		}
		if dbc.New.Balance.Users[hex.EncodeToString(requirerPubKey)] != (genUsers[hex.EncodeToString(requirerPubKey)] - modules.ToSats(2*int64(txCount))) {
			t.Errorf("Transfer amount not substracted")
		}
		if dbc.New.Balance.Users[hex.EncodeToString(acceptorPubKey)] != (genUsers[hex.EncodeToString(acceptorPubKey)] + modules.ToSats(2*int64(txCount))) {
			t.Errorf("Transfer amount not substracted")
		}
		_ = dbc.Query(mockRequestQuery())

	case messages.TxStake:
		_ = dbc.DeliverTx(mockRequestDeliverTx(messages.TxStake))
		if len(dbc.New.Balance.Stakes) != txCount {
			t.Errorf("Transaction not added")
		}
		if dbc.New.Balance.Users[hex.EncodeToString(providerPubKey)] != (genUsers[hex.EncodeToString(providerPubKey)] - modules.ToSats(1*int64(txCount))) {
			t.Errorf("Stake amount not substracted")
		}
		if dbc.New.Balance.Validators[hex.EncodeToString(stakePubKey)] != (genValidators[hex.EncodeToString(stakePubKey)] + modules.ToSats(1*int64(txCount))) {
			t.Errorf("Stake amount not added")
		}
		_ = dbc.Commit()
		if len(dbc.New.Balance.Stakes) != txCount {
			t.Errorf("Transaction not retained")
		}
		if dbc.New.Balance.Users[hex.EncodeToString(providerPubKey)] != (genUsers[hex.EncodeToString(providerPubKey)] - modules.ToSats(1*int64(txCount))) {
			t.Errorf("Stake amount not substracted")
		}
		if dbc.New.Balance.Validators[hex.EncodeToString(stakePubKey)] != (genValidators[hex.EncodeToString(stakePubKey)] + modules.ToSats(1*int64(txCount))) {
			t.Errorf("Stake amount not added")
		}
	}
}

func mockRequestInfo() types.RequestInfo {
	return types.RequestInfo{
		Version:      "",
		BlockVersion: 0,
		P2PVersion:   0,
	}
}

func mockRequestDeliverTx(txType messages.TransactionType) types.RequestDeliverTx {
	transaction := messages.Transaction{
		TxType:       txType,
		DataIndex:    0,
		VersionIndex: 0,
	}
	switch txType {
	case messages.TxAddData:
		description := mockDescription()
		transaction.Description = description
	case messages.TxAddValidation:
		validation := mockValidation(zpks[0])
		transaction.Validation = validation
	case messages.TxAddPayload:
		payload := mockPayload(zpks[0])
		transaction.Payload = payload
	case messages.TxAcceptPayload:
		acceptedPayload := mockAcceptedPayload()
		transaction.AcceptedPayload = acceptedPayload
	case messages.TxTransfer:
		transfer := mockTransfer(requirerPubKey, requirerPrivKey, acceptorPubKey, modules.ToSats(2))
		transaction.Transfer = transfer
	case messages.TxStake:
		stake := mockStake(providerPubKey, providerPrivKey, stakePubKey, stakePrivKey, modules.ToSats(1))
		transaction.Stake = stake
	}
	tx, _ := json.Marshal(transaction)
	encodedTx := make([]byte, base64.StdEncoding.EncodedLen(len(tx)))
	base64.StdEncoding.Encode(encodedTx, tx)
	return types.RequestDeliverTx{
		Tx: encodedTx,
	}
}

func mockRequestQuery() types.RequestQuery {
	query := messages.Query{
		QrType:       messages.QueryDataset,
		DataIndex:    0,
		VersionIndex: 0,
	}
	data, _ := json.Marshal(query)
	encodedData := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(encodedData, data)
	return types.RequestQuery{
		Data:   encodedData,
		Path:   "",
		Height: 0,
		Prove:  false,
	}
}

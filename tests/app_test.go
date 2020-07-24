package tests

import (
	"dbc-node/app"
	"dbc-node/messages"
	"encoding/base64"
	"encoding/json"
	"github.com/tendermint/tendermint/abci/types"
	"testing"
)

func TestApp(t *testing.T) {
	dbc := app.NewDataBlockChain()
	_ = dbc.Info(mockRequestInfo())

	_ = dbc.DeliverTx(mockRequestDeliverTx())
	if len(dbc.New.Dataset.DataList) != 1 {
		t.Errorf("Transaction not added")
	}
	_ = dbc.Commit()
	if len(dbc.New.Dataset.DataList) != 1 {
		t.Errorf("Transaction not retained")
	}
	_ = dbc.Query(mockRequestQuery())

	_ = dbc.DeliverTx(mockRequestDeliverTx())
	if len(dbc.New.Dataset.DataList) != 2 {
		t.Errorf("Transaction not added")
	}
	_ = dbc.Commit()
	if len(dbc.New.Dataset.DataList) != 2 {
		t.Errorf("Transaction not retained")
	}
	_ = dbc.Query(mockRequestQuery())

	_ = dbc.DeliverTx(mockRequestDeliverTx())
	if len(dbc.New.Dataset.DataList) != 3 {
		t.Errorf("Transaction not added")
	}
	_ = dbc.Commit()
	if len(dbc.New.Dataset.DataList) != 3 {
		t.Errorf("Transaction not retained")
	}
	_ = dbc.Query(mockRequestQuery())
}

func mockRequestInfo() types.RequestInfo {
	return types.RequestInfo{
		Version:      "",
		BlockVersion: 0,
		P2PVersion:   0,
	}
}

func mockRequestDeliverTx() types.RequestDeliverTx {
	description := mockDescription()
	transaction := messages.Transaction{
		TxType:       messages.TxAddData,
		Description:  description,
		DataIndex:    0,
		VersionIndex: 0,
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

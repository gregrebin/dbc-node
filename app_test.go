package main

import (
	"encoding/base64"
	"encoding/json"
	"github.com/tendermint/tendermint/abci/types"
	"testing"
)

func TestApp(t *testing.T) {
	dbc := NewDataBlockChain()
	_ = dbc.Info(mockRequestInfo())

	_ = dbc.DeliverTx(mockRequestDeliverTx())
	if len(dbc.new.DataList) != 1 {
		t.Errorf("Transaction not added")
	}
	_ = dbc.Commit()
	if len(dbc.new.DataList) != 1 {
		t.Errorf("Transaction not retained")
	}
	_ = dbc.Query(mockRequestQuery())

	_ = dbc.DeliverTx(mockRequestDeliverTx())
	if len(dbc.new.DataList) != 2 {
		t.Errorf("Transaction not added")
	}
	_ = dbc.Commit()
	if len(dbc.new.DataList) != 2 {
		t.Errorf("Transaction not retained")
	}
	_ = dbc.Query(mockRequestQuery())

	_ = dbc.DeliverTx(mockRequestDeliverTx())
	if len(dbc.new.DataList) != 3 {
		t.Errorf("Transaction not added")
	}
	_ = dbc.Commit()
	if len(dbc.new.DataList) != 3 {
		t.Errorf("Transaction not retained")
	}
	_ = dbc.Query(mockRequestQuery())
	// TODO: needs better tests
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
	transaction := Transaction{
		TxType:       TxAddData,
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
	query := Query{
		QrType:       QueryState,
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

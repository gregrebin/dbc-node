package main

import (
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
	// TODO: need more test!!!
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
		Description:  &description,
		DataIndex:    0,
		VersionIndex: 0,
	}
	tx, _ := json.Marshal(transaction)
	return types.RequestDeliverTx{
		Tx: tx,
	}
}

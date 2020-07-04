package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	tendermint "github.com/tendermint/tendermint/abci/types"
)

type dataBlockChain struct {
	height    int64
	confirmed []*State // written at 2nd commit
	committed *State   // written at 1st commit
	new       *State   // written at deliverTx
}

var _ tendermint.Application = (*dataBlockChain)(nil)

func NewDataBlockChain() *dataBlockChain {
	state := NewState(&State{})
	return &dataBlockChain{
		height: 0,
		new:    state,
	}
}

func (dbc *dataBlockChain) stateAtHeight(height int) *State {
	if dbc.confirmed == nil {
		return nil
	}
	switch height {
	case 0: // return state at current height: last confirmed state
		return dbc.confirmed[len(dbc.confirmed)-1]
	case 1: // at height 1 there's no confirmed state
		return nil
	default: // confirmed states start at height 2
		return dbc.confirmed[height-2]
	}
}

func (dbc *dataBlockChain) Info(requestInfo tendermint.RequestInfo) tendermint.ResponseInfo {
	responseInfo := tendermint.ResponseInfo{
		Data:             "Some arbitrary information about dbc-node app",
		Version:          "V1",
		AppVersion:       1,
		LastBlockHeight:  dbc.height,
		LastBlockAppHash: dbc.committed.Hash(),
	}
	return responseInfo
}

func (dbc *dataBlockChain) SetOption(requestSetOption tendermint.RequestSetOption) tendermint.ResponseSetOption {
	responseSetOption := tendermint.ResponseSetOption{
		Code: 0,
		Log:  "",
		Info: "",
	}
	return responseSetOption
}

func (dbc *dataBlockChain) Query(requestQuery tendermint.RequestQuery) tendermint.ResponseQuery {
	data := make([]byte, base64.StdEncoding.DecodedLen(len(requestQuery.Data)))
	_, _ = base64.StdEncoding.Decode(data, requestQuery.Data)
	data = bytes.Trim(data, "\x00")
	var query Query
	_ = json.Unmarshal(data, &query)
	var value []byte
	state := dbc.stateAtHeight(int(requestQuery.Height))
	switch query.QrType {
	case QueryState:
		value, _ = json.Marshal(state)
	case QueryData:
		value, _ = json.Marshal(state.DataList[query.DataIndex])
	case QueryVersion:
		value, _ = json.Marshal(state.DataList[query.DataIndex].VersionList[query.VersionIndex])
	case QueryDescription:
		value, _ = json.Marshal(state.DataList[query.DataIndex].Description)
	case QueryValidation:
		value, _ = json.Marshal(state.DataList[query.DataIndex].VersionList[query.VersionIndex].Validation)
	case QueryPayload:
		value, _ = json.Marshal(state.DataList[query.DataIndex].VersionList[query.VersionIndex].Payload)
	case QueryAcceptedPayload:
		value, _ = json.Marshal(state.DataList[query.DataIndex].VersionList[query.VersionIndex].AcceptedPayload)
	}
	responseQuery := tendermint.ResponseQuery{
		Code:      uint32(0),
		Log:       "",
		Info:      "",
		Index:     -1,
		Key:       requestQuery.Data,
		Value:     value,
		Proof:     nil,
		Height:    0,
		Codespace: "",
	}
	return responseQuery
}

func (dbc *dataBlockChain) CheckTx(requestCheckTx tendermint.RequestCheckTx) tendermint.ResponseCheckTx {
	responseCheckTx := tendermint.ResponseCheckTx{
		Code:      uint32(0),
		Data:      nil,
		Log:       "",
		Info:      "",
		GasWanted: 0,
		GasUsed:   0,
		Events:    nil,
		Codespace: "",
	}
	return responseCheckTx
}

func (dbc *dataBlockChain) InitChain(requestInitChain tendermint.RequestInitChain) tendermint.ResponseInitChain {
	responseInitChain := tendermint.ResponseInitChain{
		ConsensusParams: nil,
		Validators:      nil,
	}
	return responseInitChain
}

func (dbc *dataBlockChain) BeginBlock(requestBeginBlock tendermint.RequestBeginBlock) tendermint.ResponseBeginBlock {
	responseBeginBlock := tendermint.ResponseBeginBlock{
		Events: nil,
	}
	return responseBeginBlock
}

func (dbc *dataBlockChain) DeliverTx(requestDeliverTx tendermint.RequestDeliverTx) tendermint.ResponseDeliverTx {
	tx := make([]byte, base64.StdEncoding.DecodedLen(len(requestDeliverTx.Tx)))
	_, _ = base64.StdEncoding.Decode(tx, requestDeliverTx.Tx)
	tx = bytes.Trim(tx, "\x00")
	var transaction Transaction
	_ = json.Unmarshal(tx, &transaction)
	switch transaction.TxType {
	case TxAddData:
		description := transaction.Description
		dbc.new.AddData(description)
	case TxAddValidation:
		validation := transaction.Validation
		dbc.new.AddValidation(validation, transaction.DataIndex)
	case TxAddPayload:
		payload := transaction.Payload
		dbc.new.AddPayload(payload, transaction.DataIndex, transaction.VersionIndex)
	case TxAcceptPayload:
		acceptedPayload := transaction.AcceptedPayload
		dbc.new.AcceptPayload(acceptedPayload, transaction.DataIndex, transaction.VersionIndex)
	}
	responseDeliverTx := tendermint.ResponseDeliverTx{
		Code:      uint32(0),
		Data:      nil,
		Log:       "",
		Info:      "",
		GasWanted: 0,
		GasUsed:   0,
		Events:    nil,
		Codespace: "",
	}
	return responseDeliverTx
}

func (dbc *dataBlockChain) EndBlock(requestEndBlock tendermint.RequestEndBlock) tendermint.ResponseEndBlock {
	responseEndBlock := tendermint.ResponseEndBlock{
		ValidatorUpdates:      nil,
		ConsensusParamUpdates: nil,
		Events:                nil,
	}
	return responseEndBlock
}

func (dbc *dataBlockChain) Commit() tendermint.ResponseCommit {
	if dbc.height > 0 { // we don't append to confirmed in the first commit, since there's no committed state yet
		dbc.confirmed = append(dbc.confirmed, dbc.committed)
	}
	dbc.committed = dbc.new
	dbc.new = NewState(dbc.committed)
	dbc.height++
	responseCommit := tendermint.ResponseCommit{
		Data:         dbc.committed.Hash(),
		RetainHeight: 0,
	}
	return responseCommit
}

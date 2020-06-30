package main

import (
	"bytes"
	"encoding/json"
	tendermint "github.com/tendermint/tendermint/abci/types"
)

type dataBlockChain struct {
	height      int64
	hash        []byte
	data        map[string]string
	uncommitted map[string]string
}

var _ tendermint.Application = (*dataBlockChain)(nil)

func NewDataBlockChain() *dataBlockChain {
	return &dataBlockChain{
		height:      0,
		hash:        []byte{},
		data:        map[string]string{},
		uncommitted: map[string]string{},
	}
}

func (dbc *dataBlockChain) Info(info tendermint.RequestInfo) tendermint.ResponseInfo {
	responseInfo := tendermint.ResponseInfo{
		Data:             "Some arbitrary information about dbc-node app",
		Version:          "V1",
		AppVersion:       1,
		LastBlockHeight:  dbc.height,
		LastBlockAppHash: dbc.hash,
	}
	return responseInfo
}

func (dbc *dataBlockChain) SetOption(option tendermint.RequestSetOption) tendermint.ResponseSetOption {
	responseSetOption := tendermint.ResponseSetOption{
		Code:                 0,
		Log:                  "",
		Info:                 "",
	}
	return responseSetOption
}

func (dbc *dataBlockChain) Query(query tendermint.RequestQuery) tendermint.ResponseQuery {
	question := string(query.Data)
	answer := dbc.data[question]
	code := uint32(0)
	if answer == "" {
		code = 1
	}
	// TODO: form a correct response, for now it's just the code, question an answer
	responseQuery := tendermint.ResponseQuery{
		Code:                 code,
		Log:                  "",
		Info:                 "",
		Index:                0,
		Key:                  []byte(question),
		Value:                []byte(answer),
		Proof:                nil,
		Height:               0,
		Codespace:            "",
	}
	return responseQuery
}

func (dbc *dataBlockChain) CheckTx(tx tendermint.RequestCheckTx) tendermint.ResponseCheckTx {
	transaction := tx.Tx
	containsEqualSign := bytes.ContainsAny(transaction, "=")
	code := uint32(0)
	if !containsEqualSign {
		code = 1
	}
	responseCheckTx := tendermint.ResponseCheckTx{
		Code:                 code,
		Data:                 nil,
		Log:                  "",
		Info:                 "",
		GasWanted:            0,
		GasUsed:              0,
		Events:               nil,
		Codespace:            "",
	}
	return responseCheckTx
}

func (dbc *dataBlockChain) InitChain(chain tendermint.RequestInitChain) tendermint.ResponseInitChain {
	responseInitChain := tendermint.ResponseInitChain{
		ConsensusParams:      nil,
		Validators:           nil,
	}
	return responseInitChain
}

func (dbc *dataBlockChain) BeginBlock(block tendermint.RequestBeginBlock) tendermint.ResponseBeginBlock {
	responseBeginBlock := tendermint.ResponseBeginBlock{
		Events:               nil,
	}
	return responseBeginBlock
}

func (dbc *dataBlockChain) DeliverTx(tx tendermint.RequestDeliverTx) tendermint.ResponseDeliverTx {
	transaction := tx.Tx
	containsEqualSign := bytes.ContainsAny(transaction, "=")
	code := uint32(0)
	if containsEqualSign {
		parts := bytes.SplitN(transaction, []byte("="), 2)
		question := string(parts[0])
		answer := string(parts[1])
		dbc.uncommitted[question] = answer
	} else {
		code = 1
	}
	responseDeliverTx := tendermint.ResponseDeliverTx{
		Code:                 code,
		Data:                 nil,
		Log:                  "",
		Info:                 "",
		GasWanted:            0,
		GasUsed:              0,
		Events:               nil,
		Codespace:            "",
	}
	return responseDeliverTx
}

func (dbc *dataBlockChain) EndBlock(block tendermint.RequestEndBlock) tendermint.ResponseEndBlock {
	responseEndBlock := tendermint.ResponseEndBlock{
		ValidatorUpdates:      nil,
		ConsensusParamUpdates: nil,
		Events:                nil,
	}
	return responseEndBlock
}

func (dbc *dataBlockChain) Commit() tendermint.ResponseCommit {
	for question, answer := range dbc.uncommitted {
		dbc.data[question] += answer + ";"
	}
	dbc.uncommitted = map[string]string{}
	hash, _ := json.Marshal(dbc.data)
	dbc.hash = hash
	dbc.height++
	responseCommit := tendermint.ResponseCommit{
		Data:         dbc.hash,
		RetainHeight: 0,
	}
	return responseCommit
}
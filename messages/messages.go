package messages

import "dbc-node/statedt"

type TransactionType string

const (
	TxAddData       TransactionType = "TxAddData"
	TxAddValidation TransactionType = "TxAddValidation"
	TxAddPayload    TransactionType = "TxAddPayload"
	TxAcceptPayload TransactionType = "TxAcceptPayload"
)

type Transaction struct {
	TxType TransactionType

	Description     *statedt.Description
	Validation      *statedt.Validation
	Payload         *statedt.Payload
	AcceptedPayload *statedt.AcceptedPayload

	DataIndex    int
	VersionIndex int
}

type QueryType string

const (
	QueryState           QueryType = "QueryState"
	QueryData            QueryType = "QueryData"
	QueryVersion         QueryType = "QueryVersion"
	QueryDescription     QueryType = "QueryDescription"
	QueryValidation      QueryType = "QueryValidation"
	QueryPayload         QueryType = "QueryPayload"
	QueryAcceptedPayload QueryType = "QueryAcceptedPayload"
)

type Query struct {
	QrType       QueryType
	DataIndex    int
	VersionIndex int
}
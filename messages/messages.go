package messages

import (
	"dbc-node/modules"
)

type TransactionType string

const (
	TxAddData       TransactionType = "TxAddData"
	TxAddValidation TransactionType = "TxAddValidation"
	TxAddPayload    TransactionType = "TxAddPayload"
	TxAcceptPayload TransactionType = "TxAcceptPayload"
	TxTransferDBCC  TransactionType = "TxTransferDBCC"
	TxStake         TransactionType = "TxStake"
	TxUnstake       TransactionType = "TxUnstake"
)

type Transaction struct {
	TxType TransactionType

	Description     *modules.Description
	Validation      *modules.Validation
	Payload         *modules.Payload
	AcceptedPayload *modules.AcceptedPayload

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
	QueryDBCCBalance     QueryType = "QueryDBCCBalance"
	QueryStake           QueryType = "QueryStake"
)

type Query struct {
	QrType       QueryType
	DataIndex    int
	VersionIndex int
}

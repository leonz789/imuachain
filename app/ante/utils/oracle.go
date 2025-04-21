package utils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	oracletypes "github.com/imua-xyz/imuachain/x/oracle/types"
)

const (
	// TxSizeLimitOraclePrice is the max size of oracle price tx without raw data piece
	TxSizeLimitOraclePrice = 400
	// TxSizeLimitOracleRawData is the max size of oracle price tx with raw data piece
	TxSizeLimitOracleRawData = 49600
)

// OracleCreatePriceTx returns the oracle create price txs in the given tx, and whether the tx is valid oracle tx and valid raw data tx
func OracleCreatePriceTx(tx sdk.Tx) (msgsOracle []*oracletypes.MsgCreatePrice, validOracle, validRawData, mixed bool) {
	msgs := tx.GetMsgs()
	if len(msgs) == 0 {
		return nil, false, false, false
	}
	msgsOracle = make([]*oracletypes.MsgCreatePrice, 0, len(msgs))

	l := len(msgs)
	for i := 0; i < l; i++ {
		msg := msgs[i]
		msgOracle, ok := msg.(*oracletypes.MsgCreatePrice)

		if !ok {
			return nil, false, false, (i > 0 || containsOracleMsgBeyond(msgs, i))
		}

		if msgOracle.IsPhaseTwo() {
			if len(msgs) > 1 {
				return nil, false, false, true
			}
			msgsOracle = append(msgsOracle, msgOracle)
			return msgsOracle, true, true, false
		}
		msgsOracle = append(msgsOracle, msgOracle)
	}
	return msgsOracle, true, false, false
}

func containsOracleMsgBeyond(msgs []sdk.Msg, i int) bool {
	for ; i < len(msgs); i++ {
		if _, ok := (msgs[i]).(*oracletypes.MsgCreatePrice); ok {
			return true
		}
	}
	return false
}

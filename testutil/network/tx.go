package network

import (
	// "github.com/cosmos/cosmos-sdk/client"
	"context"

	sdkmath "cosmossdk.io/math"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
)

// GenSignedTx construct a tx with input messages and sign it privatekey stored in keyring
func (n *Network) GenSignedTx(msgs []sdk.Msg, keyName string) ([]byte, error) {
	val := n.Validators[0]
	ctx := val.ClientCtx
	record, err := ctx.Keyring.Key(keyName)
	if err != nil {
		return nil, err
	}
	acc, err := record.GetAddress()
	if err != nil {
		return nil, err
	}
	txConfig := ctx.TxConfig
	txBuilder := txConfig.NewTxBuilder()
	txClient := sdktx.NewServiceClient(ctx)
	txBuilder.SetMsgs(msgs...)
	txBytes, _ := txConfig.TxEncoder()(txBuilder.GetTx())
	// simulate tx to get estimate gasLimit
	gasInfo, _ := txClient.Simulate(context.Background(), &sdktx.SimulateRequest{
		TxBytes: txBytes,
	})
	gasLimit := gasInfo.GasInfo.GasUsed
	gasLimitInt := sdkmath.NewIntFromUint64(gasLimit)
	// adjustment
	gasLimitInt = gasLimitInt.Add(sdkmath.NewInt(100000))
	l := len(n.Config.MinGasPrices) - len(n.Config.NativeDenom)
	fee := sdkmath.ZeroInt()
	if l > 0 {
		minGasPriceStr := n.Config.MinGasPrices[:l]
		minGasPrice, _ := sdkmath.NewIntFromString(minGasPriceStr)
		if err != nil {
			return nil, err
		}
		fee = gasLimitInt.Mul(minGasPrice)
	}

	txBuilder.SetGasLimit(gasLimit)
	txBuilder.SetFeeAmount(sdk.Coins{sdk.NewCoin(n.Config.NativeDenom, fee)})
	num, seq, err := ctx.AccountRetriever.GetAccountNumberSequence(ctx, acc)
	if err != nil {
		return nil, err
	}
	txf := tx.Factory{}.
		WithChainID(ctx.ChainID).
		WithKeybase(ctx.Keyring).
		WithTxConfig(txConfig).
		WithAccountNumber(num).
		WithSequence(seq)

	if err := tx.Sign(txf, keyName, txBuilder, true); err != nil {
		return nil, err
	}

	signedTx := txBuilder.GetTx()
	if txBytes, err = txConfig.TxEncoder()(signedTx); err != nil {
		return nil, err
	}

	return txBytes, nil
}

func (n *Network) SendTx(msgs []sdk.Msg, keyName string) (*ctypes.ResultBroadcastTx, error) {
	signedTxBytes, err := n.GenSignedTx(msgs, keyName)
	if err != nil {
		return nil, err
	}
	// txClient.BraodcastTx()
	return n.Validators[0].RPCClient.BroadcastTxSync(context.Background(), signedTxBytes)
}

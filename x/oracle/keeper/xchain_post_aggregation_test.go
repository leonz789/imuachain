package keeper_test

import (
	"encoding/json"
	"testing"

	keepertest "github.com/imua-xyz/imuachain/testutil/keeper"
	"github.com/imua-xyz/imuachain/x/oracle/keeper"
	"github.com/stretchr/testify/require"
)

func TestUpdateXChainMsgs_Basic(t *testing.T) {
	k, ctx := keepertest.OracleKeeper(t)

	rootHash := make([]byte, 32)
	for i := 0; i < 32; i++ {
		rootHash[i] = byte(i)
	}

	batch := keeper.RawDataXChainBatch{
		SrcChainID: 123,
		BatchSeq:   1,
		Messages: []keeper.RawDataXChainMsg{
			{ID: "tx1:0", Type: "evm", PayloadB64: ""},
			{ID: "tx1:0", Type: "evm", PayloadB64: ""}, // duplicate in same batch
		},
	}
	raw, err := json.Marshal(batch)
	require.NoError(t, err)

	err = keeper.UpdateXChainMsgs(ctx, rootHash, raw, 99, 7, k)
	require.NoError(t, err)

	lastSeq, found := k.GetXChainLastSeq(ctx, 123)
	require.True(t, found)
	require.EqualValues(t, 1, lastSeq)
	require.True(t, k.HasXChainMsgProcessed(ctx, 123, "tx1:0"))

	// replay same batch should be idempotent
	err = keeper.UpdateXChainMsgs(ctx, rootHash, raw, 99, 7, k)
	require.NoError(t, err)
}

func TestUpdateXChainMsgs_SeqGap(t *testing.T) {
	k, ctx := keepertest.OracleKeeper(t)

	rootHash := make([]byte, 32)

	batch1 := keeper.RawDataXChainBatch{
		SrcChainID: 1,
		BatchSeq:   1,
		Messages:   []keeper.RawDataXChainMsg{{ID: "a", Type: "evm"}},
	}
	raw1, _ := json.Marshal(batch1)
	require.NoError(t, keeper.UpdateXChainMsgs(ctx, rootHash, raw1, 1, 1, k))

	// gap: try to apply seq=3 while last=1
	batch3 := keeper.RawDataXChainBatch{
		SrcChainID: 1,
		BatchSeq:   3,
		Messages:   []keeper.RawDataXChainMsg{{ID: "b", Type: "evm"}},
	}
	raw3, _ := json.Marshal(batch3)
	err := keeper.UpdateXChainMsgs(ctx, rootHash, raw3, 1, 2, k)
	require.Error(t, err)
}

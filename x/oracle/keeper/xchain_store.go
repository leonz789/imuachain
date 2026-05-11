package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/imua-xyz/imuachain/x/oracle/types"
)

// Decode failures on stored uint64 values indicate state corruption (we only
// ever Set via types.Uint64Bytes); panic rather than masking as missing, which
// would let an accepted batch_seq look new and bypass replay protection.

func (k Keeper) GetXChainLastSeq(ctx sdk.Context, srcChainID uint64) (uint64, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.XChainLastSeqKey(srcChainID))
	if bz == nil {
		return 0, false
	}
	seq, err := types.BytesToUint64(bz)
	if err != nil {
		panic(fmt.Sprintf("xchain_store: corrupted XChainLastSeq for srcChainID=%d (len=%d): %s",
			srcChainID, len(bz), err))
	}
	return seq, true
}

func (k Keeper) GetXChainLastExecutedSeq(ctx sdk.Context, srcChainID uint64) (uint64, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.XChainLastExecutedSeqKey(srcChainID))
	if bz == nil {
		return 0, false
	}
	seq, err := types.BytesToUint64(bz)
	if err != nil {
		panic(fmt.Sprintf("xchain_store: corrupted XChainLastExecutedSeq for srcChainID=%d (len=%d): %s",
			srcChainID, len(bz), err))
	}
	return seq, true
}

func (k Keeper) SetXChainLastSeq(ctx sdk.Context, srcChainID, seq uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.XChainLastSeqKey(srcChainID), types.Uint64Bytes(seq))
}

func (k Keeper) SetXChainLastExecutedSeq(ctx sdk.Context, srcChainID, seq uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.XChainLastExecutedSeqKey(srcChainID), types.Uint64Bytes(seq))
}

func (k Keeper) HasXChainMsgProcessed(ctx sdk.Context, srcChainID uint64, msgID string) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.XChainMsgProcessedKey(srcChainID, msgID))
}

func (k Keeper) SetXChainMsgProcessed(ctx sdk.Context, srcChainID uint64, msgID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.XChainMsgProcessedKey(srcChainID, msgID), []byte{1})
}

func (k Keeper) GetXChainMsgRetryCount(ctx sdk.Context, srcChainID uint64, msgID string) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.XChainMsgRetryKey(srcChainID, msgID))
	if bz == nil {
		return 0
	}
	v, err := types.BytesToUint64(bz)
	if err != nil {
		panic(fmt.Sprintf("xchain_store: corrupted XChainMsgRetryCount for srcChainID=%d msgID=%s (len=%d): %s",
			srcChainID, msgID, len(bz), err))
	}
	return v
}

func (k Keeper) SetXChainMsgRetryCount(ctx sdk.Context, srcChainID uint64, msgID string, count uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.XChainMsgRetryKey(srcChainID, msgID), types.Uint64Bytes(count))
}

func (k Keeper) ClearXChainMsgRetryCount(ctx sdk.Context, srcChainID uint64, msgID string) {
	ctx.KVStore(k.storeKey).Delete(types.XChainMsgRetryKey(srcChainID, msgID))
}

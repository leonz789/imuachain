package keeper

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/imua-xyz/imuachain/x/oracle/keeper/common"
	"github.com/imua-xyz/imuachain/x/oracle/types"
)

// RawDataXChainBatch is the rawData payload schema for cross-chain message batches.
// It is intentionally JSON to avoid proto regeneration while bootstrapping the cross-chain pipeline.
type RawDataXChainBatch struct {
	SrcChainID uint64             `json:"src_chain_id"`
	BatchSeq   uint64             `json:"batch_seq"`
	Messages   []RawDataXChainMsg `json:"messages"`
}

type RawDataXChainMsg struct {
	// ID must be globally unique for replay protection (txHash:logIndex).
	ID string `json:"id"`
	// Type is an application-level discriminator (e.g. "evm").
	Type string `json:"type"`
	// PayloadB64 carries the message bytes. The handler only validates base64; execution is out of scope.
	PayloadB64 string `json:"payload_b64"`
}

var _ common.PostAggregationHandler = UpdateXChainMsgs

// UpdateXChainMsgs is a post-aggregation handler for oracle 2-phases "cross-chain message batch" feeders.
//
// Semantics (minimal bootstrap version):
// - Enforces strict batch sequencing per srcChainID (batch_seq must equal last_seq+1).
// - Provides replay protection per message ID (idempotent).
// - Does NOT execute message payloads yet; it only records processed IDs and emits an event.
//
// NOTE: postHandler errors are only logged (they do not revert the block).
func UpdateXChainMsgs(
	ctx sdk.Context,
	rootHash []byte,
	rawData []byte,
	feederID, roundID uint64,
	kInf common.KeeperOracle,
) error {
	k, ok := kInf.(*Keeper)
	if !ok {
		return errors.New("input keeper interface type error")
	}

	var batch RawDataXChainBatch
	if err := json.Unmarshal(rawData, &batch); err != nil {
		return fmt.Errorf("invalid xchain rawData json: %w", err)
	}
	if batch.SrcChainID == 0 {
		return errors.New("invalid xchain rawData: src_chain_id must be > 0")
	}
	if batch.BatchSeq == 0 {
		return errors.New("invalid xchain rawData: batch_seq must be > 0")
	}

	lastSeq, found := k.GetXChainLastSeq(ctx, batch.SrcChainID)
	if !found {
		lastSeq = 0
	}
	// Idempotency: ignore already-processed (or older) batches.
	if batch.BatchSeq <= lastSeq {
		return nil
	}
	expected := lastSeq + 1
	if batch.BatchSeq != expected {
		return fmt.Errorf("xchain batch seq mismatch: srcChainID:%d expected:%d got:%d", batch.SrcChainID, expected, batch.BatchSeq)
	}

	// Mark unique message IDs as processed (idempotent).
	unique := make(map[string]struct{}, len(batch.Messages))
	processedNew := 0
	for _, m := range batch.Messages {
		if m.ID == "" {
			return errors.New("invalid xchain message: empty id")
		}
		if _, seen := unique[m.ID]; seen {
			continue
		}
		unique[m.ID] = struct{}{}

		// Validate payload encoding (for now we don't execute it).
		if m.PayloadB64 != "" {
			if _, err := base64.StdEncoding.DecodeString(m.PayloadB64); err != nil {
				return fmt.Errorf("invalid xchain message payload_b64 for id:%s: %w", m.ID, err)
			}
		}

		if k.HasXChainMsgProcessed(ctx, batch.SrcChainID, m.ID) {
			continue
		}
		k.SetXChainMsgProcessed(ctx, batch.SrcChainID, m.ID)
		processedNew++
	}

	k.SetXChainLastSeq(ctx, batch.SrcChainID, batch.BatchSeq)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeCreatePrice,
		sdk.NewAttribute(types.AttributeKeyFeederID, strconv.FormatUint(feederID, 10)),
		sdk.NewAttribute(types.AttributeKeyRoundID, strconv.FormatUint(roundID, 10)),
		sdk.NewAttribute(types.AttributeKeyXChainSrcChainID, strconv.FormatUint(batch.SrcChainID, 10)),
		sdk.NewAttribute(types.AttributeKeyXChainBatchSeq, strconv.FormatUint(batch.BatchSeq, 10)),
		sdk.NewAttribute(types.AttributeKeyXChainMsgCount, strconv.Itoa(processedNew)),
		sdk.NewAttribute(types.AttributeKeyXChainRootHash, base64.StdEncoding.EncodeToString(rootHash)),
	))

	return nil
}

package feedermanagement

import (
	"encoding/base64"
	"strconv"
	"strings"

	oracletypes "github.com/ExocoreNetwork/exocore/x/oracle/types"
	"github.com/ethereum/go-ethereum/common"
)

func (f *FeederManager) RawDataCollecting(feederID uint64) bool {
	// TODO(leonz): implement me
	// if r.m.latestIndex == pieceCount return false
	return false
}

// func (f *FeederManager) LatestPieceIndexForTokenFeederID(feederID uint64) uint64 {
// 	// TODO(leonz): implement me
// 	return 0
// }

func (f *FeederManager) NextPieceIndexByFeederID(feederID uint64) (uint32, bool) {
	r, ok := f.rounds[int64(feederID)]
	if !ok || r.m == nil {
		return 0, false
	}
	return r.m.NextPieceIndex()
}

func (f *FeederManager) MaxPieceIndexForTokenFeederID(feederID uint64) (uint32, bool) {
	r, ok := f.rounds[int64(feederID)]
	if !ok {
		return 0, false
	}
	if r.m == nil || r.m.LeafCount() < 1 {
		return 0, false
	}
	return r.m.LeafCount() - 1, true
}

// VerifyPieceProofsForTokenFeederID verifies targetPiece against feederID's corresponding merkle root, it might need proof nodes from pieces
func (f *FeederManager) VerifyPieceProofsForTokenFeederID(feederID uint64, targetPiece *oracletypes.PieceWithProof, pieces ...*oracletypes.PieceWithProof) bool {
	// TODO(leonz): implement me
	// TODO(opt. leonz): restrict check the targetPiece has exactly count of hash nodes that needed without any redundant nodes, that way we can limit the tx size efficiently
	return false
}

// GetPieceWithProof verify the message is a valid rawData message and parse the piece of rawData
func (f *FeederManager) GetPieceWithProof(msg *oracletypes.MsgCreatePrice) (*oracletypes.PieceWithProof, bool) {
	// TODO: remove comments after update proto
	//	if !msg.RawData {
	//		return false, nil
	//	}
	if !f.cs.IsRule2PhasesByFeederID(msg.FeederID) {
		return nil, false
	}

	r := f.rounds[int64(msg.FeederID)]
	if r == nil {
		return nil, false
	}
	pieceCount := r.PieceCount()

	if len(msg.Prices) != 1 || len(msg.Prices[0].Prices) < 1 || len(msg.Prices[0].Prices) > 2 {
		return nil, false
	}

	pieceStr := msg.Prices[0].Prices[0].Price
	if len(pieceStr) == 0 || len(pieceStr) > int(f.cs.RawDataPieceSize()) {
		return nil, false
	}

	pieceIndex, err := strconv.ParseUint(msg.Prices[0].Prices[0].DetID, 10, 32)
	if err != nil || pieceIndex > pieceCount || pieceIndex < 1 {
		return nil, false
	}

	ret := &oracletypes.PieceWithProof{
		Index:   uint32(pieceIndex),
		RawData: []byte(pieceStr),
	}

	if len(msg.Prices[0].Prices) == 2 {
		joinedHashesBase64 := msg.Prices[0].Prices[1].Price
		joinedIndexes := msg.Prices[0].Prices[1].DetID
		if len(joinedHashesBase64) == 0 || len(joinedIndexes) == 0 {
			return nil, false
		}

		hashesBase64 := strings.Split(joinedHashesBase64, oracletypes.DelimiterForBase64)
		indexes := strings.Split(joinedIndexes, oracletypes.DelimiterForBase64)

		if len(hashesBase64) == 0 || len(hashesBase64) != len(indexes) {
			return nil, false
		}

		proof := make([]*oracletypes.HashNode, 0, len(hashesBase64))
		for i, hashBase64 := range hashesBase64 {
			hashBytes, err := base64.StdEncoding.DecodeString(hashBase64)
			if err != nil || len(hashBytes) != common.HashLength {
				return nil, false
			}
			index, err := strconv.ParseUint(indexes[i], 10, 32)
			if err != nil || index < 1 || index > pieceCount {
				return nil, false
			}
			proof = append(proof, &oracletypes.HashNode{Index: uint32(index), Hash: hashBytes})
		}
		ret.Proof = proof
	}
	return ret, true
}

// GetTidyProofPathByIndex return the proof path with unseen nodes that under condition pieces comes as index from 0 to n, and cached all seen proof nodes
func (f *FeederManager) MinimalProofPathByIndex(feederID uint64, index uint32) []uint32 {
	r, ok := f.rounds[int64(feederID)]
	if !ok {
		return nil
	}
	if r.m == nil {
		return nil
	}
	return r.m.MinimalProofPathByIndex(index)
}

func (f *FeederManager) FeederIDsCollectingRawData() []uint64 {
	// TODO(leonz): implement me
	return nil
}

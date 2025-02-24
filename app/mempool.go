package app

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"

	"github.com/ExocoreNetwork/exocore/app/ante/utils"
	oraclekeeper "github.com/ExocoreNetwork/exocore/x/oracle/keeper"
	oracletypes "github.com/ExocoreNetwork/exocore/x/oracle/types"
)

type ExocoreMempool struct {
	// feederID -> []PieceWithProof, cached peiceWithProof for feederID
	// cachedPieces map[uint64][]*oracletypes.PieceWithProof
	//	verifiedMsgPieces map[uint64][]*oracletypes.MsgCreatePriceRawData
	// cachedPieces map[uint64][]*oracletypes.PieceWithProof
	cachedPieces map[uint64]map[uint32][]*oracletypes.PieceWithProof
	k            oraclekeeper.Keeper
	count        int
	// load from config file as a config for mempool
	//cacheWindow int
	txDecoder sdk.TxDecoder
}

// Insert inserts a tx into mempool, currently only used for rawData from tx related to 2-phases aggregation of oracle module
func (em *ExocoreMempool) Insert(ctx context.Context, tx sdk.Tx) error {
	// we don't filter tx not with message of rawData type, those tx will just be added into tendermint's txpool
	if !em.includesMsgOracle(tx) {
		return nil
	}

	piece, msgOracle, _, isRawData := em.GetPieceWithProof(tx)
	// we don't filter tx not with message of rawData type
	if !isRawData {
		return nil
	}
	if piece == nil {
		return errors.New("failed to parse pieceWithProof from tx with oracle rawData message")
	}
	fID := msgOracle.FeederID

	piecesCached, ok := em.cachedPieces[fID]
	if !ok {
		em.cachedPieces[fID] = make(map[uint32][]*oracletypes.PieceWithProof)
		em.cachedPieces[fID][piece.Index] = []*oracletypes.PieceWithProof{piece}
		em.count++
		return nil
	}
	piecesIndexCached, ok := piecesCached[piece.Index]
	if !ok {
		piecesCached[piece.Index] = []*oracletypes.PieceWithProof{piece}
		em.count++
		return nil
	}
	for _, pieceCached := range piecesIndexCached {
		if pieceCached.EqualsTo(piece) {
			return errors.New("piece exists in mempool")
		}
	}
	piecesCached[piece.Index] = append(piecesIndexCached, piece)
	em.count++
	return nil

	//	return em.addPiece(msgOracle.FeederID, piece)

	//	if em.cacheIsFull(fID) {
	//		return fmt.Errorf("mempool is full for feederID:%d", fID)
	//	}
	//	pieceIndex := piece.Index
	//	// for each piece, we only keep one copy without considering the submitter(validator)
	//	if em.pieceExist(fID, pieceIndex) {
	//		return fmt.Errorf("piece:%d for feederID:%d exists", fID, pieceIndex)
	//	}
	//
	//	// we don't check if the rawData collection has completed, it is done by anteHandler
	//	//	nextPieceIndex := em.k.FeederManager.LatestPieceIndexForTokenFeederID(fID) + 1
	//	nextPieceIndex, found := em.k.FeederManager.NextPieceIndexByFeederID(fID)
	//	if !found {
	//		// this should not happen, since we checked in anteHandler already, TODO(leonz): remove ?
	//		return fmt.Errorf("nextPieceIndex for feederID:%d not found", fID)
	//	}
	//	if pieceIndex < nextPieceIndex {
	//		return fmt.Errorf("piece:%d for feederID:%d has been collected onchain", pieceIndex, fID)
	//	}
	//
	//	if pieceIndex == nextPieceIndex {
	//		// verify on committed state and cache in mempool if it's verified
	//		if ok := em.k.FeederManager.VerifyPieceProofsForTokenFeederID(fID, piece); !ok {
	//			return fmt.Errorf("piece:%d for feederID:%d verify failed against root", pieceIndex, fID)
	//		}
	//		em.addPiece(fID, piece)
	//		return nil
	//	}
	//
	//	if pieceIndex == em.latestPieceIndex(fID)+1 {
	//		// pieceIndex > nextPieceIndex from committed state but can be verified with mempool cachec piece
	//		// we keep some future pieces to make sure proposer is able to propose a block with raw data piece when necessary
	//		// if ok := em.k.FeederManager.VerifyPieceProofsForTokenFeederID(fID, piece, em.pieceList(fID)...); !ok {
	//		if ok := em.k.FeederManager.VerifyPieceProofsForTokenFeederID(fID, piece, em.cachedPieces[fID]...); !ok {
	//			return fmt.Errorf("piece:%d for feederID:%d verify failed against root", pieceIndex, fID)
	//		}
	//		//em.addMsgPiece(fID, msg)
	//		em.addPiece(fID, piece)
	//	}
}

func (em *ExocoreMempool) Select(ctx context.Context, txList [][]byte) mempool.Iterator {
	// feederIDS:[]uint64, which are expecting rawData
	// []Tx, each feederID must have one tx
	//	feederIDS := em.FeederManager.FeederIDsCollectingRawData()
	//	newTxList := make([][]byte, 0, len(txList))
	//	txRawData := make(map[uint64])
	//	return nil
	return nil
}

func (em *ExocoreMempool) CountTx() int {
	return em.count
}

func (em *ExocoreMempool) Remove(tx sdk.Tx) error {
	// TODO(leonz): clear history sealed round on block change
	// remove all pieces with indexes <= piece index of input msg
	piece, msgOracle, _, isRawData := em.GetPieceWithProof(tx)
	// we don't process tx not with rawdata
	if !isRawData {
		return nil
	}
	piecesCached, ok := em.cachedPieces[msgOracle.FeederID]
	if !ok {
		return nil
	}
	newPiecesCached := make(map[uint32][]*oracletypes.PieceWithProof)
	for index, piecesIndexCached := range piecesCached {
		if index > piece.Index {
			newPiecesCached[index] = piecesIndexCached
		} else {
			em.count -= len(piecesIndexCached)
		}
	}
	if len(newPiecesCached) < len(piecesCached) {
		em.cachedPieces[msgOracle.FeederID] = newPiecesCached
	}
	return nil

	//	list, ok := em.cachedPieces[msgOracle.FeederID]
	//	if !ok {
	//		return nil
	//	}
	//
	//	newList := make([]*oracletypes.PieceWithProof, 0, len(list))
	//	// Remove is called during DeliverTx, when a index is included in block, all index LET it should be remved
	//	for _, p := range list {
	//		if p.Index > piece.Index {
	//			newList = append(newList, p)
	//		}
	//	}
	//
	//	if len(newList) == 0 {
	//		delete(em.cachedPieces, msgOracle.FeederID)
	//	} else {
	//		em.cachedPieces[msgOracle.FeederID] = newList
	//	}
	//
	//	em.count -= (len(list) - len(newList))
}

// check if piece with the same id from input piece exists in mempool cache
// func (em *ExocoreMempool) pieceExist(feederID uint64, pIndex uint32) bool {
// 	if list, ok := em.cachedPieces[feederID]; ok {
// 		for _, p := range list {
// 			if p.Index == pIndex {
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }

// func (em *ExocoreMempool) cacheIsFull(feederID uint64) bool {
// 	if list, ok := em.cachedPieces[feederID]; ok {
// 		return len(list) == em.cacheWindow
// 	}
// 	return false
// }

// func (em *ExocoreMempool) latestPieceIndex(feederID uint64) uint32 {
// 	if list, ok := em.cachedPieces[feederID]; ok {
// 		return list[len(list)-1].Index
// 	}
// 	return 0
// }

// func (em *ExocoreMempool) addPiece(feederID uint64, piece *oracletypes.PieceWithProof) {
// 	em.cachedPieces[feederID] = append(em.cachedPieces[feederID], piece)
// 	em.count++
// }

func (em *ExocoreMempool) includesMsgOracle(tx sdk.Tx) bool {
	msgs := tx.GetMsgs()
	for _, msg := range msgs {
		if _, ok := msg.(*oracletypes.MsgCreatePrice); ok {
			return true
		}
	}
	return false
}

func (em *ExocoreMempool) GetPieceWithProof(tx sdk.Tx) (pieceWithProof *oracletypes.PieceWithProof, msgO *oracletypes.MsgCreatePrice, isOracle, isRawData bool) {
	var msgs []*oracletypes.MsgCreatePrice
	msgs, isOracle, isRawData = utils.OracleCreatePriceTx(tx)
	if !isRawData {
		return
	}
	msgO = msgs[0]
	// msgs := tx.GetMsgs()
	// if len(msgs) > 1 {
	// 	return nil, nil, errors.New("tx includes more than one message")
	// }
	// msgOracle, ok := msgs[0].(*oracletypes.MsgCreatePrice)
	// if !ok {
	// 	return nil, nil, errors.New("tx include message not of create-price type")
	// }
	// if !em.k.FeederManager.RawDataCollecting(msgOracle.FeederID) {
	// 	return nil, nil, fmt.Errorf("raw data is not being collected for feederID:%d", msgOracle.FeederID)
	// }

	pieceWithProof, _ = em.k.FeederManager.GetPieceWithProof(msgO)
	if pieceWithProof != nil {
		pieceWithProof.Tx = tx
	}
	return
	//	if !ok {
	//		return nil, nil, errors.New("failed to get piece info from message")
	//	}
	//
	//	piece.Tx = tx

	// return piece, msgOracle, nil
}

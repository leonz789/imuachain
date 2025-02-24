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
}

func (em *ExocoreMempool) Select(ctx context.Context, txList [][]byte) mempool.Iterator {
	// feederIDS:[]uint64, which are expecting rawData
	// []Tx, each feederID must have one tx
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
}

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

	pieceWithProof, _ = em.k.FeederManager.GetPieceWithProof(msgO)
	if pieceWithProof != nil {
		pieceWithProof.Tx = tx
	}
	return
}

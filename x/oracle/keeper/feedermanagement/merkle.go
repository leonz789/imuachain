package feedermanagement

// import "github.com/ethereum/go-ethereum/common"
//
// func (m *Merkle) NewMerkle(totalSize, pieceSize int, root [32]byte) *Merkle {
// 	// TODO(leonz): implement me
// 	return nil
// }
//
// // VerifyAndCache verifies a data piece with index against merkle root, return values:
// // verified: the data piece is verified
// // idxExist: the index of input data piece exists in cache already
// // completedData: if all pieces had been cached already, return it as completedData
// // addedProofPath: which hash nodes from input proof are newly added to the cache (has not seen before)
// func (m *Merkle) VerifyAndCache(idx uint64, datePiece []byte, proof map[uint64]common.Hash) (verified bool, idxExist bool, completedData []byte, addedProofPath []uint64) {
// 	// TODO(leonz): implement me
// 	return false, false, nil, nil
// }
//
// // UncachedPathForIndex returns the hash node indexes which are not cached yet for the input index
// func (m *Merkle) UncachedPathForIndex(idx uint64) []uint64 {
// 	// TODO(leonz): implement me
//
// 	return nil
// }
//
// // ResetRoot set merkle root and totalsize for a merkle tree, this will clear all cached values in the tree if not empty
// func (m *Merkle) ResetRoot(totalSize int64, root common.Hash) bool {
// 	// TODO(leonz): implement me
// 	// if not empty: m.Reset() first
// 	return false
// }
//
// func (m *Merkle) Reset() bool {
// 	// TODO(leonz): implement me
// 	return false
// }
//
// func (m *Merkle) ResetPieceSize(size int64) bool {
// 	// TODO(leonz): implement me
// 	return false
// }

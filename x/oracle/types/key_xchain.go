package types

import "crypto/sha256"

const (
	XChainPrefix        = "XChain/"
	XChainLastSeqPrefix = XChainPrefix + "lastSeq/"
	XChainMsgPrefix     = XChainPrefix + "msg/"
)

func XChainLastSeqKey(srcChainID uint64) []byte {
	return append([]byte(XChainLastSeqPrefix), Uint64Bytes(srcChainID)...)
}

// XChainMsgProcessedKey builds a stable key for a processed message under a source chain.
// We hash msgID to keep keys short and avoid key-space abuse.
func XChainMsgProcessedKey(srcChainID uint64, msgID string) []byte {
	sum := sha256.Sum256([]byte(msgID))
	key := make([]byte, 0, len(XChainMsgPrefix)+8+len(sum))
	key = append(key, []byte(XChainMsgPrefix)...)
	key = append(key, Uint64Bytes(srcChainID)...)
	key = append(key, sum[:]...)
	return key
}

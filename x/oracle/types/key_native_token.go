package types

import (
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	// NativeTokenKeyPrefix is the prefix to retrieve all NativeToken
	NativeTokenKeyPrefix           = "NativeToken/"
	NativeTokenPriceKeyPrefix      = NativeTokenKeyPrefix + "price/value/"
	NativeTokenStakerInfoKeyPrefix = NativeTokenKeyPrefix + "stakerInfo/value/"
	NativeTokenStakerListKeyPrefix = NativeTokenKeyPrefix + "stakerList/value/"
	NativeTokenVersionKeyPrefix    = NativeTokenKeyPrefix + "version/"

	NSTKeyPrefix                = "NST/"
	NSTBalancesKeyPrefix        = NSTKeyPrefix + "balance/"
	NSTStakerAddrKeyPrefix      = NSTKeyPrefix + "stakerAddr/"
	NSTLastStakerIndexKeyPrefix = NSTKeyPrefix + "lastStakerIndex/"
	NSTStakerKeyPrefix          = NSTKeyPrefix + "staker/"
	NSTVersionKeyPrefix         = NSTKeyPrefix + "version/"
)

func NSTStakerAddrKeyChainIDPrefix(chainID uint64) []byte {
	chainIDStr := hexutil.EncodeUint64(chainID)
	return append([]byte(NSTStakerAddrKeyPrefix), []byte(chainIDStr+"/")...)
}

func NSTStakerAddrKey(chainID uint64, stakerIndex uint32) []byte {
	chainIDStr := hexutil.EncodeUint64(chainID)
	return append([]byte(NSTStakerAddrKeyPrefix), []byte(chainIDStr+"/"+hexutil.EncodeUint64(uint64(stakerIndex)))...)
}

func NSTVersionKey(chainID uint64) []byte {
	chainIDStr := hexutil.EncodeUint64(chainID)
	return append([]byte(NSTVersionKeyPrefix), []byte(chainIDStr)...)
}

func NSTBalancesKeyChainIDPrefix(chainID uint64) []byte {
	chainIDStr := hexutil.EncodeUint64(chainID)
	return append([]byte(NSTBalancesKeyPrefix), []byte(chainIDStr+"/")...)
}

func NSTBalancesKey(chainID uint64, addr string) []byte {
	chainIDStr := hexutil.EncodeUint64(chainID)
	return append([]byte(NSTBalancesKeyPrefix), []byte(chainIDStr+"/"+addr)...)
}

func NSTStakerKeyChainIDPrefix(chainID uint64) []byte {
	chainIDStr := hexutil.EncodeUint64(chainID)
	return append([]byte(NSTStakerKeyPrefix), []byte(chainIDStr+"/")...)
}

func NSTStakerKey(chainID uint64, addr string) []byte {
	chainIDStr := hexutil.EncodeUint64(chainID)
	return append([]byte(NSTStakerKeyPrefix), []byte(chainIDStr+"/"+addr)...)
}

func NSTLatestStakerIndexKey(chainID uint64) []byte {
	chainIDStr := hexutil.EncodeUint64(chainID)
	return append([]byte(NSTLastStakerIndexKeyPrefix), []byte(chainIDStr)...)
}

// NativeTokenStakerKeyPrefix returns the prefix for stakerInfo key
// NativetToken/stakerInfo/value/assetID/
func NativeTokenStakerKeyPrefix(assetID string) []byte {
	if len(assetID) == 0 {
		return []byte(NativeTokenStakerInfoKeyPrefix)
	}
	assetID += "/"
	return append([]byte(NativeTokenStakerInfoKeyPrefix), []byte(assetID)...)
}

// NativeTokenStakerKey returns stakerKey
// NativeToken/stakerInfo/value/assetID/stakerAddr
func NativeTokenStakerKey(assetID, stakerAddr string) []byte {
	return append(NativeTokenStakerKeyPrefix(assetID), []byte(stakerAddr)...)
}

// NativeTokenStakerListKey returns stakerList key
// NativeToken/stakerList/value/assetID
func NativeTokenStakerListKey(assetID string) []byte {
	return append([]byte(NativeTokenStakerListKeyPrefix), []byte(assetID)...)
}

// ParseNativeTokenStakerKey retieve assetID and stakerAddr from stakerInfoKey
// assetID/stakerAddr -> {assetID, stakerAddr}
func ParseNativeTokenStakerKey(key []byte) (assetID, stakerAddr string) {
	parsed := strings.Split(string(key), "/")
	if len(parsed) != 2 {
		panic("key of stakerInfo must be construct by 2 infos: assetID/stakerAddr")
	}
	return parsed[0], parsed[1]
}

func NativeTokenVersionKey(assetID string) []byte {
	return append([]byte(NativeTokenVersionKeyPrefix), []byte(assetID)...)
}

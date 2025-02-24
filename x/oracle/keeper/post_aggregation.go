package keeper

import (
	"github.com/ExocoreNetwork/exocore/x/oracle/keeper/common"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// the input data could be either rawData bytes of data with big size for non-price senarios or 'price' info
type PostAggregationHandler func(data []byte, ctx sdk.Context, k common.KeeperOracle) error

// RegisterPostAggregation registers handler for tokenfeeder set with deterministic source which need to do some process with the deterministic aggregated result
// this is used to register the post handlers served for some customer defined deterministic source oracle requirement
func (k *Keeper) RegisterPostAggregation() {
	// k.BondPostAggregation(1, UpdateNSTBalanceChange)
}

func (k *Keeper) BondPostAggregation(feederID int64, postHandler PostAggregationHandler) {
	k.postHandlers[feederID] = postHandler
}

func (k *Keeper) GetPostAggregation(feederID int64) (handler PostAggregationHandler, found bool) {
	if k.postHandlers == nil {
		return nil, false
	}
	handler, found = k.postHandlers[feederID]
	return
}

package network

import (
	sdkmath "cosmossdk.io/math"
	assetstypes "github.com/ExocoreNetwork/exocore/x/assets/types"
	delegationtypes "github.com/ExocoreNetwork/exocore/x/delegation/types"
	dogfoodtypes "github.com/ExocoreNetwork/exocore/x/dogfood/types"
	operatortypes "github.com/ExocoreNetwork/exocore/x/operator/types"
)

var (
	DefaultGenStateAssets = assetstypes.GenesisState{
		Params: assetstypes.Params{
			ExocoreLzAppAddress:    "0x3e108c058e8066da635321dc3018294ca82ddedf",
			ExocoreLzAppEventTopic: assetstypes.DefaultParams().ExocoreLzAppEventTopic,
		},
		ClientChains: []assetstypes.ClientChainInfo{
			{
				Name:             "Example EVM chain",
				MetaInfo:         "Example EVM chain metaInfo",
				LayerZeroChainID: 101,
				AddressLength:    20,
			},
		},
		Tokens: []assetstypes.StakingAssetInfo{
			{
				// for test this token will be registered on ETH in oracle module
				AssetBasicInfo: assetstypes.AssetInfo{
					Name:             "Tether USD",
					MetaInfo:         "Tether USD token",
					Address:          "0xdac17f958d2ee523a2206206994597c13d831ec7",
					LayerZeroChainID: 101,
				},
				StakingTotalAmount: sdkmath.NewInt(5000),
			},
			{
				AssetBasicInfo: assetstypes.AssetInfo{
					Name:             "NST ETH",
					MetaInfo:         "native restaking ETH",
					Address:          "0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
					LayerZeroChainID: 101,
				},
				StakingTotalAmount: sdkmath.NewInt(5000),
			},
		},
	}

	// DefaultGenStateOperator = operatortypes.GenesisState{}
	DefaultGenStateOperator = *operatortypes.DefaultGenesis()

	// DefaultGenStateDelegation = delegationtypes.GenesisState{}
	DefaultGenStateDelegation = *delegationtypes.DefaultGenesis()

	DefaultGenStateDogfood = *dogfoodtypes.DefaultGenesis()
)

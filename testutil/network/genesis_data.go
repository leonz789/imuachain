package network

import (
	sdkmath "cosmossdk.io/math"
	assetstypes "github.com/ExocoreNetwork/exocore/x/assets/types"
	delegationtypes "github.com/ExocoreNetwork/exocore/x/delegation/types"
	dogfoodtypes "github.com/ExocoreNetwork/exocore/x/dogfood/types"
	operatortypes "github.com/ExocoreNetwork/exocore/x/operator/types"
	oracletypes "github.com/ExocoreNetwork/exocore/x/oracle/types"
)

var (
	// DefaultGenStateAssets only includes two assets, one for ETH and the other for NST ETH
	// For the contract address of asset-ETH we filled with the address of USDT, that's ok for test
	// we bond both tokens to the price of ETH in oracle module
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
					Name:     "ETH",
					MetaInfo: "Ethereum native token",
					// the address is of USDT_Etheruem, but that's ok
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

	DefaultGenStateOperator = operatortypes.GenesisState{}
	// DefaultGenStateOperator = *operatortypes.DefaultGenesis()

	DefaultGenStateDelegation = delegationtypes.GenesisState{}
	// DefaultGenStateDelegation = *delegationtypes.DefaultGenesis()

	DefaultGenStateDogfood = *dogfoodtypes.DefaultGenesis()

	DefaultGenStateOracle = *oracletypes.DefaultGenesis()
)

func init() {
	// bond assetsIDs of ETH, NSTETH to ETH price
	DefaultGenStateOracle.Params.Tokens[1].AssetID = "0xdac17f958d2ee523a2206206994597c13d831ec7_0x65,0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee_0x65"
	// set ETH tokenfeeder's 'StartBaseBlock' to 10
	DefaultGenStateOracle.Params.TokenFeeders[1].StartBaseBlock = 10
}

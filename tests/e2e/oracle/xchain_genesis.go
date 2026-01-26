package oracle

import (
	"strings"

	"github.com/imua-xyz/imuachain/testutil/network"
	oracletypes "github.com/imua-xyz/imuachain/x/oracle/types"
)

func ensureXChainGenesis() {
	params := &network.DefaultGenStateOracle.Params
	for i, token := range params.Tokens {
		if strings.HasPrefix(strings.ToLower(token.AssetID), oracletypes.XChainIDPrefix) {
			ensureXChainTokenPrice(uint64(i))
			ensureXChainGatewayGenesis()
			return
		}
	}

	tokenID := uint64(len(params.Tokens))
	params.Tokens = append(params.Tokens, &oracletypes.Token{
		Name:            "XCHAIN_TEST",
		ChainID:         1,
		ContractAddress: "0x",
		Decimal:         0,
		Active:          true,
		AssetID:         "xchain_101",
	})
	params.TokenFeeders = append(params.TokenFeeders, &oracletypes.TokenFeeder{
		TokenID:        tokenID,
		RuleID:         3,
		StartRoundID:   1,
		StartBaseBlock: 5,
		Interval:       10,
	})

	ensureXChainTokenPrice(tokenID)
	ensureXChainGatewayGenesis()
}

func ensureXChainTokenPrice(tokenID uint64) {
	for _, prices := range network.DefaultGenStateOracle.PricesList {
		if prices.TokenID == tokenID {
			return
		}
	}
	network.DefaultGenStateOracle.PricesList = append(network.DefaultGenStateOracle.PricesList, oracletypes.Prices{
		TokenID:     tokenID,
		NextRoundID: 2,
		PriceList: []*oracletypes.PriceTimeRound{
			{Price: "1", Decimal: 0, RoundID: 1},
		},
	})
}

func ensureXChainGatewayGenesis() {
	gateway := strings.ToLower(network.ExpectedOracleGatewayAddress().Hex())
	params := &network.DefaultGenStateAssets.Params
	if len(params.Gateways) == 0 {
		params.Gateways = []string{gateway}
		return
	}
	// Ensure the oracle gateway is the first entry (used by oracle deliverer).
	newGateways := []string{gateway}
	for _, addr := range params.Gateways {
		if strings.ToLower(addr) == gateway {
			continue
		}
		newGateways = append(newGateways, addr)
	}
	params.Gateways = newGateways
}

package network

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	sdkmath "cosmossdk.io/math"
	tmos "github.com/cometbft/cometbft/libs/os"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	pvm "github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	"github.com/cometbft/cometbft/rpc/client/local"
	"github.com/cometbft/cometbft/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/ethclient"

	assetstypes "github.com/ExocoreNetwork/exocore/x/assets/types"
	delegationtypes "github.com/ExocoreNetwork/exocore/x/delegation/types"
	dogfoodtypes "github.com/ExocoreNetwork/exocore/x/dogfood/types"
	operatortypes "github.com/ExocoreNetwork/exocore/x/operator/types"
	oracletypes "github.com/ExocoreNetwork/exocore/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	servergrpc "github.com/cosmos/cosmos-sdk/server/grpc"
	srvtypes "github.com/cosmos/cosmos-sdk/server/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/evmos/evmos/v16/server"
	evmostypes "github.com/evmos/evmos/v16/types"
	evmtypes "github.com/evmos/evmos/v16/x/evm/types"
)

func startInProcess(cfg Config, val *Validator) error {
	logger := val.Ctx.Logger
	tmCfg := val.Ctx.Config
	tmCfg.Instrumentation.Prometheus = false

	if err := val.AppConfig.ValidateBasic(); err != nil {
		return err
	}

	nodeKey, err := p2p.LoadOrGenNodeKey(tmCfg.NodeKeyFile())
	if err != nil {
		return err
	}

	app := cfg.AppConstructor(*val)

	genDocProvider := node.DefaultGenesisDocProviderFunc(tmCfg)
	tmNode, err := node.NewNode(
		tmCfg,
		pvm.LoadOrGenFilePV(tmCfg.PrivValidatorKeyFile(), tmCfg.PrivValidatorStateFile()),
		nodeKey,
		proxy.NewLocalClientCreator(app),
		genDocProvider,
		node.DefaultDBProvider,
		node.DefaultMetricsProvider(tmCfg.Instrumentation),
		logger.With("module", val.Moniker),
	)
	if err != nil {
		return err
	}

	if err := tmNode.Start(); err != nil {
		return err
	}

	val.tmNode = tmNode

	if val.RPCAddress != "" {
		val.RPCClient = local.New(tmNode)
	}

	// We'll need a RPC client if the validator exposes a gRPC or REST endpoint.
	if val.APIAddress != "" || val.AppConfig.GRPC.Enable {
		val.ClientCtx = val.ClientCtx.
			WithClient(val.RPCClient)

		// Add the tx service in the gRPC router.
		app.RegisterTxService(val.ClientCtx)

		// Add the tendermint queries service in the gRPC router.
		app.RegisterTendermintService(val.ClientCtx)
	}

	if val.AppConfig.API.Enable && val.APIAddress != "" {
		apiSrv := api.New(val.ClientCtx, logger.With("module", "api-server"))
		app.RegisterAPIRoutes(apiSrv, val.AppConfig.API)

		errCh := make(chan error)

		go func() {
			if err := apiSrv.Start(val.AppConfig.Config); err != nil {
				errCh <- err
			}
		}()

		select {
		case err := <-errCh:
			return err
		case <-time.After(srvtypes.ServerStartTime): // assume server started successfully
		}

		val.api = apiSrv
	}

	if val.AppConfig.GRPC.Enable {
		grpcSrv, err := servergrpc.StartGRPCServer(val.ClientCtx, app, val.AppConfig.GRPC)
		if err != nil {
			return err
		}

		val.grpc = grpcSrv

		if val.AppConfig.GRPCWeb.Enable {
			val.grpcWeb, err = servergrpc.StartGRPCWeb(grpcSrv, val.AppConfig.Config)
			if err != nil {
				return err
			}
		}
	}

	if val.AppConfig.JSONRPC.Enable && val.AppConfig.JSONRPC.Address != "" {
		if val.Ctx == nil || val.Ctx.Viper == nil {
			return fmt.Errorf("validator %s context is nil", val.Moniker)
		}

		tmEndpoint := "/websocket"
		tmRPCAddr := fmt.Sprintf("tcp://%s", val.AppConfig.GRPC.Address)

		val.jsonrpc, val.jsonrpcDone, err = server.StartJSONRPC(val.Ctx, val.ClientCtx, tmRPCAddr, tmEndpoint, val.AppConfig, nil)
		if err != nil {
			return err
		}

		address := fmt.Sprintf("http://%s", val.AppConfig.JSONRPC.Address)

		val.JSONRPCClient, err = ethclient.Dial(address)
		if err != nil {
			return fmt.Errorf("failed to dial JSON-RPC at %s: %w", val.AppConfig.JSONRPC.Address, err)
		}
	}

	return nil
}

// collectGenFiles doesn't actually actually collect gentx and fill into genesis since we use dogfood to replace staking module. Modify genesisfile to fill validator related information in genesisfile
func collectGenFiles(cfg Config, vals []*Validator, outputDir string) error {
	genTime := tmtime.Now()

	for i := 0; i < cfg.NumValidators; i++ {
		tmCfg := vals[i].Ctx.Config

		nodeDir := filepath.Join(outputDir, vals[i].Moniker, "exocored")
		gentxsDir := filepath.Join(outputDir, "gentxs")

		tmCfg.Moniker = vals[i].Moniker
		tmCfg.SetRoot(nodeDir)

		initCfg := genutiltypes.NewInitConfig(cfg.ChainID, gentxsDir, vals[i].NodeID, vals[i].PubKey)

		genFile := tmCfg.GenesisFile()
		genDoc, err := types.GenesisDocFromFile(genFile)
		if err != nil {
			return err
		}

		appState, err := genutil.GenAppStateFromConfig(cfg.Codec, cfg.TxConfig,
			tmCfg, initCfg, *genDoc, banktypes.GenesisBalancesIterator{}, genutiltypes.DefaultMessageValidator)
		if err != nil {
			return err
		}

		// overwrite each validator's genesis file to have a canonical genesis time
		if err := genutil.ExportGenesisFileWithTime(genFile, cfg.ChainID, nil, appState, genTime); err != nil {
			return err
		}
	}

	return nil
}

func initGenFiles(cfg Config, genAccounts []authtypes.GenesisAccount, genBalances []banktypes.Balance, genFiles []string) error {
	// set the accounts in the genesis state
	var authGenState authtypes.GenesisState
	cfg.Codec.MustUnmarshalJSON(cfg.GenesisState[authtypes.ModuleName], &authGenState)

	accounts, err := authtypes.PackAccounts(genAccounts)
	if err != nil {
		return err
	}

	authGenState.Accounts = append(authGenState.Accounts, accounts...)
	cfg.GenesisState[authtypes.ModuleName] = cfg.Codec.MustMarshalJSON(&authGenState)

	// set the balances in the genesis state
	var bankGenState banktypes.GenesisState
	bankGenState.Balances = genBalances
	cfg.GenesisState[banktypes.ModuleName] = cfg.Codec.MustMarshalJSON(&bankGenState)

	var govGenState govv1.GenesisState
	cfg.Codec.MustUnmarshalJSON(cfg.GenesisState[govtypes.ModuleName], &govGenState)

	govGenState.Params.MinDeposit[0].Denom = cfg.BondDenom
	cfg.GenesisState[govtypes.ModuleName] = cfg.Codec.MustMarshalJSON(&govGenState)

	var crisisGenState crisistypes.GenesisState
	cfg.Codec.MustUnmarshalJSON(cfg.GenesisState[crisistypes.ModuleName], &crisisGenState)

	crisisGenState.ConstantFee.Denom = cfg.BondDenom
	cfg.GenesisState[crisistypes.ModuleName] = cfg.Codec.MustMarshalJSON(&crisisGenState)

	var evmGenState evmtypes.GenesisState
	cfg.Codec.MustUnmarshalJSON(cfg.GenesisState[evmtypes.ModuleName], &evmGenState)

	evmGenState.Params.EvmDenom = cfg.BondDenom
	cfg.GenesisState[evmtypes.ModuleName] = cfg.Codec.MustMarshalJSON(&evmGenState)

	// set validators related modules: assets, operator, dogfood
	var assetsGenState assetstypes.GenesisState
	cfg.Codec.MustUnmarshalJSON(cfg.GenesisState[assetstypes.ModuleName], &assetsGenState)
	NewGenStateAssets(&assetsGenState)
	cfg.GenesisState[assetstypes.ModuleName] = cfg.Codec.MustMarshalJSON(&assetsGenState)

	var operatorGenState operatortypes.GenesisState
	cfg.Codec.MustUnmarshalJSON(cfg.GenesisState[assetstypes.ModuleName], &operatorGenState)

	var dogfoodGenState dogfoodtypes.GenesisState
	cfg.Codec.MustUnmarshalJSON(cfg.GenesisState[assetstypes.ModuleName], &dogfoodGenState)

	// set oracle genesis statse
	var oracleGenState oracletypes.GenesisState
	cfg.Codec.MustUnmarshalJSON(cfg.GenesisState[assetstypes.ModuleName], &oracleGenState)

	appGenStateJSON, err := json.MarshalIndent(cfg.GenesisState, "", "  ")
	if err != nil {
		return err
	}

	genDoc := types.GenesisDoc{
		ChainID:    cfg.ChainID,
		AppState:   appGenStateJSON,
		Validators: nil,
	}

	// generate empty genesis files for each validator and save
	for i := 0; i < cfg.NumValidators; i++ {
		if err := genDoc.SaveAs(genFiles[i]); err != nil {
			return err
		}
	}

	return nil
}

func WriteFile(name string, dir string, contents []byte) error {
	file := filepath.Join(dir, name)

	err := tmos.EnsureDir(dir, 0o755)
	if err != nil {
		return err
	}

	return tmos.WriteFile(file, contents, 0o644)
}

// set deposits and operator_assets for assets genesisState
func NewGenStateAssets(operatorAccAddresses []sdktypes.AccAddress, depositAmount, stakingAmount sdkmath.Int) (assetstypes.GenesisState, error) {
	if stakingAmount.GT(depositAmount) {
		return DefaultGenStateAssets, fmt.Errorf("stakingAmount %v should be less than depositAmount %v", stakingAmount, depositAmount)
	}
	n := len(operatorAccAddresses)
	nInt := sdkmath.NewInt(int64(n))
	totalDepositAmount := depositAmount.Mul(nInt)
	depositsByStakers := make([]assetstypes.DepositsByStaker, 0, len(DefaultGenStateAssets.Tokens)*n)
	operatorsAssets := make([]assetstypes.AssetsByOperator, 0, n)
	nAssets := len(DefaultGenStateAssets.Tokens)
	for i := 0; i < nAssets; i++ {
		DefaultGenStateAssets.Tokens[i].StakingTotalAmount = totalDepositAmount
	}
	for _, operatorAccAddress := range operatorAccAddresses {
		// use the same address []byte for operator(exo..) and staker(0x...), both derived from the same pubkey and since evmos use ethsecp256k1, this address is generated from keccak-256(.) instead of ripemd160(sha256(.))
		stakerAddrStr := hexutil.Encode(operatorAccAddress)
		depositsByAssets := make([]assetstypes.DepositByAsset, 0, nAssets)
		assetsStates := make([]assetstypes.AssetByID, 0, nAssets)
		stakerID := ""
		assetID := ""
		for _, asset := range DefaultGenStateAssets.Tokens {
			stakerID, assetID = assetstypes.GetStakerIDAndAssetIDFromStr(asset.AssetBasicInfo.LayerZeroChainID, stakerAddrStr, asset.AssetBasicInfo.Address)
			depositsByAssets = append(depositsByAssets, assetstypes.DepositByAsset{
				AssetID: assetID,
				Info: assetstypes.StakerAssetInfo{
					TotalDepositAmount:        depositAmount,
					WithdrawableAmount:        depositAmount.Sub(stakingAmount),
					PendingUndelegationAmount: sdkmath.ZeroInt(),
				},
			})
			assetsStates = append(assetsStates, assetstypes.AssetByID{
				AssetID: assetID,
				Info: assetstypes.OperatorAssetInfo{
					TotalAmount:               stakingAmount,
					PendingUndelegationAmount: sdkmath.ZeroInt(),
					TotalShare:                sdkmath.LegacyNewDecFromInt(stakingAmount),
					// only take self delegation for genesis state
					OperatorShare: sdkmath.LegacyNewDecFromInt(stakingAmount),
				},
			})
		}
		depositsByStakers = append(depositsByStakers, assetstypes.DepositsByStaker{
			StakerID: stakerID,
			Deposits: depositsByAssets,
		})
		operatorsAssets = append(operatorsAssets, assetstypes.AssetsByOperator{
			Operator:    operatorAccAddress.String(),
			AssetsState: assetsStates,
		})
	}

	DefaultGenStateAssets.Deposits = depositsByStakers
	DefaultGenStateAssets.OperatorAssets = operatorsAssets

	return DefaultGenStateAssets, nil
}

// stakingAmount, each operator opt in evry AVS with stakingAmount of every assets
// each avs suppport all assets
// each operator opts in every avs
// each operator deposited and self staked all assets with: (depsitAmount, stakingAmount)
// intial price for every asset is 1 USD
func NewGenStateOperator(operatorAccAddresses []sdktypes.AccAddress, consPubKeys []string, commissionRate sdkmath.LegacyDec, chainID string, optedAVSAddresses []string, stakingAmount sdkmath.Int, genStateAssets assetstypes.GenesisState) (operatortypes.GenesisState, error) {
	// total stakingAmount one operator holds among all assets
	stakingAmount = stakingAmount.Mul(sdkmath.NewInt(int64(len(genStateAssets.Tokens))))
	if len(operatorAccAddresses) != len(consPubKeys) {
		return DefaultGenStateOperator, fmt.Errorf("length of operatorAccAddresses %d should be equal to length of consPubKeys %d", len(operatorAccAddresses), len(consPubKeys))
	}
	n := len(operatorAccAddresses)
	totalStakingAmount := stakingAmount.Mul(sdkmath.NewInt(int64(n)))
	for i, operatorAccAddress := range operatorAccAddresses {
		// operators
		DefaultGenStateOperator.Operators = append(DefaultGenStateOperator.Operators, operatortypes.OperatorDetail{
			OperatorAddress: operatorAccAddress.String(),
			OperatorInfo: operatortypes.OperatorInfo{
				EarningsAddr:     operatorAccAddress.String(),
				OperatorMetaInfo: fmt.Sprintf("operator_%d", i),
				Commission: stakingtypes.Commission{
					CommissionRates: stakingtypes.CommissionRates{
						Rate:          commissionRate,
						MaxRate:       commissionRate.Mul(sdkmath.LegacyNewDec(2)),
						MaxChangeRate: sdkmath.LegacyNewDecWithPrec(1, 1),
					},
				},
			},
		})
		// operator_records
		DefaultGenStateOperator.OperatorRecords = append(DefaultGenStateOperator.OperatorRecords, operatortypes.OperatorConsKeyRecord{
			OperatorAddress: operatorAccAddress.String(),
			Chains: []operatortypes.ChainDetails{
				{
					ChainID:      chainID,
					ConsensusKey: consPubKeys[i],
				},
			},
		})
		// OptStates
		for _, AVSAddress := range optedAVSAddresses {
			DefaultGenStateOperator.OptStates = append(DefaultGenStateOperator.OptStates, operatortypes.OptedState{
				Key: operatorAccAddress.String() + "/" + AVSAddress,
				OptInfo: operatortypes.OptedInfo{
					OptedInHeight:  1,
					OptedOutHeight: 18446744073709551615,
					Jailed:         false,
				},
			})
			// OperatorUSDValues
			// the price unit of assets is 1 not decimal 18
			stakingValue := sdktypes.TokensToConsensusPower(stakingAmount, evmostypes.PowerReduction)
			DefaultGenStateOperator.OperatorUSDValues = append(DefaultGenStateOperator.OperatorUSDValues, operatortypes.OperatorUSDValue{
				Key: AVSAddress + "/" + operatorAccAddress.String(),
				OptedUSDValue: operatortypes.OperatorOptedUSDValue{
					SelfUSDValue:   sdkmath.LegacyNewDec(stakingValue),
					TotalUSDValue:  sdkmath.LegacyNewDec(stakingValue),
					ActiveUSDValue: sdkmath.LegacyNewDec(stakingValue),
				},
			})
		}
	}
	// AVSUSDValues
	for _, AVSAddress := range optedAVSAddresses {
		DefaultGenStateOperator.AVSUSDValues = append(DefaultGenStateOperator.AVSUSDValues, operatortypes.AVSUSDValue{
			AVSAddr: AVSAddress,
			Value: operatortypes.DecValueField{
				// the price unit of assets is 1 not decimal 18
				Amount: sdkmath.LegacyNewDec(sdktypes.TokensToConsensusPower(totalStakingAmount, evmostypes.PowerReduction)),
			},
		})
	}
	return DefaultGenStateOperator, nil
}

func NewGenStateDogfood(consPubKeys []string, powers []sdkmath.LegacyDec, stakingAmount sdkmath.Int, genStateAssets assetstypes.GenesisState) (dogfoodtypes.GenesisState, error) {
	power := sdktypes.TokensToConsensusPower(stakingAmount.Mul(sdkmath.NewInt(int64(len(genStateAssets.Tokens)))), evmostypes.PowerReduction)
	// if len(consPubKeys) != len(powers) {
	// 	return DefaultGenStateDogfood, fmt.Errorf("lenght of consensusKey %d should be equal to length of powers %d", len(consPubKeys), len(powers))
	// }
	DefaultGenStateDogfood.Params.EpochIdentifier = "minute"
	DefaultGenStateDogfood.Params.EpochsUntilUnbonded = 5
	DefaultGenStateDogfood.Params.MinSelfDelegation = sdkmath.NewInt(100)
	assetIDs := make(map[string]bool)
	for _, assetID := range DefaultGenStateDogfood.Params.AssetIDs {
		assetIDs[assetID] = true
	}
	for _, asset := range genStateAssets.Tokens {
		_, assetID := assetstypes.GetStakerIDAndAssetIDFromStr(asset.AssetBasicInfo.LayerZeroChainID, "", asset.AssetBasicInfo.Address)
		if assetIDs[assetID] {
			continue
		}
		DefaultGenStateDogfood.Params.AssetIDs = append(DefaultGenStateDogfood.Params.AssetIDs, assetID)
	}
	for _, consPubKey := range consPubKeys {
		DefaultGenStateDogfood.ValSet = append(DefaultGenStateDogfood.ValSet, dogfoodtypes.GenesisValidator{
			PublicKey: consPubKey,
			Power:     power,
		})
	}
	DefaultGenStateDogfood.LastTotalPower = sdkmath.NewInt(power * int64(len(consPubKeys)))
	return DefaultGenStateDogfood, nil
}

func NewGenStateDelegation(operatorAccAddresses []sdk.AccAddress, genStateAssets assetstypes.GenesisState) (delegationtypes.GenesisState, error) {
	return DefaultGenStateDelegation, nil
}

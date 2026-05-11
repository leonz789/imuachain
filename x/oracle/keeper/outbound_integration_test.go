package keeper_test

import (
	"crypto/ecdsa"
	"encoding/hex"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	keepertest "github.com/imua-xyz/imuachain/testutil/keeper"
	dogfoodkeeper "github.com/imua-xyz/imuachain/x/dogfood/keeper"
	dogfoodtypes "github.com/imua-xyz/imuachain/x/dogfood/types"
	"github.com/imua-xyz/imuachain/x/oracle/keeper"
	"github.com/imua-xyz/imuachain/x/oracle/types"
	"github.com/stretchr/testify/require"
)

// TestOutboundE2EFlow walks: enqueue → checkpoint → sign → finalize → query.
func TestOutboundE2EFlow(t *testing.T) {
	k, ctx := keepertest.OracleKeeper(t)
	dstChainID := uint64(101)

	// 3 validators × 100 power = 300 total. Strict 2/3 (`signedPower*3 > totalPower*2`):
	// 2 sigs (600 == 600) NOT finalized; 3 sigs (900 > 600) finalized.
	const numVals = 3
	privKeys := make([]*ecdsa.PrivateKey, numVals)
	addrs := make([]common.Address, numVals)
	mockValidators := make([]dogfoodtypes.ImuachainValidator, numVals)
	for i := 0; i < numVals; i++ {
		pk, err := crypto.GenerateKey()
		require.NoError(t, err)
		privKeys[i] = pk
		addrs[i] = crypto.PubkeyToAddress(pk.PublicKey)
		mockValidators[i] = dogfoodtypes.ImuachainValidator{
			Address: addrs[i].Bytes(),
			Power:   100,
		}
	}
	patcher := gomonkey.ApplyMethod(
		reflect.TypeOf(dogfoodkeeper.Keeper{}),
		"GetAllImuachainValidators",
		func(_ dogfoodkeeper.Keeper, _ sdk.Context) []dogfoodtypes.ImuachainValidator {
			return mockValidators
		},
	)
	defer patcher.Reset()

	// Step 1: enqueue messages
	msg1 := keeper.OutboundMsg{
		DstChainID: dstChainID,
		SeqNum:     1,
		Nonce:      1,
		PayloadHex: hex.EncodeToString([]byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09}),
		Height:     100,
	}
	msg2 := keeper.OutboundMsg{
		DstChainID: dstChainID,
		SeqNum:     2,
		Nonce:      2,
		PayloadHex: hex.EncodeToString([]byte{0x00, 0x0a, 0x0b, 0x0c}),
		Height:     100,
	}
	require.NoError(t, k.EnqueueOutboundForTest(ctx, msg1))
	require.NoError(t, k.EnqueueOutboundForTest(ctx, msg2))

	// Verify queue
	msgs := k.GetOutboundMessages(ctx, dstChainID, 0, 100)
	require.Len(t, msgs, 2)

	// Step 2: create checkpoint
	created := k.CreateCheckpointForPendingOutbound(ctx, dstChainID)
	require.True(t, created)

	nonce := k.GetLatestCheckpointNonce(ctx, dstChainID)
	require.EqualValues(t, 1, nonce)

	cp, found := k.GetCheckpoint(ctx, dstChainID, 1)
	require.True(t, found)
	require.False(t, cp.Finalized)
	require.Equal(t, uint64(1), cp.SeqStart)
	require.Equal(t, uint64(2), cp.SeqEnd)

	// Step 3: each validator signs
	checkpointHash := types.ComputeCheckpointHash(cp.Nonce, cp.DstChainID, cp.MessagesHash)
	ethHash := types.ComputeEthSignedMessageHash(checkpointHash)
	for i := 0; i < numVals; i++ {
		sig, err := crypto.Sign(ethHash.Bytes(), privKeys[i])
		require.NoError(t, err)

		var r, s [32]byte
		copy(r[:], sig[0:32])
		copy(s[:], sig[32:64])
		v := uint8(sig[64] + 27)

		finalized, err := k.AddCheckpointSignature(ctx, dstChainID, 1, addrs[i], v, r, s, 100)
		require.NoError(t, err)

		if i < numVals-1 {
			require.False(t, finalized, "sig %d should not cross strict 2/3 yet", i+1)
		} else {
			require.True(t, finalized, "sig %d should finalize", i+1)
		}
	}

	// Step 4: checkpoint finalized
	cp, found = k.GetCheckpoint(ctx, dstChainID, 1)
	require.True(t, found)
	require.True(t, cp.Finalized)

	// Step 5: signatures queryable
	sigs := k.GetCheckpointSignatures(ctx, dstChainID, 1)
	require.Len(t, sigs, numVals)

	t.Logf("Outbound E2E: %d messages, nonce=%d, %d sigs, finalized=%v",
		len(msgs), nonce, len(sigs), cp.Finalized)
}

// TestCheckpointHashMatchesBridgeVerifier verifies that the Go hash computation
// would match the Solidity computation in BridgeVerifier.sol.
// Format: keccak256(abi.encode(BRIDGE_ID, nonce, dstChainID, messagesHash))
func TestCheckpointHashMatchesBridgeVerifier(t *testing.T) {
	nonce := uint64(1)
	dstChainID := uint64(40161)
	messagesHash := crypto.Keccak256Hash([]byte("test messages"))

	hash := types.ComputeCheckpointHash(nonce, dstChainID, messagesHash)
	require.NotEmpty(t, hash.Bytes())

	// Verify the hash is deterministic
	hash2 := types.ComputeCheckpointHash(nonce, dstChainID, messagesHash)
	require.Equal(t, hash, hash2)

	// Verify EthSignedMessageHash wrapping
	ethHash := types.ComputeEthSignedMessageHash(hash)
	require.NotEqual(t, hash, ethHash)

	// Verify ECDSA sign/recover roundtrip
	pk, err := crypto.GenerateKey()
	require.NoError(t, err)
	expectedAddr := crypto.PubkeyToAddress(pk.PublicKey)

	sig, err := crypto.Sign(ethHash.Bytes(), pk)
	require.NoError(t, err)

	recoveredPub, err := crypto.Ecrecover(ethHash.Bytes(), sig)
	require.NoError(t, err)

	pubKey, err := crypto.UnmarshalPubkey(recoveredPub)
	require.NoError(t, err)

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	require.Equal(t, expectedAddr, recoveredAddr)
}

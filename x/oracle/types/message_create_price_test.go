package types

import (
	"testing"

	"github.com/ExocoreNetwork/exocore/testutil/sample"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	// sdkerrors "cosmossdk.io/errors""
	"github.com/stretchr/testify/require"
)

func TestMsgPriceFeed_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgPriceFeed
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgPriceFeed{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgPriceFeed{
				Creator: sample.AccAddress(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}

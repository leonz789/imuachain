package testutil

import "testing"

//go:generate mockgen -destination expected_bank_mock.go -package testutil github.com/ExocoreNetwork/exocore/x/feedistribution/types BankKeeper
//go:generate mockgen -destination expected_epochs_mock.go -package testutil github.com/ExocoreNetwork/exocore/x/feedistribution/types EpochsKeeper
//go:generate mockgen -destination expected_account_mock.go -package testutil github.com/ExocoreNetwork/exocore/x/feedistribution/types AccountKeeper
//go:generate mockgen -destination expected_mint_mock.go -package testutil github.com/ExocoreNetwork/exocore/x/feedistribution/types MintKeeper
//go:generate mockgen -destination expected_avs_mock.go -package testutil github.com/ExocoreNetwork/exocore/x/feedistribution/types AVSKeeper
func TestMock(t *testing.T) {
	// t.Fatal("not implemented")
}

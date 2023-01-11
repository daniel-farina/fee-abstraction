package feeabs_test

import (
	"testing"

	// "github.com/CosmWasm/wasmd/x/wasm/keeper/wasmtesting"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/keeper/wasmtesting"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	wasmibctesting "github.com/notional-labs/feeabstraction/v1/x/feeabs/ibctesting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromIBCTransferToContract(t *testing.T) {
	// scenario: given two chains,
	//           with a contract on chain B
	//           then the contract can handle the receiving side of an ics20 transfer
	//           that was started on chain A via ibc transfer module

	transferAmount := sdk.NewInt(1)
	specs := map[string]struct {
		contract             wasmtesting.IBCContractCallbacks
		setupContract        func(t *testing.T, contract wasmtesting.IBCContractCallbacks, chain *wasmibctesting.TestChain)
		expChainABalanceDiff sdk.Int
		expChainBBalanceDiff sdk.Int
	}{
		"ack": {
			contract: &ackReceiverContract{},
			setupContract: func(t *testing.T, contract wasmtesting.IBCContractCallbacks, chain *wasmibctesting.TestChain) {
				c := contract.(*ackReceiverContract)
				c.t = t
				c.chain = chain
			},
			expChainABalanceDiff: transferAmount.Neg(),
			expChainBBalanceDiff: transferAmount,
		},
		// "nack": {
		// 	contract: &nackReceiverContract{},
		// 	setupContract: func(t *testing.T, contract wasmtesting.IBCContractCallbacks, chain *wasmibctesting.TestChain) {
		// 		c := contract.(*nackReceiverContract)
		// 		c.t = t
		// 	},
		// 	expChainABalanceDiff: sdk.ZeroInt(),
		// 	expChainBBalanceDiff: sdk.ZeroInt(),
		// },
		// "error": {
		// 	contract: &errorReceiverContract{},
		// 	setupContract: func(t *testing.T, contract wasmtesting.IBCContractCallbacks, chain *wasmibctesting.TestChain) {
		// 		c := contract.(*errorReceiverContract)
		// 		c.t = t
		// 	},
		// 	expChainABalanceDiff: sdk.ZeroInt(),
		// 	expChainBBalanceDiff: sdk.ZeroInt(),
		// },
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			var (
				chainAOpts = []wasmkeeper.Option{wasmkeeper.WithWasmEngine(
					wasmtesting.NewIBCContractMockWasmer(spec.contract),
				)}
				coordinator = wasmibctesting.NewCoordinator(t, 2, []wasmkeeper.Option{}, chainAOpts)
				chainA      = coordinator.GetChain(wasmibctesting.GetChainID(0))
				chainB      = coordinator.GetChain(wasmibctesting.GetChainID(1))
			)
			coordinator.CommitBlock(chainA, chainB)
			myContractAddr := chainB.SeedNewContractInstance()
			contractBPortID := chainB.ContractInfo(myContractAddr).IBCPortID

			spec.setupContract(t, spec.contract, chainB)

			path := wasmibctesting.NewPath(chainA, chainB)
			path.EndpointA.ChannelConfig = &ibctesting.ChannelConfig{
				PortID:  "transfer",
				Version: ibctransfertypes.Version,
				Order:   channeltypes.UNORDERED,
			}
			path.EndpointB.ChannelConfig = &ibctesting.ChannelConfig{
				PortID:  contractBPortID,
				Version: ibctransfertypes.Version,
				Order:   channeltypes.UNORDERED,
			}

			coordinator.SetupConnections(path)
			coordinator.CreateChannels(path)

			originalChainABalance := chainA.Balance(chainA.SenderAccount.GetAddress(), sdk.DefaultBondDenom)
			// when transfer via sdk transfer from A (module) -> B (contract)
			coinToSendToB := sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
			timeoutHeight := clienttypes.NewHeight(1, 110)
			msg := ibctransfertypes.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, coinToSendToB, chainA.SenderAccount.GetAddress().String(), chainB.SenderAccount.GetAddress().String(), timeoutHeight, 0)
			_, err := chainA.SendMsgs(msg)
			require.NoError(t, err)
			require.NoError(t, path.EndpointB.UpdateClient())

			// then
			require.Equal(t, 1, len(chainA.PendingSendPackets))
			require.Equal(t, 0, len(chainB.PendingSendPackets))

			// and when relay to chain B and handle Ack on chain A
			err = coordinator.RelayAndAckPendingPackets(path)
			require.NoError(t, err)

			// then
			require.Equal(t, 0, len(chainA.PendingSendPackets))
			require.Equal(t, 0, len(chainB.PendingSendPackets))

			// and source chain balance was decreased
			newChainABalance := chainA.Balance(chainA.SenderAccount.GetAddress(), sdk.DefaultBondDenom)
			assert.Equal(t, originalChainABalance.Amount.Add(spec.expChainABalanceDiff), newChainABalance.Amount)

			// and dest chain balance contains voucher
			expBalance := ibctransfertypes.GetTransferCoin(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, coinToSendToB.Denom, spec.expChainBBalanceDiff)
			gotBalance := chainB.Balance(chainB.SenderAccount.GetAddress(), expBalance.Denom)
			assert.Equal(t, expBalance, gotBalance, "got total balance: %s", chainB.AllBalances(chainB.SenderAccount.GetAddress()))
		})
	}
}

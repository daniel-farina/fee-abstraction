package types

import (
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgSendQuerySpotPrice{}

// Route Implements Msg.
func (m MsgSendQuerySpotPrice) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgSendQuerySpotPrice) Type() string { return sdk.MsgTypeURL(&m) }

// GetSigners returns the expected signers for a MsgMintAndAllocateExp .
func (m MsgSendQuerySpotPrice) GetSigners() []sdk.AccAddress {
	daoAccount, err := sdk.AccAddressFromBech32(m.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{daoAccount}
}

// GetSignBytes Implements Msg.
func (m MsgSendQuerySpotPrice) GetSignBytes() []byte {
	return sdk.MustSortJSON(legacy.Cdc.MustMarshalJSON(&m))
}

// ValidateBasic does a sanity check on the provided data.
func (m MsgSendQuerySpotPrice) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.FromAddress)
	if err != nil {
		return sdkerrors.Wrap(err, "from address must be valid address")
	}
	return nil
}

func NewMsgSendQuerySpotPrice(
	fromAddr sdk.AccAddress,
	poolId uint64,
	sourcePort string,
	sourceChannel string,
	baseAsset string,
	quoteAsset string,
) *MsgSendQuerySpotPrice {
	return &MsgSendQuerySpotPrice{
		FromAddress:   fromAddr.String(),
		PoolId:        poolId,
		SourcePort:    sourcePort,
		SourceChannel: sourceChannel,
		BaseDenom:     baseAsset,
		QuoteDenom:    quoteAsset,
	}
}

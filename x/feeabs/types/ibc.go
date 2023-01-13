package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// IBCPortID is the default port id that profiles module binds to.
	IBCPortID = "feeabs"
)

type SpotPrice struct {
	SpotPrice string `json:"spot_price"`
}

var ModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

// IBCPortKey defines the key to store the port ID in store.
var IBCPortKey = []byte{0x01}

// NewOsmosisQueryRequestPacketData create new packet for ibc.
func NewOsmosisQueryRequestPacketData(poolId uint64, baseDenom string, quoteDenom string) OsmosisQuerySpotPriceRequestPacketData {
	return OsmosisQuerySpotPriceRequestPacketData{
		PoolId:          poolId,
		BaseAssetDenom:  baseDenom,
		QuoteAssetDenom: quoteDenom,
	}
}

// GetBytes is a helper for serializing.
func (p OsmosisQuerySpotPriceRequestPacketData) GetBytes() []byte {
	var ibcPacket FeeabsIbcPacketData
	ibcPacket.Packet = &FeeabsIbcPacketData_IbcOsmosisQuerySpotPriceRequestPacketData{&p}

	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&ibcPacket))
}

// NewSwapAmountInRoutePacketData create new packet for swap token over ibc.
func NewSwapAmountInRoutePacketData(poolId uint64, tokenOutDenom string) SwapAmountInRoute {
	return SwapAmountInRoute{
		PoolId:        poolId,
		TokenOutDenom: tokenOutDenom,
	}
}

// GetBytes is a helper for serializing.
func (p SwapAmountInRoute) GetBytes() []byte {
	var ibcPacket FeeabsIbcPacketData
	ibcPacket.Packet = &FeeabsIbcPacketData_IbcSwapAmountInRoute{&p}

	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&ibcPacket))
}

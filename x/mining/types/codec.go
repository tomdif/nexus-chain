package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgPostJob{}, "nexus/mining/MsgPostJob", nil)
	cdc.RegisterConcrete(&MsgSubmitProof{}, "nexus/mining/MsgSubmitProof", nil)
	cdc.RegisterConcrete(&MsgClaimRewards{}, "nexus/mining/MsgClaimRewards", nil)
	cdc.RegisterConcrete(&MsgCancelJob{}, "nexus/mining/MsgCancelJob", nil)
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	// Skip SDK Msg registration for now - requires proper protobuf generation
}

func RegisterMsgServer(registry codectypes.InterfaceRegistry, srv MsgServer) {
	// Placeholder for gRPC registration
}

func RegisterQueryServer(registry codectypes.InterfaceRegistry, srv QueryServer) {
	// Placeholder for gRPC registration
}

package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgPostJob{}, "nexus/MsgPostJob")
	legacy.RegisterAminoMsg(cdc, &MsgSubmitProof{}, "nexus/MsgSubmitProof")
	legacy.RegisterAminoMsg(cdc, &MsgClaimRewards{}, "nexus/MsgClaimRewards")
	legacy.RegisterAminoMsg(cdc, &MsgCancelJob{}, "nexus/MsgCancelJob")
	legacy.RegisterAminoMsg(cdc, &MsgSubmitPublicJob{}, "nexus/MsgSubmitPublicJob")
	legacy.RegisterAminoMsg(cdc, &MsgSubmitWork{}, "nexus/MsgSubmitWork")
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgPostJob{},
		&MsgSubmitProof{},
		&MsgClaimRewards{},
		&MsgCancelJob{},
		&MsgSubmitPublicJob{},
		&MsgSubmitWork{},
	)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterCodec(Amino)
}
package types

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
)

// GenesisState defines the mining module's genesis state
type GenesisState struct {
	Params              Params       `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	Jobs                []Job        `protobuf:"bytes,2,rep,name=jobs,proto3" json:"jobs"`
	Checkpoints         []Checkpoint `protobuf:"bytes,3,rep,name=checkpoints,proto3" json:"checkpoints"`
	ValidatorRewardPool int64        `protobuf:"varint,4,opt,name=validator_reward_pool,json=validatorRewardPool,proto3" json:"validator_reward_pool"`
	EmissionEscrow      int64        `protobuf:"varint,5,opt,name=emission_escrow,json=emissionEscrow,proto3" json:"emission_escrow"`
	LastCheckpointID    uint64       `protobuf:"varint,6,opt,name=last_checkpoint_id,json=lastCheckpointId,proto3" json:"last_checkpoint_id"`
	CurrentProblemSize  int64        `protobuf:"varint,7,opt,name=current_problem_size,json=currentProblemSize,proto3" json:"current_problem_size"`
	BackgroundJobCount  int64        `protobuf:"varint,8,opt,name=background_job_count,json=backgroundJobCount,proto3" json:"background_job_count"`
}

func (gs *GenesisState) Reset()         { *gs = GenesisState{} }
func (gs *GenesisState) String() string { return "" }
func (gs *GenesisState) ProtoMessage()  {}

// DefaultGenesis returns default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:              DefaultParams(),
		Jobs:                []Job{},
		Checkpoints:         []Checkpoint{},
		ValidatorRewardPool: 0,
		EmissionEscrow:      0,
		LastCheckpointID:    0,
		CurrentProblemSize:  64,
		BackgroundJobCount:  0,
	}
}

// Validate performs genesis state validation
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	if gs.CurrentProblemSize < 64 || gs.CurrentProblemSize > 2048 {
		return ErrInvalidParams
	}
	return nil
}

// GetGenesisStateFromAppState returns the mining GenesisState from raw app state
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState
	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}
	return &genesisState
}

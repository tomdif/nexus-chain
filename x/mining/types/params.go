package types

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultMinerSharePercent     = 80
	DefaultValidatorSharePercent = 20
	DefaultCheckpointInterval    = 300
	DefaultMinProofPeriod        = 7 * 24 * time.Hour
	DefaultJobFeeBurnPercent     = 2
	DefaultTxFeeBurnPercent      = 50
)

var (
	DefaultBackgroundEmissionRate = math.NewInt(1000000)
	DefaultMinJobReward           = sdk.NewCoins(sdk.NewCoin("unexus", math.NewInt(1000000)))
	DefaultMaxJobDuration         = 24 * time.Hour
)

type Params struct {
	MinerSharePercent      uint64        `protobuf:"varint,1,opt,name=miner_share_percent,proto3" json:"miner_share_percent"`
	ValidatorSharePercent  uint64        `protobuf:"varint,2,opt,name=validator_share_percent,proto3" json:"validator_share_percent"`
	CheckpointInterval     int64         `protobuf:"varint,3,opt,name=checkpoint_interval,proto3" json:"checkpoint_interval"`
	MinProofPeriod         time.Duration `protobuf:"varint,4,opt,name=min_proof_period,proto3,casttype=time.Duration" json:"min_proof_period"`
	JobFeeBurnPercent      uint64        `protobuf:"varint,5,opt,name=job_fee_burn_percent,proto3" json:"job_fee_burn_percent"`
	TxFeeBurnPercent       uint64        `protobuf:"varint,6,opt,name=tx_fee_burn_percent,proto3" json:"tx_fee_burn_percent"`
	BackgroundEmissionRate math.Int      `protobuf:"bytes,7,opt,name=background_emission_rate,proto3,customtype=cosmossdk.io/math.Int" json:"background_emission_rate"`
	MinJobReward           sdk.Coins     `protobuf:"bytes,8,rep,name=min_job_reward,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"min_job_reward"`
	MaxJobDuration         time.Duration `protobuf:"varint,9,opt,name=max_job_duration,proto3,casttype=time.Duration" json:"max_job_duration"`
}

func (p *Params) Reset()         { *p = Params{} }
func (p *Params) String() string { return "" }
func (p *Params) ProtoMessage()  {}

func DefaultParams() Params {
	return Params{
		MinerSharePercent:      DefaultMinerSharePercent,
		ValidatorSharePercent:  DefaultValidatorSharePercent,
		CheckpointInterval:     DefaultCheckpointInterval,
		MinProofPeriod:         DefaultMinProofPeriod,
		JobFeeBurnPercent:      DefaultJobFeeBurnPercent,
		TxFeeBurnPercent:       DefaultTxFeeBurnPercent,
		BackgroundEmissionRate: DefaultBackgroundEmissionRate,
		MinJobReward:           DefaultMinJobReward,
		MaxJobDuration:         DefaultMaxJobDuration,
	}
}

func (p Params) Validate() error {
	if p.MinerSharePercent+p.ValidatorSharePercent != 100 {
		return ErrInvalidParams
	}
	return nil
}

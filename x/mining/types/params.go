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
	MinerSharePercent      uint64        `json:"miner_share_percent"`
	ValidatorSharePercent  uint64        `json:"validator_share_percent"`
	CheckpointInterval     int64         `json:"checkpoint_interval"`
	MinProofPeriod         time.Duration `json:"min_proof_period"`
	JobFeeBurnPercent      uint64        `json:"job_fee_burn_percent"`
	TxFeeBurnPercent       uint64        `json:"tx_fee_burn_percent"`
	BackgroundEmissionRate math.Int      `json:"background_emission_rate"`
	MinJobReward           sdk.Coins     `json:"min_job_reward"`
	MaxJobDuration         time.Duration `json:"max_job_duration"`
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

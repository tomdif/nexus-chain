package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type QueryParamsRequest struct{}

type QueryParamsResponse struct {
	Params Params `json:"params"`
}

type QueryJobRequest struct {
	JobId string `json:"job_id"`
}

type QueryJobResponse struct {
	Job Job `json:"job"`
}

type QueryJobsRequest struct {
	Status uint32 `json:"status"`
}

type QueryJobsResponse struct {
	Jobs []Job `json:"jobs"`
}

type QueryMinerSharesRequest struct {
	Miner string `json:"miner"`
	JobId string `json:"job_id"`
}

type QueryMinerSharesResponse struct {
	Miner  string `json:"miner"`
	JobId  string `json:"job_id"`
	Shares int64  `json:"shares"`
}

type QueryValidatorMiningRecordRequest struct {
	Validator string `json:"validator"`
}

type QueryValidatorMiningRecordResponse struct {
	Record ValidatorMiningRecord `json:"record"`
}

type QueryCheckpointRequest struct {
	Id uint64 `json:"id"`
}

type QueryCheckpointResponse struct {
	Checkpoint Checkpoint `json:"checkpoint"`
}

type QueryLatestCheckpointRequest struct{}

type QueryLatestCheckpointResponse struct {
	Checkpoint Checkpoint `json:"checkpoint"`
}

type MsgPostJobResponse struct {
	JobId string `json:"job_id"`
}

type MsgSubmitProofResponse struct {
	SharesEarned int64 `json:"shares_earned"`
}

type MsgClaimRewardsResponse struct {
	Amount sdk.Coins `json:"amount"`
}

type MsgCancelJobResponse struct {
	Success bool `json:"success"`
}

type MsgServer interface {
	PostJob(context.Context, *MsgPostJob) (*MsgPostJobResponse, error)
	SubmitProof(context.Context, *MsgSubmitProof) (*MsgSubmitProofResponse, error)
	ClaimRewards(context.Context, *MsgClaimRewards) (*MsgClaimRewardsResponse, error)
	CancelJob(context.Context, *MsgCancelJob) (*MsgCancelJobResponse, error)
}

type QueryServer interface {
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	Job(context.Context, *QueryJobRequest) (*QueryJobResponse, error)
	Jobs(context.Context, *QueryJobsRequest) (*QueryJobsResponse, error)
	MinerShares(context.Context, *QueryMinerSharesRequest) (*QueryMinerSharesResponse, error)
	ValidatorMiningRecord(context.Context, *QueryValidatorMiningRecordRequest) (*QueryValidatorMiningRecordResponse, error)
	Checkpoint(context.Context, *QueryCheckpointRequest) (*QueryCheckpointResponse, error)
	LatestCheckpoint(context.Context, *QueryLatestCheckpointRequest) (*QueryLatestCheckpointResponse, error)
}

package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type QueryParamsRequest struct{}

type QueryParamsResponse struct {
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
}

type QueryJobRequest struct {
	JobId string `protobuf:"bytes,1,opt,name=job_id,json=jobId,proto3" json:"job_id"`
}

type QueryJobResponse struct {
	Job Job `protobuf:"bytes,1,opt,name=job,proto3" json:"job"`
}

type QueryJobsRequest struct {
	Status uint32 `protobuf:"varint,1,opt,name=status,proto3" json:"status"`
}

type QueryJobsResponse struct {
	Jobs []Job `protobuf:"bytes,1,rep,name=jobs,proto3" json:"jobs"`
}

type QueryMinerSharesRequest struct {
	Miner string `protobuf:"bytes,1,opt,name=miner,proto3" json:"miner"`
	JobId string `protobuf:"bytes,2,opt,name=job_id,json=jobId,proto3" json:"job_id"`
}

type QueryMinerSharesResponse struct {
	Miner  string `protobuf:"bytes,1,opt,name=miner,proto3" json:"miner"`
	JobId  string `protobuf:"bytes,2,opt,name=job_id,json=jobId,proto3" json:"job_id"`
	Shares int64  `protobuf:"varint,3,opt,name=shares,proto3" json:"shares"`
}

type QueryValidatorMiningRecordRequest struct {
	Validator string `protobuf:"bytes,1,opt,name=validator,proto3" json:"validator"`
}

type QueryValidatorMiningRecordResponse struct {
	Record ValidatorMiningRecord `protobuf:"bytes,1,opt,name=record,proto3" json:"record"`
}

type QueryCheckpointRequest struct {
	Id uint64 `protobuf:"varint,1,opt,name=id,proto3" json:"id"`
}

type QueryCheckpointResponse struct {
	Checkpoint Checkpoint `protobuf:"bytes,1,opt,name=checkpoint,proto3" json:"checkpoint"`
}

type QueryLatestCheckpointRequest struct{}

type QueryLatestCheckpointResponse struct {
	Checkpoint Checkpoint `protobuf:"bytes,1,opt,name=checkpoint,proto3" json:"checkpoint"`
}

// Message responses
type MsgPostJobResponse struct {
	JobId string `protobuf:"bytes,1,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
}

func (m *MsgPostJobResponse) Reset()         { *m = MsgPostJobResponse{} }
func (m *MsgPostJobResponse) String() string { return "MsgPostJobResponse" }
func (m *MsgPostJobResponse) ProtoMessage()  {}

type MsgSubmitProofResponse struct {
	SharesEarned int64 `protobuf:"varint,1,opt,name=shares_earned,json=sharesEarned,proto3" json:"shares_earned,omitempty"`
}

func (m *MsgSubmitProofResponse) Reset()         { *m = MsgSubmitProofResponse{} }
func (m *MsgSubmitProofResponse) String() string { return "MsgSubmitProofResponse" }
func (m *MsgSubmitProofResponse) ProtoMessage()  {}

type MsgClaimRewardsResponse struct {
	Amount sdk.Coins `protobuf:"bytes,1,rep,name=amount,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"amount"`
}

func (m *MsgClaimRewardsResponse) Reset()         { *m = MsgClaimRewardsResponse{} }
func (m *MsgClaimRewardsResponse) String() string { return "MsgClaimRewardsResponse" }
func (m *MsgClaimRewardsResponse) ProtoMessage()  {}

type MsgCancelJobResponse struct {
	Success bool `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
}

func (m *MsgCancelJobResponse) Reset()         { *m = MsgCancelJobResponse{} }
func (m *MsgCancelJobResponse) String() string { return "MsgCancelJobResponse" }
func (m *MsgCancelJobResponse) ProtoMessage()  {}

// Server interfaces
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

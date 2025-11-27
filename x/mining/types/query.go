package types

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc"
)

type QueryParamsRequest struct{}

type QueryParamsResponse struct {
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
}

func (m *QueryParamsResponse) Reset()         { *m = QueryParamsResponse{} }
func (m *QueryParamsResponse) String() string { return "QueryParamsResponse" }
func (m *QueryParamsResponse) ProtoMessage()  {}

type QueryJobRequest struct {
	JobId string `protobuf:"bytes,1,opt,name=job_id,json=jobId,proto3" json:"job_id"`
}

type QueryJobResponse struct {
	Job Job `protobuf:"bytes,1,opt,name=job,proto3" json:"job"`
}

func (m *QueryJobResponse) Reset()         { *m = QueryJobResponse{} }
func (m *QueryJobResponse) String() string { return "QueryJobResponse" }
func (m *QueryJobResponse) ProtoMessage()  {}

type QueryJobsRequest struct {
	Status uint32 `protobuf:"varint,1,opt,name=status,proto3" json:"status"`
}

type QueryJobsResponse struct {
	Jobs []Job `protobuf:"bytes,1,rep,name=jobs,proto3" json:"jobs"`
}

func (m *QueryJobsResponse) Reset()         { *m = QueryJobsResponse{} }
func (m *QueryJobsResponse) String() string { return "QueryJobsResponse" }
func (m *QueryJobsResponse) ProtoMessage()  {}

type QueryMinerSharesRequest struct {
	Miner string `protobuf:"bytes,1,opt,name=miner,proto3" json:"miner"`
	JobId string `protobuf:"bytes,2,opt,name=job_id,json=jobId,proto3" json:"job_id"`
}

type QueryMinerSharesResponse struct {
	Miner  string `protobuf:"bytes,1,opt,name=miner,proto3" json:"miner"`
	JobId  string `protobuf:"bytes,2,opt,name=job_id,json=jobId,proto3" json:"job_id"`
	Shares int64  `protobuf:"varint,3,opt,name=shares,proto3" json:"shares"`
}

func (m *QueryMinerSharesResponse) Reset()         { *m = QueryMinerSharesResponse{} }
func (m *QueryMinerSharesResponse) String() string { return "QueryMinerSharesResponse" }
func (m *QueryMinerSharesResponse) ProtoMessage()  {}

type QueryMinerStatsRequest struct {
	MinerAddress string `protobuf:"bytes,1,opt,name=miner_address,json=minerAddress,proto3" json:"miner_address"`
}

type QueryMinerStatsResponse struct {
	MinerAddress     string         `protobuf:"bytes,1,opt,name=miner_address,json=minerAddress,proto3" json:"miner_address"`
	TotalShares      int64          `protobuf:"varint,2,opt,name=total_shares,json=totalShares,proto3" json:"total_shares"`
	PendingRewards   sdk.Coins      `protobuf:"bytes,3,rep,name=pending_rewards,json=pendingRewards,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"pending_rewards"`
	JobsParticipated int64          `protobuf:"varint,4,opt,name=jobs_participated,json=jobsParticipated,proto3" json:"jobs_participated"`
	TotalClaimed     sdk.Coins      `protobuf:"bytes,5,rep,name=total_claimed,json=totalClaimed,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"total_claimed"`
	ActiveJobs       []MinerJobInfo `protobuf:"bytes,6,rep,name=active_jobs,json=activeJobs,proto3" json:"active_jobs"`
}

func (m *QueryMinerStatsResponse) Reset()         { *m = QueryMinerStatsResponse{} }
func (m *QueryMinerStatsResponse) String() string { return "QueryMinerStatsResponse" }
func (m *QueryMinerStatsResponse) ProtoMessage()  {}

type MinerJobInfo struct {
	JobId       string `protobuf:"bytes,1,opt,name=job_id,json=jobId,proto3" json:"job_id"`
	Shares      int64  `protobuf:"varint,2,opt,name=shares,proto3" json:"shares"`
	BestEnergy  int64  `protobuf:"varint,3,opt,name=best_energy,json=bestEnergy,proto3" json:"best_energy"`
	TotalShares int64  `protobuf:"varint,4,opt,name=total_shares,json=totalShares,proto3" json:"total_shares"`
}

func (m *MinerJobInfo) Reset()         { *m = MinerJobInfo{} }
func (m *MinerJobInfo) String() string { return m.JobId }
func (m *MinerJobInfo) ProtoMessage()  {}

type QueryActiveJobRequest struct{}

type QueryActiveJobResponse struct {
	Job           *Job  `protobuf:"bytes,1,opt,name=job,proto3" json:"job,omitempty"`
	TimeRemaining int64 `protobuf:"varint,2,opt,name=time_remaining,json=timeRemaining,proto3" json:"time_remaining"`
	ProblemSize   int64 `protobuf:"varint,3,opt,name=problem_size,json=problemSize,proto3" json:"problem_size"`
}

func (m *QueryActiveJobResponse) Reset()         { *m = QueryActiveJobResponse{} }
func (m *QueryActiveJobResponse) String() string { return "QueryActiveJobResponse" }
func (m *QueryActiveJobResponse) ProtoMessage()  {}

type QueryQueueStatusRequest struct{}

type QueryQueueStatusResponse struct {
	PaidQueueLength   int64           `protobuf:"varint,1,opt,name=paid_queue_length,json=paidQueueLength,proto3" json:"paid_queue_length"`
	PublicQueueLength int64           `protobuf:"varint,2,opt,name=public_queue_length,json=publicQueueLength,proto3" json:"public_queue_length"`
	NextPaidJobs      []QueuedJobInfo `protobuf:"bytes,3,rep,name=next_paid_jobs,json=nextPaidJobs,proto3" json:"next_paid_jobs"`
	NextPublicJobs    []string        `protobuf:"bytes,4,rep,name=next_public_jobs,json=nextPublicJobs,proto3" json:"next_public_jobs"`
}

func (m *QueryQueueStatusResponse) Reset()         { *m = QueryQueueStatusResponse{} }
func (m *QueryQueueStatusResponse) String() string { return "QueryQueueStatusResponse" }
func (m *QueryQueueStatusResponse) ProtoMessage()  {}

type QueuedJobInfo struct {
	JobId       string `protobuf:"bytes,1,opt,name=job_id,json=jobId,proto3" json:"job_id"`
	Customer    string `protobuf:"bytes,2,opt,name=customer,proto3" json:"customer"`
	PriorityFee int64  `protobuf:"varint,3,opt,name=priority_fee,json=priorityFee,proto3" json:"priority_fee"`
	Reward      int64  `protobuf:"varint,4,opt,name=reward,proto3" json:"reward"`
}

func (q *QueuedJobInfo) Reset()         { *q = QueuedJobInfo{} }
func (q *QueuedJobInfo) String() string { return q.JobId }
func (q *QueuedJobInfo) ProtoMessage()  {}

type QueryEmissionInfoRequest struct{}

type QueryEmissionInfoResponse struct {
	CurrentEpoch     int32 `protobuf:"varint,1,opt,name=current_epoch,json=currentEpoch,proto3" json:"current_epoch"`
	EmissionRate     int64 `protobuf:"varint,2,opt,name=emission_rate,json=emissionRate,proto3" json:"emission_rate"`
	EmissionEscrow   int64 `protobuf:"varint,3,opt,name=emission_escrow,json=emissionEscrow,proto3" json:"emission_escrow"`
	GenesisTime      int64 `protobuf:"varint,4,opt,name=genesis_time,json=genesisTime,proto3" json:"genesis_time"`
	MinutesIntoEpoch int64 `protobuf:"varint,5,opt,name=minutes_into_epoch,json=minutesIntoEpoch,proto3" json:"minutes_into_epoch"`
	MinutesUntilNext int64 `protobuf:"varint,6,opt,name=minutes_until_next,json=minutesUntilNext,proto3" json:"minutes_until_next"`
	EpochDuration    int64 `protobuf:"varint,7,opt,name=epoch_duration,json=epochDuration,proto3" json:"epoch_duration"`
	NextEpochRate    int64 `protobuf:"varint,8,opt,name=next_epoch_rate,json=nextEpochRate,proto3" json:"next_epoch_rate"`
}

func (m *QueryEmissionInfoResponse) Reset()         { *m = QueryEmissionInfoResponse{} }
func (m *QueryEmissionInfoResponse) String() string { return "QueryEmissionInfoResponse" }
func (m *QueryEmissionInfoResponse) ProtoMessage()  {}

type QueryValidatorMiningRecordRequest struct {
	Validator string `protobuf:"bytes,1,opt,name=validator,proto3" json:"validator"`
}

type QueryValidatorMiningRecordResponse struct {
	Record ValidatorMiningRecord `protobuf:"bytes,1,opt,name=record,proto3" json:"record"`
}

func (m *QueryValidatorMiningRecordResponse) Reset() { *m = QueryValidatorMiningRecordResponse{} }
func (m *QueryValidatorMiningRecordResponse) String() string {
	return "QueryValidatorMiningRecordResponse"
}
func (m *QueryValidatorMiningRecordResponse) ProtoMessage() {}

type QueryCheckpointRequest struct {
	Id uint64 `protobuf:"varint,1,opt,name=id,proto3" json:"id"`
}

type QueryCheckpointResponse struct {
	Checkpoint Checkpoint `protobuf:"bytes,1,opt,name=checkpoint,proto3" json:"checkpoint"`
}

func (m *QueryCheckpointResponse) Reset()         { *m = QueryCheckpointResponse{} }
func (m *QueryCheckpointResponse) String() string { return "QueryCheckpointResponse" }
func (m *QueryCheckpointResponse) ProtoMessage()  {}

type QueryLatestCheckpointRequest struct{}

type QueryLatestCheckpointResponse struct {
	Checkpoint Checkpoint `protobuf:"bytes,1,opt,name=checkpoint,proto3" json:"checkpoint"`
}

func (m *QueryLatestCheckpointResponse) Reset()         { *m = QueryLatestCheckpointResponse{} }
func (m *QueryLatestCheckpointResponse) String() string { return "QueryLatestCheckpointResponse" }
func (m *QueryLatestCheckpointResponse) ProtoMessage()  {}

type MsgPostJobResponse struct {
	JobId         string `protobuf:"bytes,1,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
	QueuePosition int64  `protobuf:"varint,2,opt,name=queue_position,json=queuePosition,proto3" json:"queue_position,omitempty"`
}

func (m *MsgPostJobResponse) Reset()         { *m = MsgPostJobResponse{} }
func (m *MsgPostJobResponse) String() string { return "MsgPostJobResponse" }
func (m *MsgPostJobResponse) ProtoMessage()  {}

type MsgSubmitProofResponse struct {
	Accepted bool  `protobuf:"varint,1,opt,name=accepted,proto3" json:"accepted,omitempty"`
	Shares   int64 `protobuf:"varint,2,opt,name=shares,proto3" json:"shares,omitempty"`
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

type MsgSubmitPublicJobResponse struct {
	JobId         string `protobuf:"bytes,1,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
	QueuePosition int64  `protobuf:"varint,2,opt,name=queue_position,json=queuePosition,proto3" json:"queue_position,omitempty"`
}

func (m *MsgSubmitPublicJobResponse) Reset()         { *m = MsgSubmitPublicJobResponse{} }
func (m *MsgSubmitPublicJobResponse) String() string { return "MsgSubmitPublicJobResponse" }
func (m *MsgSubmitPublicJobResponse) ProtoMessage()  {}

func RegisterQueryServer(s grpc.ServiceRegistrar, srv QueryServer) {
	s.RegisterService(&QueryServiceDesc, srv)
}

func RegisterMsgServer(s grpc.ServiceRegistrar, srv MsgServer) {
	s.RegisterService(&MsgServiceDesc, srv)
}

var QueryServiceDesc = grpc.ServiceDesc{
	ServiceName: "nexus.mining.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "Params", Handler: _Query_Params_Handler},
		{MethodName: "Job", Handler: _Query_Job_Handler},
		{MethodName: "Jobs", Handler: _Query_Jobs_Handler},
		{MethodName: "MinerShares", Handler: _Query_MinerShares_Handler},
		{MethodName: "MinerStats", Handler: _Query_MinerStats_Handler},
		{MethodName: "ActiveJob", Handler: _Query_ActiveJob_Handler},
		{MethodName: "QueueStatus", Handler: _Query_QueueStatus_Handler},
		{MethodName: "EmissionInfo", Handler: _Query_EmissionInfo_Handler},
		{MethodName: "ValidatorMiningRecord", Handler: _Query_ValidatorMiningRecord_Handler},
		{MethodName: "Checkpoint", Handler: _Query_Checkpoint_Handler},
		{MethodName: "LatestCheckpoint", Handler: _Query_LatestCheckpoint_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "nexus/mining/v1/query.proto",
}

var MsgServiceDesc = grpc.ServiceDesc{
	ServiceName: "nexus.mining.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "SubmitWork", Handler: _Msg_SubmitWork_Handler},
		{MethodName: "PostJob", Handler: _Msg_PostJob_Handler},
		{MethodName: "SubmitProof", Handler: _Msg_SubmitProof_Handler},
		{MethodName: "ClaimRewards", Handler: _Msg_ClaimRewards_Handler},
		{MethodName: "CancelJob", Handler: _Msg_CancelJob_Handler},
		{MethodName: "SubmitPublicJob", Handler: _Msg_SubmitPublicJob_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "nexus/mining/v1/tx.proto",
}

func _Query_Params_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Params(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/nexus.mining.v1.Query/Params"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Params(ctx, req.(*QueryParamsRequest))
	})
}

func _Query_Job_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryJobRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Job(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/nexus.mining.v1.Query/Job"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Job(ctx, req.(*QueryJobRequest))
	})
}

func _Query_Jobs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryJobsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Jobs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/nexus.mining.v1.Query/Jobs"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Jobs(ctx, req.(*QueryJobsRequest))
	})
}

func _Query_MinerShares_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryMinerSharesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).MinerShares(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/nexus.mining.v1.Query/MinerShares"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).MinerShares(ctx, req.(*QueryMinerSharesRequest))
	})
}

func _Query_MinerStats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryMinerStatsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).MinerStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/nexus.mining.v1.Query/MinerStats"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).MinerStats(ctx, req.(*QueryMinerStatsRequest))
	})
}

func _Query_ActiveJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryActiveJobRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ActiveJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/nexus.mining.v1.Query/ActiveJob"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ActiveJob(ctx, req.(*QueryActiveJobRequest))
	})
}

func _Query_QueueStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryQueueStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).QueueStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/nexus.mining.v1.Query/QueueStatus"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).QueueStatus(ctx, req.(*QueryQueueStatusRequest))
	})
}

func _Query_EmissionInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryEmissionInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).EmissionInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/nexus.mining.v1.Query/EmissionInfo"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).EmissionInfo(ctx, req.(*QueryEmissionInfoRequest))
	})
}

func _Query_ValidatorMiningRecord_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryValidatorMiningRecordRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ValidatorMiningRecord(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/nexus.mining.v1.Query/ValidatorMiningRecord"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ValidatorMiningRecord(ctx, req.(*QueryValidatorMiningRecordRequest))
	})
}

func _Query_Checkpoint_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryCheckpointRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Checkpoint(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/nexus.mining.v1.Query/Checkpoint"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Checkpoint(ctx, req.(*QueryCheckpointRequest))
	})
}

func _Query_LatestCheckpoint_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryLatestCheckpointRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).LatestCheckpoint(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/nexus.mining.v1.Query/LatestCheckpoint"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).LatestCheckpoint(ctx, req.(*QueryLatestCheckpointRequest))
	})
}

func _Msg_SubmitWork_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSubmitWork)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SubmitWork(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/nexus.mining.v1.Msg/SubmitWork"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SubmitWork(ctx, req.(*MsgSubmitWork))
	})
}

func _Msg_PostJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgPostJob)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).PostJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/nexus.mining.v1.Msg/PostJob"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).PostJob(ctx, req.(*MsgPostJob))
	})
}

func _Msg_SubmitProof_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSubmitProof)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SubmitProof(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/nexus.mining.v1.Msg/SubmitProof"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SubmitProof(ctx, req.(*MsgSubmitProof))
	})
}

func _Msg_ClaimRewards_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgClaimRewards)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).ClaimRewards(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/nexus.mining.v1.Msg/ClaimRewards"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).ClaimRewards(ctx, req.(*MsgClaimRewards))
	})
}

func _Msg_CancelJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCancelJob)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CancelJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/nexus.mining.v1.Msg/CancelJob"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CancelJob(ctx, req.(*MsgCancelJob))
	})
}

func _Msg_SubmitPublicJob_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSubmitPublicJob)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SubmitPublicJob(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/nexus.mining.v1.Msg/SubmitPublicJob"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SubmitPublicJob(ctx, req.(*MsgSubmitPublicJob))
	})
}

type MsgServer interface {
	SubmitWork(context.Context, *MsgSubmitWork) (*MsgSubmitWorkResponse, error)
	PostJob(context.Context, *MsgPostJob) (*MsgPostJobResponse, error)
	SubmitProof(context.Context, *MsgSubmitProof) (*MsgSubmitProofResponse, error)
	ClaimRewards(context.Context, *MsgClaimRewards) (*MsgClaimRewardsResponse, error)
	CancelJob(context.Context, *MsgCancelJob) (*MsgCancelJobResponse, error)
	SubmitPublicJob(context.Context, *MsgSubmitPublicJob) (*MsgSubmitPublicJobResponse, error)
}

type QueryServer interface {
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	Job(context.Context, *QueryJobRequest) (*QueryJobResponse, error)
	Jobs(context.Context, *QueryJobsRequest) (*QueryJobsResponse, error)
	MinerShares(context.Context, *QueryMinerSharesRequest) (*QueryMinerSharesResponse, error)
	MinerStats(context.Context, *QueryMinerStatsRequest) (*QueryMinerStatsResponse, error)
	ActiveJob(context.Context, *QueryActiveJobRequest) (*QueryActiveJobResponse, error)
	QueueStatus(context.Context, *QueryQueueStatusRequest) (*QueryQueueStatusResponse, error)
	EmissionInfo(context.Context, *QueryEmissionInfoRequest) (*QueryEmissionInfoResponse, error)
	ValidatorMiningRecord(context.Context, *QueryValidatorMiningRecordRequest) (*QueryValidatorMiningRecordResponse, error)
	Checkpoint(context.Context, *QueryCheckpointRequest) (*QueryCheckpointResponse, error)
	LatestCheckpoint(context.Context, *QueryLatestCheckpointRequest) (*QueryLatestCheckpointResponse, error)
}

func NewQueryClient(clientCtx client.Context) QueryClient {
	return &queryClient{clientCtx: clientCtx}
}

type QueryClient interface {
	Params(ctx context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error)
	Job(ctx context.Context, req *QueryJobRequest) (*QueryJobResponse, error)
	Jobs(ctx context.Context, req *QueryJobsRequest) (*QueryJobsResponse, error)
	MinerShares(ctx context.Context, req *QueryMinerSharesRequest) (*QueryMinerSharesResponse, error)
	MinerStats(ctx context.Context, req *QueryMinerStatsRequest) (*QueryMinerStatsResponse, error)
	ActiveJob(ctx context.Context, req *QueryActiveJobRequest) (*QueryActiveJobResponse, error)
	QueueStatus(ctx context.Context, req *QueryQueueStatusRequest) (*QueryQueueStatusResponse, error)
	EmissionInfo(ctx context.Context, req *QueryEmissionInfoRequest) (*QueryEmissionInfoResponse, error)
	ValidatorMiningRecord(ctx context.Context, req *QueryValidatorMiningRecordRequest) (*QueryValidatorMiningRecordResponse, error)
	Checkpoint(ctx context.Context, req *QueryCheckpointRequest) (*QueryCheckpointResponse, error)
	LatestCheckpoint(ctx context.Context, req *QueryLatestCheckpointRequest) (*QueryLatestCheckpointResponse, error)
}

type queryClient struct {
	clientCtx client.Context
}

func (q *queryClient) Params(ctx context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error) {
	out := new(QueryParamsResponse)
	err := q.clientCtx.Invoke(ctx, "/nexus.mining.v1.Query/Params", req, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (q *queryClient) Job(ctx context.Context, req *QueryJobRequest) (*QueryJobResponse, error) {
	out := new(QueryJobResponse)
	err := q.clientCtx.Invoke(ctx, "/nexus.mining.v1.Query/Job", req, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (q *queryClient) Jobs(ctx context.Context, req *QueryJobsRequest) (*QueryJobsResponse, error) {
	out := new(QueryJobsResponse)
	err := q.clientCtx.Invoke(ctx, "/nexus.mining.v1.Query/Jobs", req, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (q *queryClient) MinerShares(ctx context.Context, req *QueryMinerSharesRequest) (*QueryMinerSharesResponse, error) {
	out := new(QueryMinerSharesResponse)
	err := q.clientCtx.Invoke(ctx, "/nexus.mining.v1.Query/MinerShares", req, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (q *queryClient) MinerStats(ctx context.Context, req *QueryMinerStatsRequest) (*QueryMinerStatsResponse, error) {
	out := new(QueryMinerStatsResponse)
	err := q.clientCtx.Invoke(ctx, "/nexus.mining.v1.Query/MinerStats", req, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (q *queryClient) ActiveJob(ctx context.Context, req *QueryActiveJobRequest) (*QueryActiveJobResponse, error) {
	out := new(QueryActiveJobResponse)
	err := q.clientCtx.Invoke(ctx, "/nexus.mining.v1.Query/ActiveJob", req, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (q *queryClient) QueueStatus(ctx context.Context, req *QueryQueueStatusRequest) (*QueryQueueStatusResponse, error) {
	out := new(QueryQueueStatusResponse)
	err := q.clientCtx.Invoke(ctx, "/nexus.mining.v1.Query/QueueStatus", req, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (q *queryClient) EmissionInfo(ctx context.Context, req *QueryEmissionInfoRequest) (*QueryEmissionInfoResponse, error) {
	out := new(QueryEmissionInfoResponse)
	err := q.clientCtx.Invoke(ctx, "/nexus.mining.v1.Query/EmissionInfo", req, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (q *queryClient) ValidatorMiningRecord(ctx context.Context, req *QueryValidatorMiningRecordRequest) (*QueryValidatorMiningRecordResponse, error) {
	out := new(QueryValidatorMiningRecordResponse)
	err := q.clientCtx.Invoke(ctx, "/nexus.mining.v1.Query/ValidatorMiningRecord", req, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (q *queryClient) Checkpoint(ctx context.Context, req *QueryCheckpointRequest) (*QueryCheckpointResponse, error) {
	out := new(QueryCheckpointResponse)
	err := q.clientCtx.Invoke(ctx, "/nexus.mining.v1.Query/Checkpoint", req, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (q *queryClient) LatestCheckpoint(ctx context.Context, req *QueryLatestCheckpointRequest) (*QueryLatestCheckpointResponse, error) {
	out := new(QueryLatestCheckpointResponse)
	err := q.clientCtx.Invoke(ctx, "/nexus.mining.v1.Query/LatestCheckpoint", req, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
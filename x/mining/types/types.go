package types

import (
	"time"
)

const (
	ModuleName  = "mining"
	StoreKey    = ModuleName
	MemStoreKey = "mem_" + ModuleName
	RouterKey   = ModuleName
)

var (
	JobKeyPrefix             = []byte{0x01}
	CheckpointKeyPrefix      = []byte{0x02}
	ShareKeyPrefix           = []byte{0x03}
	ValidatorRecordKeyPrefix = []byte{0x04}
	LastCheckpointIDKey      = []byte{0x05}
)

type JobStatus uint32

const (
	JobStatusPending   JobStatus = 0
	JobStatusActive    JobStatus = 1
	JobStatusCompleted JobStatus = 2
	JobStatusExpired   JobStatus = 3
	JobStatusCancelled JobStatus = 4
)

type ProofType uint32

const (
	ProofTypeNova  ProofType = 0
	ProofTypeSTARK ProofType = 1
)

type Job struct {
	Id           string    `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Customer     string    `protobuf:"bytes,2,opt,name=customer,proto3" json:"customer,omitempty"`
	ProblemType  string    `protobuf:"bytes,3,opt,name=problem_type,json=problemType,proto3" json:"problem_type,omitempty"`
	ProblemData  []byte    `protobuf:"bytes,4,opt,name=problem_data,json=problemData,proto3" json:"problem_data,omitempty"`
	ProblemHash  string    `protobuf:"bytes,5,opt,name=problem_hash,json=problemHash,proto3" json:"problem_hash,omitempty"`
	Threshold    int64     `protobuf:"varint,6,opt,name=threshold,proto3" json:"threshold,omitempty"`
	Reward       int64     `protobuf:"varint,7,opt,name=reward,proto3" json:"reward,omitempty"`
	Status       JobStatus `protobuf:"varint,8,opt,name=status,proto3,casttype=JobStatus" json:"status,omitempty"`
	BestEnergy   int64     `protobuf:"varint,9,opt,name=best_energy,json=bestEnergy,proto3" json:"best_energy,omitempty"`
	BestSolver   string    `protobuf:"bytes,10,opt,name=best_solver,json=bestSolver,proto3" json:"best_solver,omitempty"`
	TotalShares  int64     `protobuf:"varint,11,opt,name=total_shares,json=totalShares,proto3" json:"total_shares,omitempty"`
	CreatedAt    int64     `protobuf:"varint,12,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	Deadline     int64     `protobuf:"varint,13,opt,name=deadline,proto3" json:"deadline,omitempty"`
	IsBackground bool      `protobuf:"varint,14,opt,name=is_background,json=isBackground,proto3" json:"is_background,omitempty"`
}

func (j *Job) Reset()         { *j = Job{} }
func (j *Job) String() string { return j.Id }
func (j *Job) ProtoMessage()  {}

type Checkpoint struct {
	Id               uint64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	StartHeight      int64  `protobuf:"varint,2,opt,name=start_height,json=startHeight,proto3" json:"start_height,omitempty"`
	EndHeight        int64  `protobuf:"varint,3,opt,name=end_height,json=endHeight,proto3" json:"end_height,omitempty"`
	JobsCompleted    uint64 `protobuf:"varint,4,opt,name=jobs_completed,json=jobsCompleted,proto3" json:"jobs_completed,omitempty"`
	TotalShares      int64  `protobuf:"varint,5,opt,name=total_shares,json=totalShares,proto3" json:"total_shares,omitempty"`
	MinerRewards     int64  `protobuf:"varint,6,opt,name=miner_rewards,json=minerRewards,proto3" json:"miner_rewards,omitempty"`
	ValidatorRewards int64  `protobuf:"varint,7,opt,name=validator_rewards,json=validatorRewards,proto3" json:"validator_rewards,omitempty"`
	StateRoot        string `protobuf:"bytes,8,opt,name=state_root,json=stateRoot,proto3" json:"state_root,omitempty"`
	Timestamp        int64  `protobuf:"varint,9,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (c *Checkpoint) Reset()         { *c = Checkpoint{} }
func (c *Checkpoint) String() string { return "" }
func (c *Checkpoint) ProtoMessage()  {}

type ValidatorMiningRecord struct {
	Validator     string    `protobuf:"bytes,1,opt,name=validator,proto3" json:"validator,omitempty"`
	LastProofTime time.Time `protobuf:"bytes,2,opt,name=last_proof_time,json=lastProofTime,proto3,stdtime" json:"last_proof_time"`
	TotalProofs   uint64    `protobuf:"varint,3,opt,name=total_proofs,json=totalProofs,proto3" json:"total_proofs,omitempty"`
	TotalShares   int64     `protobuf:"varint,4,opt,name=total_shares,json=totalShares,proto3" json:"total_shares,omitempty"`
	IsEligible    bool      `protobuf:"varint,5,opt,name=is_eligible,json=isEligible,proto3" json:"is_eligible,omitempty"`
}

func (v *ValidatorMiningRecord) Reset()         { *v = ValidatorMiningRecord{} }
func (v *ValidatorMiningRecord) String() string { return v.Validator }
func (v *ValidatorMiningRecord) ProtoMessage()  {}

type ShareRecord struct {
	Miner     string `protobuf:"bytes,1,opt,name=miner,proto3" json:"miner,omitempty"`
	JobId     string `protobuf:"bytes,2,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
	Shares    int64  `protobuf:"varint,3,opt,name=shares,proto3" json:"shares,omitempty"`
	Timestamp int64  `protobuf:"varint,4,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (s *ShareRecord) Reset()         { *s = ShareRecord{} }
func (s *ShareRecord) String() string { return s.Miner }
func (s *ShareRecord) ProtoMessage()  {}

type MiningProof struct {
	Miner        string    `protobuf:"bytes,1,opt,name=miner,proto3" json:"miner,omitempty"`
	JobId        string    `protobuf:"bytes,2,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
	Energy       int64     `protobuf:"varint,3,opt,name=energy,proto3" json:"energy,omitempty"`
	Proof        []byte    `protobuf:"bytes,4,opt,name=proof,proto3" json:"proof,omitempty"`
	ProofType    ProofType `protobuf:"varint,5,opt,name=proof_type,json=proofType,proto3,casttype=ProofType" json:"proof_type,omitempty"`
	SolutionHash string    `protobuf:"bytes,6,opt,name=solution_hash,json=solutionHash,proto3" json:"solution_hash,omitempty"`
	Height       int64     `protobuf:"varint,7,opt,name=height,proto3" json:"height,omitempty"`
	Timestamp    int64     `protobuf:"varint,8,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	SharesEarned int64     `protobuf:"varint,9,opt,name=shares_earned,json=sharesEarned,proto3" json:"shares_earned,omitempty"`
}

func (m *MiningProof) Reset()         { *m = MiningProof{} }
func (m *MiningProof) String() string { return m.Miner }
func (m *MiningProof) ProtoMessage()  {}

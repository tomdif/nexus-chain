package types

// Job Status
type JobStatus uint32

const (
	JobStatusPending   JobStatus = 0
	JobStatusActive    JobStatus = 1
	JobStatusCompleted JobStatus = 2
	JobStatusExpired   JobStatus = 3
	JobStatusCancelled JobStatus = 4
	JobStatusQueued    JobStatus = 5
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
	PriorityFee  int64     `protobuf:"varint,15,opt,name=priority_fee,json=priorityFee,proto3" json:"priority_fee,omitempty"`
	Title        string    `protobuf:"bytes,16,opt,name=title,proto3" json:"title,omitempty"`
	IpfsCid      string    `protobuf:"bytes,17,opt,name=ipfs_cid,json=ipfsCid,proto3" json:"ipfs_cid,omitempty"`
}

func (j *Job) Reset()         { *j = Job{} }
func (j *Job) String() string { return j.Id }
func (j *Job) ProtoMessage()  {}

type Checkpoint struct {
	Id               uint64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	StartHeight      int64  `protobuf:"varint,2,opt,name=start_height,json=startHeight,proto3" json:"start_height,omitempty"`
	EndHeight        int64  `protobuf:"varint,3,opt,name=end_height,json=endHeight,proto3" json:"end_height,omitempty"`
	ValidatorRewards int64  `protobuf:"varint,4,opt,name=validator_rewards,json=validatorRewards,proto3" json:"validator_rewards,omitempty"`
	Timestamp        int64  `protobuf:"varint,5,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
}

func (c *Checkpoint) Reset()         { *c = Checkpoint{} }
func (c *Checkpoint) String() string { return "" }
func (c *Checkpoint) ProtoMessage()  {}

// ValidatorMiningRecord tracks a validator's mining activity
type ValidatorMiningRecord struct {
	ValidatorAddr  string `protobuf:"bytes,1,opt,name=validator_addr,json=validatorAddr,proto3" json:"validator_addr,omitempty"`
	TotalShares    int64  `protobuf:"varint,2,opt,name=total_shares,json=totalShares,proto3" json:"total_shares,omitempty"`
	TotalRewards   int64  `protobuf:"varint,3,opt,name=total_rewards,json=totalRewards,proto3" json:"total_rewards,omitempty"`
	JobsCompleted  int64  `protobuf:"varint,4,opt,name=jobs_completed,json=jobsCompleted,proto3" json:"jobs_completed,omitempty"`
	LastActiveTime int64  `protobuf:"varint,5,opt,name=last_active_time,json=lastActiveTime,proto3" json:"last_active_time,omitempty"`
}

func (v *ValidatorMiningRecord) Reset()         { *v = ValidatorMiningRecord{} }
func (v *ValidatorMiningRecord) String() string { return v.ValidatorAddr }
func (v *ValidatorMiningRecord) ProtoMessage()  {}

package types

import (
	"time"
)

const (
	ModuleName = "mining"
	StoreKey   = ModuleName
	MemStoreKey = "mem_" + ModuleName
	RouterKey  = ModuleName
)

var (
	JobKeyPrefix        = []byte{0x01}
	CheckpointKeyPrefix = []byte{0x02}
	ShareKeyPrefix      = []byte{0x03}
	ValidatorRecordKeyPrefix = []byte{0x04}
	LastCheckpointIDKey = []byte{0x05}
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
	Id          string    `json:"id"`
	Customer    string    `json:"customer"`
	ProblemType string    `json:"problem_type"`
	ProblemData []byte    `json:"problem_data"`
	ProblemHash string    `json:"problem_hash"`
	Threshold   int64     `json:"threshold"`
	Reward      int64     `json:"reward"`
	Status      JobStatus `json:"status"`
	BestEnergy  int64     `json:"best_energy"`
	BestSolver  string    `json:"best_solver"`
	TotalShares int64     `json:"total_shares"`
	CreatedAt   int64     `json:"created_at"`
	Deadline    int64     `json:"deadline"`
	IsBackground bool     `json:"is_background"`
}

func (j *Job) Reset()         { *j = Job{} }
func (j *Job) String() string { return j.Id }
func (j *Job) ProtoMessage()  {}

type Checkpoint struct {
	Id               uint64 `json:"id"`
	StartHeight      int64  `json:"start_height"`
	EndHeight        int64  `json:"end_height"`
	JobsCompleted    uint64 `json:"jobs_completed"`
	TotalShares      int64  `json:"total_shares"`
	MinerRewards     int64  `json:"miner_rewards"`
	ValidatorRewards int64  `json:"validator_rewards"`
	StateRoot        string `json:"state_root"`
	Timestamp        int64  `json:"timestamp"`
}

func (c *Checkpoint) Reset()         { *c = Checkpoint{} }
func (c *Checkpoint) String() string { return "" }
func (c *Checkpoint) ProtoMessage()  {}

type ValidatorMiningRecord struct {
	Validator     string    `json:"validator"`
	LastProofTime time.Time `json:"last_proof_time"`
	TotalProofs   uint64    `json:"total_proofs"`
	TotalShares   int64     `json:"total_shares"`
	IsEligible    bool      `json:"is_eligible"`
}

func (v *ValidatorMiningRecord) Reset()         { *v = ValidatorMiningRecord{} }
func (v *ValidatorMiningRecord) String() string { return v.Validator }
func (v *ValidatorMiningRecord) ProtoMessage()  {}

type ShareRecord struct {
	Miner     string `json:"miner"`
	JobId     string `json:"job_id"`
	Shares    int64  `json:"shares"`
	Timestamp int64  `json:"timestamp"`
}

func (s *ShareRecord) Reset()         { *s = ShareRecord{} }
func (s *ShareRecord) String() string { return s.Miner }
func (s *ShareRecord) ProtoMessage()  {}

type MiningProof struct {
	Miner        string    `json:"miner"`
	JobId        string    `json:"job_id"`
	Energy       int64     `json:"energy"`
	Proof        []byte    `json:"proof"`
	ProofType    ProofType `json:"proof_type"`
	SolutionHash string    `json:"solution_hash"`
	Height       int64     `json:"height"`
	Timestamp    int64     `json:"timestamp"`
	SharesEarned int64     `json:"shares_earned"`
}

func (m *MiningProof) Reset()         { *m = MiningProof{} }
func (m *MiningProof) String() string { return m.Miner }
func (m *MiningProof) ProtoMessage()  {}

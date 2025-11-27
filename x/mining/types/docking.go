package types

// DockingJob represents a molecular docking job
type DockingJob struct {
	Id            string  `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	ProteinId     string  `protobuf:"bytes,2,opt,name=protein_id,json=proteinId,proto3" json:"protein_id,omitempty"`
	TargetHash    string  `protobuf:"bytes,3,opt,name=target_hash,json=targetHash,proto3" json:"target_hash,omitempty"`
	ProteinPDB    string  `protobuf:"bytes,4,opt,name=protein_pdb,json=proteinPdb,proto3" json:"protein_pdb,omitempty"`
	TotalLigands  int64   `protobuf:"varint,5,opt,name=total_ligands,json=totalLigands,proto3" json:"total_ligands,omitempty"`
	DockedCount   int64   `protobuf:"varint,6,opt,name=docked_count,json=dockedCount,proto3" json:"docked_count,omitempty"`
	HitCount      int64   `protobuf:"varint,7,opt,name=hit_count,json=hitCount,proto3" json:"hit_count,omitempty"`
	CenterX       float64 `protobuf:"fixed64,8,opt,name=center_x,json=centerX,proto3" json:"center_x,omitempty"`
	CenterY       float64 `protobuf:"fixed64,9,opt,name=center_y,json=centerY,proto3" json:"center_y,omitempty"`
	CenterZ       float64 `protobuf:"fixed64,10,opt,name=center_z,json=centerZ,proto3" json:"center_z,omitempty"`
	SizeX         float64 `protobuf:"fixed64,11,opt,name=size_x,json=sizeX,proto3" json:"size_x,omitempty"`
	SizeY         float64 `protobuf:"fixed64,12,opt,name=size_y,json=sizeY,proto3" json:"size_y,omitempty"`
	SizeZ         float64 `protobuf:"fixed64,13,opt,name=size_z,json=sizeZ,proto3" json:"size_z,omitempty"`
	IsBackground  bool    `protobuf:"varint,14,opt,name=is_background,json=isBackground,proto3" json:"is_background,omitempty"`
	Status        string  `protobuf:"bytes,15,opt,name=status,proto3" json:"status,omitempty"`
	CreatedAt     int64   `protobuf:"varint,16,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	Deadline      int64   `protobuf:"varint,17,opt,name=deadline,proto3" json:"deadline,omitempty"`
	RewardPool    int64   `protobuf:"varint,18,opt,name=reward_pool,json=rewardPool,proto3" json:"reward_pool,omitempty"`
	NextLigandIdx int64   `protobuf:"varint,19,opt,name=next_ligand_idx,json=nextLigandIdx,proto3" json:"next_ligand_idx,omitempty"`
	License       string  `protobuf:"bytes,20,opt,name=license,proto3" json:"license,omitempty"`
}

func (m *DockingJob) Reset()         { *m = DockingJob{} }
func (m *DockingJob) String() string { return "DockingJob" }
func (m *DockingJob) ProtoMessage()  {}

// DockingResult represents a single ligand docking result
type DockingResult struct {
	Id             string  `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	JobId          string  `protobuf:"bytes,2,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
	LigandId       string  `protobuf:"bytes,3,opt,name=ligand_id,json=ligandId,proto3" json:"ligand_id,omitempty"`
	LigandSMILES   string  `protobuf:"bytes,4,opt,name=ligand_smiles,json=ligandSmiles,proto3" json:"ligand_smiles,omitempty"`
	BindingScore   float64 `protobuf:"fixed64,5,opt,name=binding_score,json=bindingScore,proto3" json:"binding_score,omitempty"`
	RotatableBonds int32   `protobuf:"varint,6,opt,name=rotatable_bonds,json=rotatableBonds,proto3" json:"rotatable_bonds,omitempty"`
	Miner          string  `protobuf:"bytes,7,opt,name=miner,proto3" json:"miner,omitempty"`
	Reward         int64   `protobuf:"varint,8,opt,name=reward,proto3" json:"reward,omitempty"`
	IsHit          bool    `protobuf:"varint,9,opt,name=is_hit,json=isHit,proto3" json:"is_hit,omitempty"`
	BlockHeight    int64   `protobuf:"varint,10,opt,name=block_height,json=blockHeight,proto3" json:"block_height,omitempty"`
	SubmittedAt    int64   `protobuf:"varint,11,opt,name=submitted_at,json=submittedAt,proto3" json:"submitted_at,omitempty"`
}

func (m *DockingResult) Reset()         { *m = DockingResult{} }
func (m *DockingResult) String() string { return "DockingResult" }
func (m *DockingResult) ProtoMessage()  {}

// DockingClaim tracks which ligands a miner has claimed
type DockingClaim struct {
	Miner       string `protobuf:"bytes,1,opt,name=miner,proto3" json:"miner,omitempty"`
	JobId       string `protobuf:"bytes,2,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
	StartLigand int64  `protobuf:"varint,3,opt,name=start_ligand,json=startLigand,proto3" json:"start_ligand,omitempty"`
	EndLigand   int64  `protobuf:"varint,4,opt,name=end_ligand,json=endLigand,proto3" json:"end_ligand,omitempty"`
	ClaimedAt   int64  `protobuf:"varint,5,opt,name=claimed_at,json=claimedAt,proto3" json:"claimed_at,omitempty"`
}

func (m *DockingClaim) Reset()         { *m = DockingClaim{} }
func (m *DockingClaim) String() string { return "DockingClaim" }
func (m *DockingClaim) ProtoMessage()  {}

const (
	DockingJobStatusActive    = "active"
	DockingJobStatusCompleted = "completed"
	DockingJobStatusExpired   = "expired"
	DockingHitThreshold       = -7.0
	DockingBaseReward         = 1000
)

var BondMultipliers = map[int]float64{
	0: 0.5, 1: 0.6, 2: 0.7, 3: 0.8, 4: 0.9,
	5: 1.0, 6: 1.2, 7: 1.4, 8: 1.6, 9: 1.8,
	10: 2.0, 11: 2.3, 12: 2.6, 13: 3.0,
}

func GetBondMultiplier(bonds int) float64 {
	if bonds < 0 {
		bonds = 0
	}
	if bonds > 13 {
		bonds = 13
	}
	if m, ok := BondMultipliers[bonds]; ok {
		return m
	}
	return 1.0
}

func CalculateDockingReward(rotatableBonds int) int64 {
	return int64(float64(DockingBaseReward) * GetBondMultiplier(rotatableBonds))
}

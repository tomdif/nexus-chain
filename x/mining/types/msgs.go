package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	TypeMsgPostJob         = "post_job"
	TypeMsgSubmitProof     = "submit_proof"
	TypeMsgClaimRewards    = "claim_rewards"
	TypeMsgCancelJob       = "cancel_job"
	TypeMsgSubmitPublicJob = "submit_public_job"
)

// MsgPostJob - paid job submission with optional priority fee
type MsgPostJob struct {
	Customer    string    `protobuf:"bytes,1,opt,name=customer,proto3" json:"customer,omitempty"`
	ProblemType string    `protobuf:"bytes,2,opt,name=problem_type,json=problemType,proto3" json:"problem_type,omitempty"`
	ProblemData []byte    `protobuf:"bytes,3,opt,name=problem_data,json=problemData,proto3" json:"problem_data,omitempty"`
	ProblemHash string    `protobuf:"bytes,4,opt,name=problem_hash,json=problemHash,proto3" json:"problem_hash,omitempty"`
	Threshold   int64     `protobuf:"varint,5,opt,name=threshold,proto3" json:"threshold,omitempty"`
	Reward      sdk.Coins `protobuf:"bytes,6,rep,name=reward,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"reward"`
	PriorityFee sdk.Coins `protobuf:"bytes,7,rep,name=priority_fee,json=priorityFee,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"priority_fee"`
	Duration    int64     `protobuf:"varint,8,opt,name=duration,proto3" json:"duration,omitempty"`
	QuantumSafe bool      `protobuf:"varint,9,opt,name=quantum_safe,json=quantumSafe,proto3" json:"quantum_safe,omitempty"`
}

func (m *MsgPostJob) Reset()                  { *m = MsgPostJob{} }
func (m *MsgPostJob) String() string          { return "MsgPostJob" }
func (m *MsgPostJob) ProtoMessage()           {}
func (m *MsgPostJob) XXX_MessageName() string { return "nexus.mining.MsgPostJob" }

func (msg MsgPostJob) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Customer); err != nil {
		return ErrInvalidJob
	}
	return nil
}

func (msg MsgPostJob) GetSigners() []sdk.AccAddress {
	customer, _ := sdk.AccAddressFromBech32(msg.Customer)
	return []sdk.AccAddress{customer}
}

// MsgSubmitProof
type MsgSubmitProof struct {
	Miner        string `protobuf:"bytes,1,opt,name=miner,proto3" json:"miner,omitempty"`
	JobId        string `protobuf:"bytes,2,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
	Energy       int64  `protobuf:"varint,3,opt,name=energy,proto3" json:"energy,omitempty"`
	Proof        []byte `protobuf:"bytes,4,opt,name=proof,proto3" json:"proof,omitempty"`
	ProofType    string `protobuf:"bytes,5,opt,name=proof_type,json=proofType,proto3" json:"proof_type,omitempty"`
	SolutionHash string `protobuf:"bytes,6,opt,name=solution_hash,json=solutionHash,proto3" json:"solution_hash,omitempty"`
}

func (m *MsgSubmitProof) Reset()                  { *m = MsgSubmitProof{} }
func (m *MsgSubmitProof) String() string          { return "MsgSubmitProof" }
func (m *MsgSubmitProof) ProtoMessage()           {}
func (m *MsgSubmitProof) XXX_MessageName() string { return "nexus.mining.MsgSubmitProof" }

func (msg MsgSubmitProof) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Miner); err != nil {
		return ErrInvalidMiner
	}
	return nil
}

func (msg MsgSubmitProof) GetSigners() []sdk.AccAddress {
	miner, _ := sdk.AccAddressFromBech32(msg.Miner)
	return []sdk.AccAddress{miner}
}

// MsgClaimRewards
type MsgClaimRewards struct {
	Claimer string `protobuf:"bytes,1,opt,name=claimer,proto3" json:"claimer,omitempty"`
	JobId   string `protobuf:"bytes,2,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
}

func (m *MsgClaimRewards) Reset()                  { *m = MsgClaimRewards{} }
func (m *MsgClaimRewards) String() string          { return "MsgClaimRewards" }
func (m *MsgClaimRewards) ProtoMessage()           {}
func (m *MsgClaimRewards) XXX_MessageName() string { return "nexus.mining.MsgClaimRewards" }

func (msg MsgClaimRewards) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Claimer); err != nil {
		return ErrUnauthorized
	}
	return nil
}

func (msg MsgClaimRewards) GetSigners() []sdk.AccAddress {
	claimer, _ := sdk.AccAddressFromBech32(msg.Claimer)
	return []sdk.AccAddress{claimer}
}

// MsgCancelJob
type MsgCancelJob struct {
	Customer string `protobuf:"bytes,1,opt,name=customer,proto3" json:"customer,omitempty"`
	JobId    string `protobuf:"bytes,2,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
}

func (m *MsgCancelJob) Reset()                  { *m = MsgCancelJob{} }
func (m *MsgCancelJob) String() string          { return "MsgCancelJob" }
func (m *MsgCancelJob) ProtoMessage()           {}
func (m *MsgCancelJob) XXX_MessageName() string { return "nexus.mining.MsgCancelJob" }

func (msg MsgCancelJob) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Customer); err != nil {
		return ErrUnauthorized
	}
	return nil
}

func (msg MsgCancelJob) GetSigners() []sdk.AccAddress {
	customer, _ := sdk.AccAddressFromBech32(msg.Customer)
	return []sdk.AccAddress{customer}
}

// MsgSubmitPublicJob - free background job for public benefit
type MsgSubmitPublicJob struct {
	Submitter   string `protobuf:"bytes,1,opt,name=submitter,proto3" json:"submitter,omitempty"`
	Title       string `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty"`
	Description string `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	Category    string `protobuf:"bytes,4,opt,name=category,proto3" json:"category,omitempty"`
	ProblemData []byte `protobuf:"bytes,5,opt,name=problem_data,json=problemData,proto3" json:"problem_data,omitempty"`
	ProblemHash string `protobuf:"bytes,6,opt,name=problem_hash,json=problemHash,proto3" json:"problem_hash,omitempty"`
	Threshold   int64  `protobuf:"varint,7,opt,name=threshold,proto3" json:"threshold,omitempty"`
	IpfsCid     string `protobuf:"bytes,8,opt,name=ipfs_cid,json=ipfsCid,proto3" json:"ipfs_cid,omitempty"`
}

func (m *MsgSubmitPublicJob) Reset()                  { *m = MsgSubmitPublicJob{} }
func (m *MsgSubmitPublicJob) String() string          { return "MsgSubmitPublicJob" }
func (m *MsgSubmitPublicJob) ProtoMessage()           {}
func (m *MsgSubmitPublicJob) XXX_MessageName() string { return "nexus.mining.MsgSubmitPublicJob" }

func (msg MsgSubmitPublicJob) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Submitter); err != nil {
		return ErrUnauthorized
	}
	if len(msg.Title) == 0 || len(msg.Title) > 64 {
		return ErrInvalidJob
	}
	if len(msg.Category) == 0 || len(msg.Category) > 32 {
		return ErrInvalidJob
	}
	if len(msg.ProblemHash) == 0 {
		return ErrInvalidJob
	}
	return nil
}

func (msg MsgSubmitPublicJob) GetSigners() []sdk.AccAddress {
	submitter, _ := sdk.AccAddressFromBech32(msg.Submitter)
	return []sdk.AccAddress{submitter}
}

// MsgSubmitWork - collaborative mining work submission
// Proves: "I ran L steps of algorithm A from seed S, achieving energy E"
type MsgSubmitWork struct {
	Miner          string `protobuf:"bytes,1,opt,name=miner,proto3" json:"miner,omitempty"`
	JobId          string `protobuf:"bytes,2,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
	Epoch          uint64 `protobuf:"varint,3,opt,name=epoch,proto3" json:"epoch,omitempty"`
	NumSteps       uint64 `protobuf:"varint,4,opt,name=num_steps,json=numSteps,proto3" json:"num_steps,omitempty"`
	FinalEnergy    int64  `protobuf:"varint,5,opt,name=final_energy,json=finalEnergy,proto3" json:"final_energy,omitempty"`
	BestEnergy     int64  `protobuf:"varint,6,opt,name=best_energy,json=bestEnergy,proto3" json:"best_energy,omitempty"`
	BestConfigHash string `protobuf:"bytes,7,opt,name=best_config_hash,json=bestConfigHash,proto3" json:"best_config_hash,omitempty"`
	Proof          []byte `protobuf:"bytes,8,opt,name=proof,proto3" json:"proof,omitempty"`
	AlgorithmId    string `protobuf:"bytes,9,opt,name=algorithm_id,json=algorithmId,proto3" json:"algorithm_id,omitempty"`
}

func (m *MsgSubmitWork) Reset()                  { *m = MsgSubmitWork{} }
func (m *MsgSubmitWork) String() string          { return "MsgSubmitWork" }
func (m *MsgSubmitWork) ProtoMessage()           {}
func (m *MsgSubmitWork) XXX_MessageName() string { return "nexus.mining.MsgSubmitWork" }

func (msg MsgSubmitWork) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Miner); err != nil {
		return ErrInvalidMiner
	}
	if msg.NumSteps == 0 {
		return ErrInvalidProof
	}
	if len(msg.Proof) == 0 {
		return ErrInvalidProof
	}
	return nil
}

func (msg MsgSubmitWork) GetSigners() []sdk.AccAddress {
	miner, _ := sdk.AccAddressFromBech32(msg.Miner)
	return []sdk.AccAddress{miner}
}

// Response types
type MsgSubmitWorkResponse struct {
	Accepted    bool  `protobuf:"varint,1,opt,name=accepted,proto3" json:"accepted,omitempty"`
	WorkShares  int64 `protobuf:"varint,2,opt,name=work_shares,json=workShares,proto3" json:"work_shares,omitempty"`
	BonusShares int64 `protobuf:"varint,3,opt,name=bonus_shares,json=bonusShares,proto3" json:"bonus_shares,omitempty"`
}

func (m *MsgSubmitWorkResponse) Reset()         { *m = MsgSubmitWorkResponse{} }
func (m *MsgSubmitWorkResponse) String() string { return "MsgSubmitWorkResponse" }
func (m *MsgSubmitWorkResponse) ProtoMessage()  {}

// ============================================
// Molecular Docking Messages
// ============================================

// MsgSubmitDockingResult - submit a single ligand docking result
type MsgSubmitDockingResult struct {
	Miner          string  `protobuf:"bytes,1,opt,name=miner,proto3" json:"miner,omitempty"`
	JobId          string  `protobuf:"bytes,2,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
	LigandId       string  `protobuf:"bytes,3,opt,name=ligand_id,json=ligandId,proto3" json:"ligand_id,omitempty"`
	LigandSMILES   string  `protobuf:"bytes,4,opt,name=ligand_smiles,json=ligandSmiles,proto3" json:"ligand_smiles,omitempty"`
	BindingScore   float64 `protobuf:"fixed64,5,opt,name=binding_score,json=bindingScore,proto3" json:"binding_score,omitempty"`
	RotatableBonds int32   `protobuf:"varint,6,opt,name=rotatable_bonds,json=rotatableBonds,proto3" json:"rotatable_bonds,omitempty"`
	PoseData       []byte  `protobuf:"bytes,7,opt,name=pose_data,json=poseData,proto3" json:"pose_data,omitempty"`
}

func (m *MsgSubmitDockingResult) Reset()                  { *m = MsgSubmitDockingResult{} }
func (m *MsgSubmitDockingResult) String() string          { return "MsgSubmitDockingResult" }
func (m *MsgSubmitDockingResult) ProtoMessage()           {}
func (m *MsgSubmitDockingResult) XXX_MessageName() string { return "nexus.mining.MsgSubmitDockingResult" }

func (msg MsgSubmitDockingResult) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Miner); err != nil {
		return ErrInvalidMiner
	}
	if len(msg.JobId) == 0 {
		return ErrInvalidJob
	}
	if len(msg.LigandId) == 0 {
		return ErrInvalidJob
	}
	return nil
}

func (msg MsgSubmitDockingResult) GetSigners() []sdk.AccAddress {
	miner, _ := sdk.AccAddressFromBech32(msg.Miner)
	return []sdk.AccAddress{miner}
}

type MsgSubmitDockingResultResponse struct {
	Accepted bool  `protobuf:"varint,1,opt,name=accepted,proto3" json:"accepted,omitempty"`
	Reward   int64 `protobuf:"varint,2,opt,name=reward,proto3" json:"reward,omitempty"`
	IsHit    bool  `protobuf:"varint,3,opt,name=is_hit,json=isHit,proto3" json:"is_hit,omitempty"`
}

func (m *MsgSubmitDockingResultResponse) Reset()         { *m = MsgSubmitDockingResultResponse{} }
func (m *MsgSubmitDockingResultResponse) String() string { return "MsgSubmitDockingResultResponse" }
func (m *MsgSubmitDockingResultResponse) ProtoMessage()  {}

// MsgClaimDockingJob - miner claims a batch of ligands to dock
type MsgClaimDockingJob struct {
	Miner     string `protobuf:"bytes,1,opt,name=miner,proto3" json:"miner,omitempty"`
	JobId     string `protobuf:"bytes,2,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
	BatchSize int32  `protobuf:"varint,3,opt,name=batch_size,json=batchSize,proto3" json:"batch_size,omitempty"`
}

func (m *MsgClaimDockingJob) Reset()                  { *m = MsgClaimDockingJob{} }
func (m *MsgClaimDockingJob) String() string          { return "MsgClaimDockingJob" }
func (m *MsgClaimDockingJob) ProtoMessage()           {}
func (m *MsgClaimDockingJob) XXX_MessageName() string { return "nexus.mining.MsgClaimDockingJob" }

func (msg MsgClaimDockingJob) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Miner); err != nil {
		return ErrInvalidMiner
	}
	return nil
}

func (msg MsgClaimDockingJob) GetSigners() []sdk.AccAddress {
	miner, _ := sdk.AccAddressFromBech32(msg.Miner)
	return []sdk.AccAddress{miner}
}

type MsgClaimDockingJobResponse struct {
	JobId       string `protobuf:"bytes,1,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
	ProteinPDB  string `protobuf:"bytes,2,opt,name=protein_pdb,json=proteinPdb,proto3" json:"protein_pdb,omitempty"`
	TargetHash  string `protobuf:"bytes,3,opt,name=target_hash,json=targetHash,proto3" json:"target_hash,omitempty"`
	StartLigand int64  `protobuf:"varint,4,opt,name=start_ligand,json=startLigand,proto3" json:"start_ligand,omitempty"`
	EndLigand   int64  `protobuf:"varint,5,opt,name=end_ligand,json=endLigand,proto3" json:"end_ligand,omitempty"`
	CenterX     float64 `protobuf:"fixed64,6,opt,name=center_x,json=centerX,proto3" json:"center_x,omitempty"`
	CenterY     float64 `protobuf:"fixed64,7,opt,name=center_y,json=centerY,proto3" json:"center_y,omitempty"`
	CenterZ     float64 `protobuf:"fixed64,8,opt,name=center_z,json=centerZ,proto3" json:"center_z,omitempty"`
	SizeX       float64 `protobuf:"fixed64,9,opt,name=size_x,json=sizeX,proto3" json:"size_x,omitempty"`
	SizeY       float64 `protobuf:"fixed64,10,opt,name=size_y,json=sizeY,proto3" json:"size_y,omitempty"`
	SizeZ       float64 `protobuf:"fixed64,11,opt,name=size_z,json=sizeZ,proto3" json:"size_z,omitempty"`
}

func (m *MsgClaimDockingJobResponse) Reset()         { *m = MsgClaimDockingJobResponse{} }
func (m *MsgClaimDockingJobResponse) String() string { return "MsgClaimDockingJobResponse" }
func (m *MsgClaimDockingJobResponse) ProtoMessage()  {}

// MsgCreateDockingJob - create a new docking job (background or paid)
type MsgCreateDockingJob struct {
	Creator      string  `protobuf:"bytes,1,opt,name=creator,proto3" json:"creator,omitempty"`
	TargetHash   string  `protobuf:"bytes,2,opt,name=target_hash,json=targetHash,proto3" json:"target_hash,omitempty"`
	ProteinPDB   string  `protobuf:"bytes,3,opt,name=protein_pdb,json=proteinPdb,proto3" json:"protein_pdb,omitempty"`
	TotalLigands int64   `protobuf:"varint,4,opt,name=total_ligands,json=totalLigands,proto3" json:"total_ligands,omitempty"`
	CenterX      float64 `protobuf:"fixed64,5,opt,name=center_x,json=centerX,proto3" json:"center_x,omitempty"`
	CenterY      float64 `protobuf:"fixed64,6,opt,name=center_y,json=centerY,proto3" json:"center_y,omitempty"`
	CenterZ      float64 `protobuf:"fixed64,7,opt,name=center_z,json=centerZ,proto3" json:"center_z,omitempty"`
	IsBackground bool    `protobuf:"varint,8,opt,name=is_background,json=isBackground,proto3" json:"is_background,omitempty"`
	Reward       sdk.Coins `protobuf:"bytes,9,rep,name=reward,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"reward,omitempty"`
}

func (m *MsgCreateDockingJob) Reset()                  { *m = MsgCreateDockingJob{} }
func (m *MsgCreateDockingJob) String() string          { return "MsgCreateDockingJob" }
func (m *MsgCreateDockingJob) ProtoMessage()           {}
func (m *MsgCreateDockingJob) XXX_MessageName() string { return "nexus.mining.MsgCreateDockingJob" }

func (msg MsgCreateDockingJob) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return ErrUnauthorized
	}
	if len(msg.TargetHash) == 0 {
		return ErrInvalidJob
	}
	return nil
}

func (msg MsgCreateDockingJob) GetSigners() []sdk.AccAddress {
	creator, _ := sdk.AccAddressFromBech32(msg.Creator)
	return []sdk.AccAddress{creator}
}

type MsgCreateDockingJobResponse struct {
	JobId string `protobuf:"bytes,1,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
}

func (m *MsgCreateDockingJobResponse) Reset()         { *m = MsgCreateDockingJobResponse{} }
func (m *MsgCreateDockingJobResponse) String() string { return "MsgCreateDockingJobResponse" }
func (m *MsgCreateDockingJobResponse) ProtoMessage()  {}

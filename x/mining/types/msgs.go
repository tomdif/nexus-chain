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

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	TypeMsgPostJob      = "post_job"
	TypeMsgSubmitProof  = "submit_proof"
	TypeMsgClaimRewards = "claim_rewards"
	TypeMsgCancelJob    = "cancel_job"
)

// MsgPostJob
type MsgPostJob struct {
	Customer    string    `json:"customer"`
	ProblemType string    `json:"problem_type"`
	ProblemData []byte    `json:"problem_data"`
	ProblemHash string    `json:"problem_hash"`
	Threshold   int64     `json:"threshold"`
	Reward      sdk.Coins `json:"reward"`
	Duration    int64     `json:"duration"`
	QuantumSafe bool      `json:"quantum_safe"`
}

func (m *MsgPostJob) Reset()         { *m = MsgPostJob{} }
func (m *MsgPostJob) String() string { return "MsgPostJob" }
func (m *MsgPostJob) ProtoMessage()  {}
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
	Miner        string `json:"miner"`
	JobId        string `json:"job_id"`
	Energy       int64  `json:"energy"`
	Proof        []byte `json:"proof"`
	ProofType    string `json:"proof_type"`
	SolutionHash string `json:"solution_hash"`
}

func (m *MsgSubmitProof) Reset()         { *m = MsgSubmitProof{} }
func (m *MsgSubmitProof) String() string { return "MsgSubmitProof" }
func (m *MsgSubmitProof) ProtoMessage()  {}
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
	Claimer string `json:"claimer"`
	JobId   string `json:"job_id"`
}

func (m *MsgClaimRewards) Reset()         { *m = MsgClaimRewards{} }
func (m *MsgClaimRewards) String() string { return "MsgClaimRewards" }
func (m *MsgClaimRewards) ProtoMessage()  {}
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
	Customer string `json:"customer"`
	JobId    string `json:"job_id"`
}

func (m *MsgCancelJob) Reset()         { *m = MsgCancelJob{} }
func (m *MsgCancelJob) String() string { return "MsgCancelJob" }
func (m *MsgCancelJob) ProtoMessage()  {}
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

# NEXUS Architecture

## System Overview

NEXUS is built on Cosmos SDK v0.50 with CometBFT v0.38.9 for BFT consensus. The core innovation is replacing proof-of-work mining with proof-of-useful-work: solving Ising model optimization problems verified by Nova zero-knowledge proofs.

## Module Structure

### Mining Module (`x/mining/`)

The mining module handles all job management, proof verification, and reward distribution.

#### State

| Key Prefix | Description |
|------------|-------------|
| `job/` | Job records |
| `miner/` | Miner statistics |
| `checkpoint/` | Validator reward checkpoints |
| `params` | Module parameters |

#### Messages

| Message | Description |
|---------|-------------|
| `MsgPostJob` | Create paid optimization job |
| `MsgSubmitProof` | Submit ZK proof for job |
| `MsgClaimRewards` | Claim pending rewards |
| `MsgCancelJob` | Cancel queued job |
| `MsgSubmitPublicJob` | Submit free research job |

#### Keeper Methods

**Job Management:**
- `PostJob()` - Create job, burn fees, escrow rewards
- `SubmitProof()` - Verify proof, distribute rewards
- `CancelJob()` - Refund unstarted jobs

**Background Jobs:**
- `GenerateSyntheticBackgroundJob()` - Create Ising problem from block hash
- `ActivateRandomPublicJob()` - Pick random from public queue
- `ActivateNextPaidJob()` - Pop highest priority paid job
- `AdjustDifficulty()` - Scale problem size to 10-min target

**Emissions:**
- `ProcessEmissions()` - Calculate time-based rewards
- `GetCurrentEmissionRate()` - Get NEX/minute for current epoch
- `ClaimEmissionReward()` - Transfer emission bonus

**Checkpoints:**
- `createCheckpointAndDistribute()` - Every 300 blocks
- `DistributeValidatorRewards()` - Pro-rata by stake

## Consensus Flow
```
Block N                           Block N+1
   │                                  │
   ▼                                  ▼
┌─────────────────┐            ┌─────────────────┐
│  BeginBlocker   │            │  BeginBlocker   │
│  - Check jobs   │            │  - Check jobs   │
│  - Process emit │            │  - Process emit │
└────────┬────────┘            └────────┬────────┘
         │                              │
         ▼                              ▼
┌─────────────────┐            ┌─────────────────┐
│  DeliverTx      │            │  DeliverTx      │
│  - PostJob      │            │  - SubmitProof  │
│  - SubmitProof  │            │  - ClaimRewards │
└────────┬────────┘            └────────┬────────┘
         │                              │
         ▼                              ▼
┌─────────────────┐            ┌─────────────────┐
│  EndBlocker     │            │  EndBlocker     │
│  - Expire jobs  │            │  - Checkpoint?  │
│  - Difficulty   │            │  - Distribute   │
└─────────────────┘            └─────────────────┘
```

## Job Queue Priority
```
1. Paid Jobs (by priority fee, highest first)
        │
        ▼
2. Public Jobs (random selection)
        │
        ▼
3. Synthetic Jobs (auto-generated)
```

## ZK Proof Verification
```go
type VerifyRequest struct {
    JobID             string `json:"job_id"`
    ProblemCommitment string `json:"problem_commitment"`
    SpinCommitment    string `json:"spin_commitment"`
    ClaimedEnergy     int64  `json:"claimed_energy"`
    Threshold         int64  `json:"threshold"`
    Proof             string `json:"proof"`
}
```

The Nova verifier service validates:
1. Proof is well-formed
2. Energy calculation is correct
3. Energy meets threshold
4. Commitments match

## Fee Flow
```
Job Posted (1000 NEX)
    │
    ├── 2% Job Fee (20 NEX) ──────► BURNED
    │
    └── 98% Escrowed (980 NEX)
            │
            │ Job Solved
            ▼
    ┌───────┴───────┐
    │               │
    80%           20%
    │               │
    ▼               ▼
  Miner      Validator Pool
  (784 NEX)    (196 NEX)
                   │
                   │ Every 300 blocks
                   ▼
              Pro-rata to
              all validators
```

## Difficulty Adjustment

Target solve time: 10 minutes
```
If avg_solve_time < 8 min:
    problem_size *= 1.25  (harder)
    
If avg_solve_time > 12 min:
    problem_size *= 0.8   (easier)
    
Bounds: 64 ≤ size ≤ 2048 spins
```

## Files Reference

| File | Lines | Purpose |
|------|-------|---------|
| `keeper/msg_server.go` | 510 | Transaction handlers |
| `keeper/background_jobs.go` | 779 | Job queues, difficulty |
| `keeper/emissions.go` | 252 | 20-year schedule |
| `keeper/query_server.go` | 277 | Query handlers |
| `keeper/abci.go` | 194 | Block lifecycle |
| `types/msgs.go` | 154 | Message definitions |
| `types/types.go` | 83 | Core types |

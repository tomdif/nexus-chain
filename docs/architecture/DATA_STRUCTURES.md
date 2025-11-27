# NEXUS Data Structures

## Overview

NEXUS uses a three-tier data architecture:
- **On-chain**: Minimal data (hashes, summaries, proofs, rewards)
- **IPFS**: Full data (problems, configurations, results, landscapes)
- **API**: Query interface for labs and researchers

---

## 1. On-Chain Data Structures (Go)

### Core Job Types
```go
type JobType uint8

const (
    JobTypeIsingOptimization JobType = 1
    JobTypeProteinStability  JobType = 2  // ΔΔG predictions
    JobTypeMolecularDocking  JobType = 3  // Ligand binding
    JobTypeMutationalScan    JobType = 4  // Full landscape
)

type JobSource uint8
const (
    JobSourceBounty     JobSource = 1  // Paid by lab
    JobSourceBackground JobSource = 2  // Protocol emissions
)

type JobStatus uint8
const (
    JobStatusPending     JobStatus = 0
    JobStatusActive      JobStatus = 1
    JobStatusAggregating JobStatus = 2
    JobStatusCompleted   JobStatus = 3
    JobStatusExpired     JobStatus = 4
    JobStatusCancelled   JobStatus = 5
)
```

### Job Definition
```go
type Job struct {
    Id              string      `json:"id"`
    JobType         JobType     `json:"job_type"`
    Source          JobSource   `json:"source"`
    Customer        string      `json:"customer"`
    ProblemHash     string      `json:"problem_hash"`
    ProblemIpfsCid  string      `json:"problem_ipfs_cid"`
    ProteinId       string      `json:"protein_id"`
    PdbReference    string      `json:"pdb_reference"`
    Mutations       []string    `json:"mutations"`
    LigandSmiles    string      `json:"ligand_smiles"`
    LigandInchi     string      `json:"ligand_inchi"`
    Budget          sdk.Coins   `json:"budget"`
    MevFee          int64       `json:"mev_fee"`
    MinMiners       int64       `json:"min_miners"`
    MinSteps        int64       `json:"min_steps"`
    TargetCoverage  int64       `json:"target_coverage"`
    Status          JobStatus   `json:"status"`
    CreatedAt       int64       `json:"created_at"`
    StartedAt       int64       `json:"started_at"`
    CompletedAt     int64       `json:"completed_at"`
    Deadline        int64       `json:"deadline"`
    ResultsHash     string      `json:"results_hash"`
    ResultsIpfsCid  string      `json:"results_ipfs_cid"`
}
```

### Miner Work Submission
```go
type WorkSubmission struct {
    Id              string    `json:"id"`
    JobId           string    `json:"job_id"`
    Miner           string    `json:"miner"`
    Epoch           int64     `json:"epoch"`
    AssignedSeed    string    `json:"assigned_seed"`
    AssignedRegion  string    `json:"assigned_region"`
    AlgorithmId     string    `json:"algorithm_id"`
    StepsCompleted  int64     `json:"steps_completed"`
    BestEnergy      int64     `json:"best_energy"`
    BestConfigHash  string    `json:"best_config_hash"`
    ConfigIpfsCid   string    `json:"config_ipfs_cid"`
    DdgMicroKcal    int64     `json:"ddg_micro_kcal"`
    ProofHash       string    `json:"proof_hash"`
    ProofIpfsCid    string    `json:"proof_ipfs_cid"`
    ProofVerified   bool      `json:"proof_verified"`
    SubmittedAt     int64     `json:"submitted_at"`
    BlockHeight     int64     `json:"block_height"`
}
```

### Aggregated Results
```go
type JobResults struct {
    JobId               string        `json:"job_id"`
    TotalMiners         int64         `json:"total_miners"`
    TotalSteps          int64         `json:"total_steps"`
    TotalSubmissions    int64         `json:"total_submissions"`
    Results             []ResultEntry `json:"results"`
    ConvergenceScore    int64         `json:"convergence_score"`
    CoveragePercent     int64         `json:"coverage_percent"`
    FullResultsIpfsCid  string        `json:"full_results_ipfs_cid"`
    LandscapeIpfsCid    string        `json:"landscape_ipfs_cid"`
    FinalizedAt         int64         `json:"finalized_at"`
    FinalizedBlock      int64         `json:"finalized_block"`
}

type ResultEntry struct {
    TargetId            string    `json:"target_id"`
    Value               int64     `json:"value"`
    ValueUnit           string    `json:"value_unit"`
    StdDev              int64     `json:"std_dev"`
    Confidence          int64     `json:"confidence"`
    NumMiners           int64     `json:"num_miners"`
    AgreementPercent    int64     `json:"agreement_percent"`
    BestValue           int64     `json:"best_value"`
    BestMiner           string    `json:"best_miner"`
    BestConfigIpfsCid   string    `json:"best_config_ipfs_cid"`
    BestProofHash       string    `json:"best_proof_hash"`
}
```

### Reward Distribution
```go
type EpochRewards struct {
    Epoch               int64         `json:"epoch"`
    JobId               string        `json:"job_id"`
    TotalPool           int64         `json:"total_pool"`
    WorkPool            int64         `json:"work_pool"`
    ImprovementPool     int64         `json:"improvement_pool"`
    ValidatorPool       int64         `json:"validator_pool"`
    Distributions       []MinerReward `json:"distributions"`
}

type MinerReward struct {
    Miner               string    `json:"miner"`
    WorkReward          int64     `json:"work_reward"`
    ImprovementReward   int64     `json:"improvement_reward"`
    TotalReward         int64     `json:"total_reward"`
    StepsCompleted      int64     `json:"steps_completed"`
    BestEnergy          int64     `json:"best_energy"`
    WasGlobalBest       bool      `json:"was_global_best"`
}
```

---

## 2. IPFS Data Structures (JSON)

### Problem Definition - Protein Stability
```json
{
  "type": "protein_stability",
  "version": "1.0",
  "protein": {
    "uniprot_id": "P00533",
    "gene_name": "EGFR",
    "organism": "Homo sapiens",
    "sequence": "MRPSGTAGAALLALLAALCPASRALEEK...",
    "length": 1210
  },
  "structure": {
    "pdb_id": "4HJO",
    "chain": "A",
    "resolution_angstrom": 2.8,
    "structure_ipfs_cid": "Qm..."
  },
  "mutations_to_compute": [
    {"position": 858, "wild_type": "L", "mutant": "R"},
    {"position": 790, "wild_type": "T", "mutant": "M"}
  ],
  "algorithm_params": {
    "algorithm": "nexus_sa_v1",
    "steps_per_miner": 100000,
    "temperature_schedule": "exponential",
    "scoring_function": "rosetta_ref15"
  }
}
```

### Aggregated Results
```json
{
  "job_id": "job_2025_12345",
  "job_type": "protein_stability",
  "completed_at": "2025-11-27T15:30:00Z",
  "summary": {
    "protein_id": "P00533",
    "pdb_reference": "4HJO",
    "total_miners": 847,
    "total_compute_steps": 84700000,
    "coverage_percent": 99.2
  },
  "results": [
    {
      "mutation": "L858R",
      "hgvs_notation": "p.Leu858Arg",
      "position": 858,
      "wild_type": "L",
      "mutant": "R",
      "ddg_kcal_mol": -2.31,
      "ddg_std": 0.15,
      "confidence": 0.95,
      "classification": "stabilizing",
      "statistics": {
        "num_miners": 284,
        "agreement_percent": 94.2
      }
    }
  ],
  "verification": {
    "chain": "nexus-mainnet",
    "finalized_block": 1250500,
    "all_proofs_verified": true
  }
}
```

### Full Landscape Data
```json
{
  "landscape_id": "landscape_P00533_v3",
  "protein_id": "P00533",
  "gene_name": "EGFR",
  "coverage": {
    "positions_total": 1210,
    "positions_computed": 1185,
    "mutations_computed": 21847,
    "completeness_percent": 95.0
  },
  "matrix": {
    "format": "position_by_aminoacid",
    "amino_acids": ["A","C","D","E","F","G","H","I","K","L","M","N","P","Q","R","S","T","V","W","Y"],
    "values": "[[0.2, -1.5, ...], ...]",
    "uncertainties": "[[0.1, 0.2, ...], ...]"
  },
  "export_formats": {
    "csv_cid": "Qm...",
    "parquet_cid": "Qm...",
    "heatmap_png_cid": "Qm..."
  }
}
```

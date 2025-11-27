# NEXUS Collaborative Mining Architecture

## Overview

NEXUS uses a collaborative mining model where all miners work together on jobs rather than competing individually. This maximizes network efficiency and builds comprehensive scientific datasets.

---

## Dual Queue System

### Queue Priority

1. **Bounty Queue** (always first)
   - Labs post jobs with budget + MEV fee
   - Higher MEV = processed sooner
   - ALL miners work on top bounty job
   - Each miner gets unique assigned seed

2. **Background Queue** (fills gaps)
   - Auto-generated public good jobs
   - Funded by protocol emissions
   - Only when bounty queue empty

---

## Miner Assignment

Each miner gets a deterministic, unique starting point:
```
seed = Hash(job_id || epoch || miner_address || VRF_randomness)
```

This ensures:
- No two miners explore the same region
- Assignment is verifiable
- No coordination needed

---

## Proof Requirements

The Nova circuit must verify:
1. Initial state derived correctly from seed
2. Each step follows the canonical algorithm
3. Final energy computed correctly

Proves: "Starting from seed S, I ran algorithm A for L steps, ending at config C with energy E"

---

## Reward Distribution

### Pool Allocation

| Pool | Percentage | Description |
|------|------------|-------------|
| Work Pool | 70% | Pro-rata by verified steps |
| Improvement Pool | 20% | Bonus for best results |
| Validator Pool | 10% | To staking validators |

### Work Rewards
```
work_reward[miner] = work_pool * (miner_steps / total_steps)
```

### Improvement Rewards

Distributed based on:
- Finding the global best energy
- Improving on previous best
- Discovering new local minima

---

## Algorithm: nexus_sa_v1

All miners run identical simulated annealing:
```
Parameters:
- T_start: 10.0
- T_end: 0.01
- Decay: 0.99995
- Steps: 100,000

For each step:
1. Select position from PRG(seed, step)
2. Propose new value
3. Calculate energy change
4. Accept if improves or by Metropolis criterion
5. Update temperature
```

---

## Work Unit Sizing

| Problem Type | Steps | Total Time |
|--------------|-------|------------|
| Ising (64 spins) | 100,000 | ~10s |
| Ising (1024 spins) | 100,000 | ~12s |
| Protein stability | 10,000 | ~70s |

---

## Solution Landscape

Results aggregated into comprehensive datasets:
```
| Mutation | DDG | StdDev | Confidence | Miners |
|----------|-----|--------|------------|--------|
| L858R | -2.31 | 0.15 | 0.95 | 284 |
| T790M | +1.72 | 0.08 | 0.98 | 312 |
```

Stored on IPFS with on-chain verification hashes.

---

## Implementation Phases

### Phase 1: Collaborative Ising
- Assigned seeds per miner
- Canonical simulated annealing
- Work-based rewards

### Phase 2: Lattice Proteins (3-6 months)
- Simplified protein model
- Same circuit structure

### Phase 3: Real Proteins (6-12 months)
- Coarse-grained model
- UniProt/AlphaFold integration
- Human proteome database

# NEXUS Network

**Bitcoin, but useful.** A proof-of-useful-work blockchain where validators solve real optimization problems instead of wasteful hash puzzles.

## Overview

NEXUS replaces Bitcoin's SHA-256 mining with Ising model optimization - the same mathematical framework used for protein folding, drug discovery, and logistics optimization. Miners submit zero-knowledge proofs (Nova/Halo2) that verify their computational work without revealing solutions.

## Key Features

- **Proof of Useful Work**: Solve real-world optimization problems
- **Zero-Knowledge Proofs**: Nova recursive SNARKs verify work in ~10 seconds
- **20-Year Emission Schedule**: 75B NEX distributed to miners
- **Dual Fee Burn**: 2% job fees + 50% transaction fees burned
- **80/20 Split**: Miners get 80%, validators get 20%
- **2-Second Finality**: CometBFT consensus with instant finality

## Quick Start

### Prerequisites

- Go 1.22+
- jq

### Build
```bash
git clone https://github.com/tomdif/nexus-chain.git
cd nexus-chain
go build -o nexusd ./cmd/nexusd/
```

### Run Single Node
```bash
./nexusd init my-validator
./nexusd start
```

### Run 3-Node Testnet
```bash
./scripts/testnet.sh      # Initialize
./scripts/start-testnet.sh # Start all nodes
./scripts/stop-testnet.sh  # Stop all nodes
```

## CLI Commands

### Transactions
```bash
# Post a paid optimization job
nexusd tx mining post-job <problem-hash> <threshold> <reward> --priority-fee 100000

# Submit proof for a job
nexusd tx mining submit-proof <job-id> <solution-hash> <energy> <proof-base64>

# Submit free public research job
nexusd tx mining submit-public-job <title> <category> <hash> <threshold> <ipfs-cid>

# Claim mining rewards
nexusd tx mining claim-rewards <job-id>

# Cancel job and get refund
nexusd tx mining cancel-job <job-id>
```

### Queries
```bash
nexusd query mining get-job <job-id>
nexusd query mining list-jobs
nexusd query mining get-miner <address>
nexusd query mining get-params
nexusd query mining get-active-job
nexusd query mining get-queue-status
nexusd query mining get-emission-info
```

## Architecture
```
┌─────────────────────────────────────────────────────────────┐
│                      NEXUS Network                          │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │   Mining    │  │   Staking   │  │   Bank (NEX Token)  │  │
│  │   Module    │  │   Module    │  │      Module         │  │
│  └──────┬──────┘  └──────┬──────┘  └──────────┬──────────┘  │
│         │                │                     │             │
│  ┌──────▼────────────────▼─────────────────────▼──────────┐ │
│  │                    Cosmos SDK v0.50                    │ │
│  └────────────────────────┬───────────────────────────────┘ │
│                           │                                  │
│  ┌────────────────────────▼───────────────────────────────┐ │
│  │                  CometBFT v0.38.9                      │ │
│  │            (Byzantine Fault Tolerant Consensus)        │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Tokenomics

| Parameter | Value |
|-----------|-------|
| Token | NEX |
| Denomination | unexus (1 NEX = 1,000,000 unexus) |
| Total Supply | 100B NEX |
| Initial Allocation | 25B NEX (founders/treasury) |
| Mining Emissions | 75B NEX over 20+ years |
| Job Fee Burn | 2% |
| Transaction Fee Burn | 50% |

### Emission Schedule

| Years | Rate (NEX/min) | % of Max |
|-------|----------------|----------|
| 1-2 | 35,950 | 100% |
| 3-4 | 17,975 | 50% |
| 5-6 | 8,988 | 25% |
| 7-8 | 4,494 | 12.5% |
| 9-10 | 2,229 | 6.2% |
| 11-12 | 1,114 | 3.1% |
| 13+ | 539 | Perpetual |

## Job System

### Job Types

1. **Paid Jobs**: Customers pay NEX for priority computation
   - Priority fee determines queue position
   - Full reward + emission bonus to solver

2. **Public Jobs**: Free research problems (requires stake)
   - Random selection from queue
   - Emission rewards only

3. **Synthetic Jobs**: Auto-generated when queues empty
   - Ising problems from block hash
   - Keeps miners productive

### Job Flow
```
Customer → Post Job → Priority Queue → Active Job → Miner Solves
                ↓                            ↓
         2% Fee Burned              ZK Proof Verified
                                           ↓
                                    Rewards Distributed
                                    (80% Miner, 20% Validators)
```

## ZK Proof Integration

NEXUS uses Nova recursive SNARKs to verify optimization work:
```
Miner                          NEXUS Chain
  │                                 │
  │  Solve Ising Problem            │
  │  Generate Nova Proof            │
  │                                 │
  │──── Submit Proof ──────────────►│
  │                                 │
  │     POST /verify                │
  │     ┌──────────────────────┐    │
  │     │ Nova Verifier Service│    │
  │     │ (localhost:3000)     │    │
  │     └──────────────────────┘    │
  │                                 │
  │◄─── Rewards if Valid ──────────│
```

## Network Ports

| Node | P2P Port | RPC Port |
|------|----------|----------|
| Node 1 | 26656 | 26657 |
| Node 2 | 26666 | 26667 |
| Node 3 | 26676 | 26677 |

## Development

### Run Tests
```bash
go test ./x/mining/keeper/...
```

### Project Structure
```
nexus-chain/
├── app/                    # Application wiring
│   ├── app.go             # Main app definition
│   └── ante/              # Transaction preprocessing
├── cmd/nexusd/            # CLI entry point
├── scripts/               # Testnet scripts
├── x/mining/              # Mining module
│   ├── client/cli/        # CLI commands
│   ├── keeper/            # State management
│   │   ├── msg_server.go  # Transaction handlers
│   │   ├── query_server.go# Query handlers
│   │   ├── emissions.go   # Emission schedule
│   │   └── background_jobs.go # Job queues
│   └── types/             # Type definitions
└── docs/                  # Documentation
```

## License

MIT

## Links

- [GitHub](https://github.com/tomdif/nexus-chain)
- [ZK Proof System](https://github.com/tomdif/zkproof-system)

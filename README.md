# NEXUS Chain

A Cosmos SDK blockchain with Proof-of-Useful-Work mining.

> **Note:** This is a development scaffold. Chain name, token name, and economics are all placeholders that can be changed before mainnet.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    NEXUS Network                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Fast Layer (PoS)              Slow Layer (PoUW)           │
│  ├─ 2-second blocks            ├─ 10-minute checkpoints    │
│  ├─ BFT validators             ├─ Miners solve problems    │
│  ├─ Instant finality           ├─ ZK proof verification    │
│  └─ Handles transactions       └─ Deep finality            │
│                                                             │
│  x/mining module:                                           │
│  ├─ Job posting (MsgPostJob)                               │
│  ├─ Proof submission (MsgSubmitProof)                      │
│  ├─ Universal share formula                                │
│  ├─ Checkpoint management                                  │
│  └─ Validator mining requirements                          │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Prerequisites

- Go 1.22+
- Make
- Git

## Quick Start

```bash
# Clone
git clone https://github.com/nexus-chain/nexus
cd nexus

# Build
make build

# Initialize local testnet
chmod +x scripts/init.sh
./scripts/init.sh

# Start the chain
./build/nexusd start
```

## Project Structure

```
nexus-chain/
├── app/                    # Main application
├── cmd/nexusd/            # Binary entry point
├── x/mining/              # Mining module (core logic)
│   ├── keeper/            # State management
│   │   ├── keeper.go      # Core business logic
│   │   ├── msg_server.go  # Transaction handlers
│   │   └── abci.go        # Block-level hooks
│   ├── types/             # Data structures
│   │   ├── types.go       # Core types (Job, Proof, etc.)
│   │   ├── msgs.go        # Transaction messages
│   │   ├── params.go      # Configurable parameters
│   │   ├── errors.go      # Error definitions
│   │   └── genesis.go     # Genesis state
│   └── module.go          # AppModule interface
├── scripts/               # Helper scripts
├── Makefile              # Build commands
└── go.mod                # Dependencies
```

## Mining Module

### Core Types

- **Job**: Optimization problem posted by customers
- **MiningProof**: ZK proof submission from miners
- **Checkpoint**: 10-minute mining checkpoint
- **ValidatorMiningRecord**: Tracks validator's mining activity

### Messages (Transactions)

| Message | Description | Who |
|---------|-------------|-----|
| `MsgPostJob` | Post optimization problem with reward | Customer |
| `MsgSubmitProof` | Submit solution with ZK proof | Miner |
| `MsgClaimRewards` | Claim earned rewards | Miner |
| `MsgCancelJob` | Cancel job (if no solutions) | Customer |

### Universal Share Formula

```
shares = max(0, previous_best_objective − your_objective)
```

- Hardware-agnostic (GPU, CPU, quantum all use same formula)
- Outcome-based (rewards improvement, not effort)
- Ungameable (must actually improve solution)

### Parameters (Configurable)

| Parameter | Default | Description |
|-----------|---------|-------------|
| `MinerSharePercent` | 80% | Portion of rewards to miners |
| `ValidatorSharePercent` | 20% | Portion of rewards to validators |
| `CheckpointInterval` | 300 blocks | ~10 min at 2s blocks |
| `MinProofPeriod` | 1 week | Validators must mine this often |
| `JobFeeBurnPercent` | 2% | Burned from job rewards |
| `TxFeeBurnPercent` | 50% | Burned from tx fees |

## Development

### Build Commands

```bash
make build          # Build binary
make install        # Install to GOPATH
make test           # Run tests
make lint           # Run linter
make proto-gen      # Generate protobuf
```

### Testing

```bash
# Unit tests
go test ./x/mining/...

# Integration test (requires running chain)
./scripts/test_mining.sh
```

## Next Steps

### 1. ZK Verifier Integration

The `VerifyProof` function in `keeper.go` is a placeholder. To integrate your Nova prover:

**Option A: FFI Bridge (Recommended)**
```go
// #cgo LDFLAGS: -L./lib -lnova_verifier
// #include "nova_verifier.h"
import "C"

func (k Keeper) VerifyProof(...) bool {
    result := C.verify_nova_proof(
        C.CBytes(problemHash),
        C.int64_t(energy),
        C.CBytes(proof),
    )
    return result == 1
}
```

**Option B: Subprocess**
```go
func (k Keeper) VerifyProof(...) bool {
    cmd := exec.Command("nova-verifier", "--proof", base64.StdEncoding.EncodeToString(proof))
    return cmd.Run() == nil
}
```

### 2. Complete the App Wiring

The `app/app.go` needs to:
- Register the mining module
- Wire up keeper dependencies
- Add to module manager

### 3. Add CLI Commands

Create CLI commands in `x/mining/client/cli/` for:
- `nexusd tx mining post-job`
- `nexusd tx mining submit-proof`
- `nexusd query mining job`
- etc.

### 4. Add Protobuf Definitions

Generate proper protobuf files in `proto/mining/v1/` for:
- Messages
- Queries
- Genesis state

## Configuration

### Placeholder Values (Change Before Mainnet)

| Item | Current | Notes |
|------|---------|-------|
| Chain ID | `nexus-testnet-1` | Change to final name |
| Binary | `nexusd` | Change to final name |
| Denom | `unexus` | Change to final token |
| Max Supply | 100M | Adjust per tokenomics |
| Emission Rate | 1M/checkpoint | Adjust per economics |

All of these are JSON/config values, not hardcoded in Go.

## License

[Your license here]

## Contact

- GitHub: github.com/tomdif/ising-verifier-final
- Email: tomdif@gmail.com

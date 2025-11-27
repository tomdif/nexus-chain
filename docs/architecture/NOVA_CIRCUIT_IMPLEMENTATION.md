# Nova Circuit Implementation Notes

## Build Requirements

The nova-prover requires either:
1. CUDA toolkit with gcc <= 12, OR
2. Modified dependencies to disable GPU acceleration

### Option 1: Fix CUDA Build (Recommended for Production)
```bash
# Install gcc-12 alongside newer gcc
sudo apt install gcc-12 g++-12

# Set environment for CUDA builds
export CC=gcc-12
export CXX=g++-12
```

### Option 2: Disable GPU (For Testing)

Modify `nova-prover/Cargo.toml` to use CPU-only dependencies.
This requires patching nova-snark to not use pasta-msm with CUDA.

---

## Collaborative Mining Circuit Changes

### Files Modified

1. `nova-prover/src/collaborative.rs` - New module with:
   - `derive_seed()` - Deterministic seed from job + miner + epoch
   - `derive_initial_config()` - Initial spins from seed
   - `PrgState` - Poseidon-based PRG for SA steps
   - `sa_step()` - Single simulated annealing step
   - `CollaborativeMiningCircuit` - Nova step circuit
   - `CollaborativeProver` - Prover that generates all fold circuits

2. `nova-prover/src/lib.rs` - Add `pub mod collaborative;`

### Circuit State (arity = 11)
```
z[0]  = step_count        - Number of SA steps completed
z[1]  = running_energy    - Current energy (biased)
z[2]  = best_energy       - Best energy found (biased)
z[3]  = temperature       - Current temperature (fixed-point)
z[4]  = prg_state         - PRG state for determinism
z[5]  = problem_commitment - Hash of Ising problem
z[6]  = seed              - Derived seed
z[7]  = job_hash          - Job identifier
z[8]  = epoch             - Epoch number
z[9]  = miner_address     - Miner's address
z[10] = vrf_randomness    - Network randomness
```

### Proof Flow
```
1. Miner receives job assignment
2. Derives seed = Poseidon(job_hash, epoch, miner_address, vrf_randomness)
3. Derives initial config from seed
4. Runs L steps of SA, generating fold circuits
5. Nova folds all circuits into single proof
6. Submits proof with:
   - job_id, epoch, miner_address
   - num_steps, final_energy, best_energy
   - Nova proof bytes
```

### Verification

On-chain verifier:
1. Recomputes expected seed from public inputs
2. Verifies Nova proof with seed as public input
3. Confirms num_steps matches claimed work
4. Records submission for reward distribution

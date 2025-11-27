# Nova Circuit Design for Collaborative Mining

## Current Circuit (Competitive Model)

The existing circuit proves:
```
Public Inputs:
  - problem_hash: Hash of the Ising problem
  - threshold: Energy threshold to beat

Private Inputs (Witness):
  - spins: Configuration [+1/-1 for each spin]
  - couplings: Edge weights

Circuit Constraints:
  1. Each spin is binary (+1 or -1)
  2. Energy = -Σ J_ij × s_i × s_j (computed correctly)
  3. Energy ≤ threshold
```

**Problem:** This only proves the result, not the work done to find it.

---

## New Circuit (Collaborative Model)

The new circuit must prove:
```
Public Inputs:
  - job_id: Unique job identifier
  - miner_address: Miner's address
  - epoch: Current epoch
  - vrf_randomness: Network randomness
  - algorithm_id: Which algorithm version (e.g., "nexus_sa_v1")
  - num_steps: Number of steps claimed (L)
  - final_energy: Energy of final configuration

Private Inputs (Witness):
  - problem_data: Full Ising problem (couplings)
  - initial_config: Starting configuration (derived from seed)
  - trajectory: All intermediate states OR compressed representation
  - final_config: Ending configuration

Circuit Constraints:
  1. seed = Hash(job_id || epoch || miner_address || vrf_randomness)
  2. initial_config = DeriveConfig(seed, problem_size)
  3. For step i in 0..L:
     - next_state = SimulatedAnnealingStep(current_state, seed, i)
  4. final_config matches after L steps
  5. final_energy computed correctly from final_config
```

---

## Nova's Incremental Approach

Nova is perfect for this because it handles **repeated computation** efficiently.

### Step Circuit (F)

Define a single step that gets folded L times:
```rust
struct StepCircuit {
    // Public inputs (same across all steps)
    job_hash: Field,
    seed: Field,
    algorithm_params: AlgorithmParams,
    
    // State that evolves
    step_number: u64,
    current_config: Vec<Field>,  // Spin values
    current_energy: Field,
    best_energy: Field,
    best_config_hash: Field,
    prg_state: Field,  // Pseudo-random generator state
    temperature: Field,
}

impl StepCircuit {
    fn synthesize(&self, cs: &mut ConstraintSystem) -> Result<Self> {
        // 1. Derive random values for this step
        let (position, flip_random, accept_random, new_prg_state) = 
            self.advance_prg(cs)?;
        
        // 2. Calculate energy delta if we flip spin at 'position'
        let delta_e = self.calculate_flip_delta(cs, position)?;
        
        // 3. Metropolis acceptance criterion
        let accept = self.metropolis_accept(cs, delta_e, accept_random)?;
        
        // 4. Conditionally update configuration
        let new_config = self.conditional_flip(cs, position, accept)?;
        
        // 5. Update energy
        let new_energy = self.current_energy + select(accept, delta_e, 0);
        
        // 6. Update best if improved
        let is_better = new_energy < self.best_energy;
        let new_best_energy = select(is_better, new_energy, self.best_energy);
        let new_best_hash = select(is_better, 
            hash(new_config), self.best_config_hash);
        
        // 7. Cool temperature
        let new_temperature = self.temperature * DECAY_FACTOR;
        
        Ok(StepCircuit {
            step_number: self.step_number + 1,
            current_config: new_config,
            current_energy: new_energy,
            best_energy: new_best_energy,
            best_config_hash: new_best_hash,
            prg_state: new_prg_state,
            temperature: new_temperature,
            ..self
        })
    }
}
```

### Initial State Derivation
```rust
fn derive_initial_state(
    job_hash: Field,
    epoch: u64,
    miner_address: Field,
    vrf_randomness: Field,
    problem_size: usize,
) -> StepCircuit {
    // Deterministic seed from public inputs
    let seed = poseidon_hash([
        job_hash,
        Field::from(epoch),
        miner_address,
        vrf_randomness,
    ]);
    
    // Derive initial configuration from seed
    let mut prg_state = seed;
    let mut initial_config = Vec::with_capacity(problem_size);
    
    for i in 0..problem_size {
        let (bit, new_state) = prg_next_bit(prg_state);
        initial_config.push(if bit { 1 } else { -1 });
        prg_state = new_state;
    }
    
    // Calculate initial energy
    let initial_energy = calculate_energy(&initial_config, &couplings);
    
    StepCircuit {
        job_hash,
        seed,
        step_number: 0,
        current_config: initial_config,
        current_energy: initial_energy,
        best_energy: initial_energy,
        best_config_hash: hash(&initial_config),
        prg_state,
        temperature: INITIAL_TEMPERATURE,
        algorithm_params: default_sa_params(),
    }
}
```

---

## Key Circuit Components

### 1. Pseudo-Random Generator (PRG)

Must be ZK-friendly (use Poseidon hash):
```rust
fn prg_advance(state: Field, step: u64) -> (Field, Field, Field, Field) {
    // Generate multiple random values from state
    let new_state = poseidon_hash([state, Field::from(step)]);
    
    let position_rand = poseidon_hash([new_state, Field::from(0)]);
    let flip_rand = poseidon_hash([new_state, Field::from(1)]);
    let accept_rand = poseidon_hash([new_state, Field::from(2)]);
    
    (position_rand, flip_rand, accept_rand, new_state)
}

fn prg_to_index(rand: Field, max: usize) -> usize {
    // Convert field element to index in range [0, max)
    rand.to_bits()[..log2(max)].to_usize() % max
}
```

### 2. Energy Delta Calculation

Only need to compute change from flipping one spin:
```rust
fn calculate_flip_delta(
    cs: &mut ConstraintSystem,
    config: &[Field],
    couplings: &[(usize, usize, Field)],  // (i, j, J_ij)
    flip_position: usize,
) -> Field {
    // ΔE = 2 × s_i × Σ_j J_ij × s_j
    // (flipping s_i changes sign, so delta is 2× the interaction)
    
    let s_i = config[flip_position];
    let mut delta = Field::zero();
    
    for (i, j, j_ij) in couplings {
        if *i == flip_position {
            delta = delta + j_ij * config[*j];
        } else if *j == flip_position {
            delta = delta + j_ij * config[*i];
        }
    }
    
    // Multiply by 2 * s_i (flipping changes sign)
    delta * Field::from(2) * s_i
}
```

### 3. Metropolis Acceptance
```rust
fn metropolis_accept(
    cs: &mut ConstraintSystem,
    delta_e: Field,
    temperature: Field,
    random: Field,
) -> Boolean {
    // Accept if:
    // 1. delta_e < 0 (energy improves), OR
    // 2. random < exp(-delta_e / temperature)
    
    let improves = delta_e.is_negative();
    
    // For exp(-delta_e / T), we use lookup table or approximation
    // Since we're in a finite field, we compare:
    // random * T < threshold_for_delta_e
    
    let threshold = exp_lookup(delta_e, temperature);
    let accept_by_random = random.lt(threshold);
    
    improves.or(accept_by_random)
}
```

### 4. Conditional Configuration Update
```rust
fn conditional_flip(
    cs: &mut ConstraintSystem,
    config: &[Field],
    position: usize,
    accept: Boolean,
) -> Vec<Field> {
    let mut new_config = config.to_vec();
    
    // new_spin = accept ? -old_spin : old_spin
    new_config[position] = select(
        accept,
        config[position].neg(),
        config[position]
    );
    
    new_config
}
```

---

## Nova Folding Structure
```
Initial State (z_0):
  - Derived from seed
  - step=0, initial_config, initial_energy

After 1 fold (z_1 = F(z_0)):
  - step=1, config after 1 SA step

After L folds (z_L = F^L(z_0)):
  - step=L, final_config, best_energy_found

Final Proof:
  - Proves z_L = F^L(z_0)
  - Public inputs: job_hash, seed, L, final_energy, best_config_hash
  - Verifier checks in O(1) time regardless of L
```

---

## Proof Structure
```rust
struct CollaborativeProof {
    // Nova IVC proof
    nova_proof: NovaProof,
    
    // Public inputs
    pub job_id: String,
    pub epoch: u64,
    pub miner_address: String,
    pub vrf_randomness: [u8; 32],
    pub num_steps: u64,
    pub final_energy: i64,
    pub best_energy: i64,
    pub best_config_hash: [u8; 32],
    
    // For verification
    pub algorithm_version: String,
}

impl CollaborativeProof {
    fn verify(&self, problem_hash: &[u8; 32]) -> bool {
        // 1. Recompute seed from public inputs
        let expected_seed = hash(
            self.job_id,
            self.epoch,
            self.miner_address,
            self.vrf_randomness
        );
        
        // 2. Verify Nova proof
        let public_inputs = vec![
            problem_hash,
            expected_seed,
            self.num_steps,
            self.final_energy,
            self.best_energy,
            self.best_config_hash,
        ];
        
        self.nova_proof.verify(public_inputs)
    }
}
```

---

## On-Chain Verification
```go
// In x/mining/keeper/msg_server.go

func (k Keeper) SubmitWork(ctx context.Context, msg *MsgSubmitWork) (*MsgSubmitWorkResponse, error) {
    // 1. Verify miner is registered
    // 2. Get current job and epoch
    // 3. Recompute expected seed
    expectedSeed := sha256(msg.JobId + msg.Epoch + msg.Miner + currentVRF)
    
    // 4. Verify Nova proof
    publicInputs := NovaPublicInputs{
        ProblemHash: job.ProblemHash,
        Seed: expectedSeed,
        NumSteps: msg.NumSteps,
        FinalEnergy: msg.FinalEnergy,
        BestEnergy: msg.BestEnergy,
        BestConfigHash: msg.BestConfigHash,
    }
    
    valid := k.novaVerifier.Verify(msg.Proof, publicInputs)
    if !valid {
        return nil, ErrInvalidProof
    }
    
    // 5. Record work submission
    submission := WorkSubmission{
        JobId: msg.JobId,
        Miner: msg.Miner,
        Epoch: msg.Epoch,
        StepsCompleted: msg.NumSteps,
        BestEnergy: msg.BestEnergy,
        ProofHash: sha256(msg.Proof),
        Verified: true,
    }
    k.SaveSubmission(ctx, submission)
    
    // 6. Update job aggregation
    k.UpdateJobResults(ctx, msg.JobId, submission)
    
    return &MsgSubmitWorkResponse{}, nil
}
```

---

## Circuit Size Estimates

| Component | Constraints per Step |
|-----------|---------------------|
| PRG advance (Poseidon) | ~300 |
| Position selection | ~50 |
| Energy delta calc | ~200 (depends on connectivity) |
| Metropolis decision | ~100 |
| Conditional flip | ~50 |
| Best update | ~100 |
| **Total per step** | **~800** |

For 100,000 steps:
- Without Nova: 80M constraints (infeasible)
- With Nova: ~800 constraints + O(log L) recursion overhead

**Nova makes this tractable!**

---

## Migration Path

### Phase 1: Add seed verification
- Circuit proves initial state derived from seed
- Rest of circuit unchanged (just proves final energy)

### Phase 2: Add step counting
- Circuit proves L steps executed
- But doesn't verify each step (trust miner did SA)

### Phase 3: Full trajectory verification
- Circuit verifies every step of canonical algorithm
- Maximum security, verifiable work

---

## Alternative: Checkpoint Proofs

If full trajectory is too expensive, use checkpoints:
```
Instead of proving: z_L = F^L(z_0)

Prove: 
  z_1000 = F^1000(z_0)      [checkpoint 1]
  z_2000 = F^1000(z_1000)   [checkpoint 2]
  ...
  z_L = F^(L mod 1000)(z_{L-1000})

Submit checkpoints as separate proofs.
Chain verifies they chain together.
```

This lets miners submit partial work and get paid incrementally.

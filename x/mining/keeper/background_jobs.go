package keeper

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"nexus/x/mining/types"
)

// Background job constants
const (
	BackgroundJobCustomer        = "nexus_network"
	DefaultBackgroundJobDuration = 24 * 60 * 60 // 24 hours max
	MinActiveJobs                = 1

	// Difficulty adjustment
	TargetSolveTimeSeconds = 10 * 60 // 10 minutes
	MinSolveTimeSeconds    = 8 * 60  // 8 minutes (too fast)
	MaxSolveTimeSeconds    = 12 * 60 // 12 minutes (too slow)
	AdjustmentInterval     = 3       // Adjust every 3 jobs

	// Problem size bounds
	MinProblemSize = 64
	MaxProblemSize = 2048
	SizeIncrement  = 16

	// Public submission requirements
	MinStakeForSubmission = 10_000_000_000 // 10,000 NEX in unexus
	PublicJobPostingFee   = 100_000_000    // 100 NEX in unexus (burned)
	MaxTitleLength        = 64
	MaxCategoryLength     = 32
)

// Job categories for public submissions
var ValidCategories = map[string]bool{
	"protein_folding":    true,
	"molecular_docking":  true,
	"materials_science":  true,
	"logistics":          true,
	"scheduling":         true,
	"graph_optimization": true,
	"machine_learning":   true,
	"cryptography":       true,
	"physics_simulation": true,
	"other":              true,
}

// Storage keys
var (
	ActiveJobCountKey     = []byte("active_job_count")
	CurrentJobIDKey       = []byte("current_job_id")
	BackgroundJobCountKey = []byte("bg_job_count")
	CurrentProblemSizeKey = []byte("current_problem_size")
	SolveTimeHistoryKey   = []byte("solve_time_history")
	PublicJobQueueKey     = []byte("public_job_queue")
	PublicJobCountKey     = []byte("public_job_count")
	SolutionKeyPrefix     = []byte{0x10}
)

// PublicJobMetadata contains on-chain metadata for public submissions
type PublicJobMetadata struct {
	Title       string `json:"title"`
	Category    string `json:"category"`
	Submitter   string `json:"submitter"`
	IpfsCid     string `json:"ipfs_cid"`
	Description string `json:"description,omitempty"`
}

// PublishedSolution contains the published solution data
type PublishedSolution struct {
	JobID            string `json:"job_id"`
	SolutionHash     string `json:"solution_hash"`
	EnergyAchieved   int64  `json:"energy_achieved"`
	SolverAddress    string `json:"solver_address"`
	SolveTimeSeconds int64  `json:"solve_time_seconds"`
	SolutionIpfsCid  string `json:"solution_ipfs_cid"`
	Timestamp        int64  `json:"timestamp"`
}

// ============================================
// BASIC GETTERS/SETTERS
// ============================================

func (k Keeper) GetActiveJobCount(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(ActiveJobCountKey)
	if bz == nil {
		return 0
	}
	return int64(bytesToUint64(bz))
}

func (k Keeper) SetActiveJobCount(ctx sdk.Context, count int64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(ActiveJobCountKey, uint64ToBytes(uint64(count)))
}

func (k Keeper) IncrementActiveJobCount(ctx sdk.Context) {
	k.SetActiveJobCount(ctx, k.GetActiveJobCount(ctx)+1)
}

func (k Keeper) DecrementActiveJobCount(ctx sdk.Context) {
	count := k.GetActiveJobCount(ctx)
	if count > 0 {
		k.SetActiveJobCount(ctx, count-1)
	}
}

func (k Keeper) GetCurrentJobID(ctx sdk.Context) string {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(CurrentJobIDKey)
	if bz == nil {
		return ""
	}
	return string(bz)
}

func (k Keeper) SetCurrentJobID(ctx sdk.Context, jobID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(CurrentJobIDKey, []byte(jobID))
}

func (k Keeper) GetBackgroundJobCount(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(BackgroundJobCountKey)
	if bz == nil {
		return 0
	}
	return int64(bytesToUint64(bz))
}

func (k Keeper) IncrementBackgroundJobCount(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	count := k.GetBackgroundJobCount(ctx)
	store.Set(BackgroundJobCountKey, uint64ToBytes(uint64(count+1)))
}

func (k Keeper) GetCurrentProblemSize(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(CurrentProblemSizeKey)
	if bz == nil {
		return MinProblemSize
	}
	return int64(bytesToUint64(bz))
}

func (k Keeper) SetCurrentProblemSize(ctx sdk.Context, size int64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(CurrentProblemSizeKey, uint64ToBytes(uint64(size)))
}

// ============================================
// SOLVE TIME TRACKING & DIFFICULTY ADJUSTMENT
// ============================================

func (k Keeper) GetSolveTimeHistory(ctx sdk.Context) []int64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(SolveTimeHistoryKey)
	if bz == nil || len(bz) == 0 {
		return []int64{}
	}

	times := make([]int64, len(bz)/8)
	for i := range times {
		times[i] = int64(binary.BigEndian.Uint64(bz[i*8 : (i+1)*8]))
	}
	return times
}

func (k Keeper) AddSolveTime(ctx sdk.Context, solveTimeSeconds int64) bool {
	store := ctx.KVStore(k.storeKey)
	history := k.GetSolveTimeHistory(ctx)

	history = append(history, solveTimeSeconds)
	if len(history) > AdjustmentInterval {
		history = history[len(history)-AdjustmentInterval:]
	}

	bz := make([]byte, len(history)*8)
	for i, t := range history {
		binary.BigEndian.PutUint64(bz[i*8:(i+1)*8], uint64(t))
	}
	store.Set(SolveTimeHistoryKey, bz)

	return len(history) >= AdjustmentInterval
}

func (k Keeper) AdjustDifficulty(ctx sdk.Context) {
	history := k.GetSolveTimeHistory(ctx)
	if len(history) < AdjustmentInterval {
		return
	}

	var total int64
	for _, t := range history {
		total += t
	}
	avgTime := total / int64(len(history))

	currentSize := k.GetCurrentProblemSize(ctx)
	newSize := currentSize

	if avgTime < MinSolveTimeSeconds {
		newSize = currentSize + SizeIncrement
		if newSize > MaxProblemSize {
			newSize = MaxProblemSize
		}
	} else if avgTime > MaxSolveTimeSeconds {
		newSize = currentSize - SizeIncrement
		if newSize < MinProblemSize {
			newSize = MinProblemSize
		}
	}

	if newSize != currentSize {
		k.SetCurrentProblemSize(ctx, newSize)
		k.Logger(ctx).Info("Difficulty adjusted",
			"avg_solve_time", avgTime,
			"old_size", currentSize,
			"new_size", newSize,
		)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"difficulty_adjusted",
				sdk.NewAttribute("avg_solve_time", fmt.Sprintf("%d", avgTime)),
				sdk.NewAttribute("old_size", fmt.Sprintf("%d", currentSize)),
				sdk.NewAttribute("new_size", fmt.Sprintf("%d", newSize)),
			),
		)
	}

	// Clear history after adjustment
	store := ctx.KVStore(k.storeKey)
	store.Delete(SolveTimeHistoryKey)
}

// ============================================
// PUBLIC JOB QUEUE - RANDOM SELECTION
// ============================================

func (k Keeper) GetPublicJobQueue(ctx sdk.Context) []string {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(PublicJobQueueKey)
	if bz == nil {
		return []string{}
	}

	queue := []string{}
	offset := 0
	for offset < len(bz) {
		if offset+4 > len(bz) {
			break
		}
		length := int(binary.BigEndian.Uint32(bz[offset : offset+4]))
		offset += 4
		if offset+length > len(bz) {
			break
		}
		queue = append(queue, string(bz[offset:offset+length]))
		offset += length
	}
	return queue
}

func (k Keeper) SetPublicJobQueue(ctx sdk.Context, queue []string) {
	store := ctx.KVStore(k.storeKey)

	totalSize := 0
	for _, id := range queue {
		totalSize += 4 + len(id)
	}

	bz := make([]byte, totalSize)
	offset := 0
	for _, id := range queue {
		binary.BigEndian.PutUint32(bz[offset:offset+4], uint32(len(id)))
		offset += 4
		copy(bz[offset:], id)
		offset += len(id)
	}

	store.Set(PublicJobQueueKey, bz)
}

func (k Keeper) AddToPublicJobQueue(ctx sdk.Context, jobID string) {
	queue := k.GetPublicJobQueue(ctx)
	queue = append(queue, jobID)
	k.SetPublicJobQueue(ctx, queue)
}

// SelectRandomFromQueue selects a random job from queue using block hash as randomness
// Returns job ID and removes it from queue
func (k Keeper) SelectRandomFromQueue(ctx sdk.Context) string {
	queue := k.GetPublicJobQueue(ctx)
	if len(queue) == 0 {
		return ""
	}

	// Use block height + timestamp as deterministic randomness source
	// This is predictable but fair - no one can game it easily
	seed := fmt.Sprintf("%d:%d", ctx.BlockHeight(), ctx.BlockTime().UnixNano())
	hash := sha256.Sum256([]byte(seed))
	
	// Convert first 8 bytes of hash to uint64 for random index
	randomValue := binary.BigEndian.Uint64(hash[:8])
	randomIndex := int(randomValue % uint64(len(queue)))

	// Select the job
	selectedJobID := queue[randomIndex]

	// Remove from queue (swap with last, then truncate)
	queue[randomIndex] = queue[len(queue)-1]
	queue = queue[:len(queue)-1]
	k.SetPublicJobQueue(ctx, queue)

	k.Logger(ctx).Info("Randomly selected public job",
		"job_id", selectedJobID,
		"queue_size", len(queue)+1,
		"selected_index", randomIndex,
	)

	return selectedJobID
}

func (k Keeper) GetPublicJobQueueLength(ctx sdk.Context) int {
	return len(k.GetPublicJobQueue(ctx))
}

// ============================================
// SOLUTION PUBLICATION
// ============================================

func (k Keeper) PublishSolution(ctx sdk.Context, solution PublishedSolution) {
	store := ctx.KVStore(k.storeKey)
	key := append(SolutionKeyPrefix, []byte(solution.JobID)...)

	bz := []byte(fmt.Sprintf("%s|%s|%d|%s|%d|%s|%d",
		solution.JobID,
		solution.SolutionHash,
		solution.EnergyAchieved,
		solution.SolverAddress,
		solution.SolveTimeSeconds,
		solution.SolutionIpfsCid,
		solution.Timestamp,
	))
	store.Set(key, bz)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"solution_published",
			sdk.NewAttribute("job_id", solution.JobID),
			sdk.NewAttribute("solver", solution.SolverAddress),
			sdk.NewAttribute("energy", fmt.Sprintf("%d", solution.EnergyAchieved)),
			sdk.NewAttribute("ipfs_cid", solution.SolutionIpfsCid),
		),
	)
}

// ============================================
// PROBLEM GENERATION
// ============================================

func GenerateIsingProblem(seed []byte, size int64) ([]byte, string) {
	h := sha256.New()
	h.Write(seed)

	numCouplings := size * size
	couplings := make([]byte, numCouplings)

	currentHash := h.Sum(nil)
	idx := 0
	for idx < int(numCouplings) {
		for _, b := range currentHash {
			if idx >= int(numCouplings) {
				break
			}
			couplings[idx] = b
			idx++
		}
		h.Reset()
		h.Write(currentHash)
		currentHash = h.Sum(nil)
	}

	problemHasher := sha256.New()
	problemHasher.Write(seed)
	problemHasher.Write([]byte(fmt.Sprintf("size:%d", size)))
	problemHash := hex.EncodeToString(problemHasher.Sum(nil))

	return couplings, problemHash
}

func CalculateThreshold(size int64) int64 {
	return -size / 2
}

// ============================================
// BACKGROUND JOB LIFECYCLE
// ============================================

func (k Keeper) GenerateSyntheticBackgroundJob(ctx sdk.Context) (*types.Job, error) {
	height := ctx.BlockHeight()
	timestamp := ctx.BlockTime().Unix()
	problemSize := k.GetCurrentProblemSize(ctx)

	seedData := fmt.Sprintf("nexus_ising_%d_%d_%d", height, timestamp, problemSize)
	seed := sha256.Sum256([]byte(seedData))

	problemData, problemHash := GenerateIsingProblem(seed[:], problemSize)
	threshold := CalculateThreshold(problemSize)

	jobID := fmt.Sprintf("sys_%d_%s", height, problemHash[:8])

	job := types.Job{
		Id:           jobID,
		Customer:     BackgroundJobCustomer,
		ProblemType:  "ising_synthetic",
		ProblemData:  problemData,
		ProblemHash:  problemHash,
		Threshold:    threshold,
		Reward:       0,
		Status:       types.JobStatusActive,
		BestEnergy:   0,
		TotalShares:  0,
		CreatedAt:    timestamp,
		Deadline:     timestamp + DefaultBackgroundJobDuration,
		IsBackground: true,
	}

	k.SetJob(ctx, job)
	k.SetCurrentJobID(ctx, jobID)
	k.IncrementActiveJobCount(ctx)

	k.Logger(ctx).Info("Generated synthetic background job",
		"job_id", jobID,
		"problem_size", problemSize,
		"threshold", threshold,
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"synthetic_job_generated",
			sdk.NewAttribute("job_id", jobID),
			sdk.NewAttribute("problem_size", fmt.Sprintf("%d", problemSize)),
			sdk.NewAttribute("threshold", fmt.Sprintf("%d", threshold)),
		),
	)

	return &job, nil
}

func (k Keeper) ActivateRandomPublicJob(ctx sdk.Context) (*types.Job, error) {
	jobID := k.SelectRandomFromQueue(ctx)
	if jobID == "" {
		return nil, nil
	}

	job, found := k.GetJob(ctx, jobID)
	if !found {
		return nil, fmt.Errorf("queued job not found: %s", jobID)
	}

	job.Status = types.JobStatusActive
	job.CreatedAt = ctx.BlockTime().Unix()
	job.Deadline = ctx.BlockTime().Unix() + DefaultBackgroundJobDuration
	k.SetJob(ctx, job)
	k.SetCurrentJobID(ctx, jobID)
	k.IncrementActiveJobCount(ctx)

	k.Logger(ctx).Info("Activated random public background job",
		"job_id", jobID,
		"submitter", job.Customer,
		"remaining_queue", k.GetPublicJobQueueLength(ctx),
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"public_job_activated",
			sdk.NewAttribute("job_id", jobID),
			sdk.NewAttribute("submitter", job.Customer),
		),
	)

	return &job, nil
}

// CheckAndGenerateBackgroundJob ensures there's always an active job
// Priority: Random selection from public queue, then synthetic generation

// CheckAndGenerateBackgroundJob ensures there's always an active job
// Priority: 1. Paid jobs (by priority fee), 2. Public jobs (random), 3. Synthetic
func (k Keeper) CheckAndGenerateBackgroundJob(ctx sdk.Context) {
	k.Logger(ctx).Info("CheckAndGenerateBackgroundJob called", "activeCount", k.GetActiveJobCount(ctx))
	activeCount := k.GetActiveJobCount(ctx)
	if activeCount >= MinActiveJobs {
		return
	}

	// Check if current job expired
	currentJobID := k.GetCurrentJobID(ctx)
	if currentJobID != "" {
		job, found := k.GetJob(ctx, currentJobID)
		if found && job.Status == types.JobStatusActive {
			if ctx.BlockTime().Unix() < job.Deadline {
				return
			}
			k.ExpireJob(ctx, currentJobID)
		}
	}

	// Priority 1: Try to activate highest priority PAID job
	paidJob, err := k.ActivateNextPaidJob(ctx)
	if err != nil {
		k.Logger(ctx).Error("Failed to activate paid job", "error", err)
	}
	if paidJob != nil {
		return
	}

	// Priority 2: Try to activate random PUBLIC job
	publicJob, err := k.ActivateRandomPublicJob(ctx)
	if err != nil {
		k.Logger(ctx).Error("Failed to activate public job", "error", err)
	}
	if publicJob != nil {
		return
	}

	// Priority 3: No queued jobs - generate SYNTHETIC
	_, err = k.GenerateSyntheticBackgroundJob(ctx)
	if err != nil {
		k.Logger(ctx).Error("Failed to generate synthetic job", "error", err)
	}
}

// ActivateNextPaidJob activates the highest priority paid job from queue
func (k Keeper) ActivateNextPaidJob(ctx sdk.Context) (*types.Job, error) {
	jobID := k.PopFromPaidJobQueue(ctx)
	if jobID == "" {
		return nil, nil
	}

	job, found := k.GetJob(ctx, jobID)
	if !found {
		return nil, fmt.Errorf("paid job not found: %s", jobID)
	}

	job.Status = types.JobStatusActive
	job.CreatedAt = ctx.BlockTime().Unix()
	job.Deadline = ctx.BlockTime().Unix() + DefaultBackgroundJobDuration
	k.SetJob(ctx, job)
	k.SetCurrentJobID(ctx, jobID)
	k.IncrementActiveJobCount(ctx)

	k.Logger(ctx).Info("Activated paid job",
		"job_id", jobID,
		"customer", job.Customer,
		"reward", job.Reward,
		"priority_fee", job.PriorityFee,
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"paid_job_activated",
			sdk.NewAttribute("job_id", jobID),
			sdk.NewAttribute("customer", job.Customer),
			sdk.NewAttribute("priority_fee", fmt.Sprintf("%d", job.PriorityFee)),
		),
	)

	return &job, nil
}
func (k Keeper) ExpireJob(ctx sdk.Context, jobID string) {
	job, found := k.GetJob(ctx, jobID)
	if !found || job.Status != types.JobStatusActive {
		return
	}

	job.Status = types.JobStatusExpired
	k.SetJob(ctx, job)
	k.DecrementActiveJobCount(ctx)

	if k.GetCurrentJobID(ctx) == jobID {
		k.SetCurrentJobID(ctx, "")
	}

	k.Logger(ctx).Info("Job expired", "job_id", jobID, "is_background", job.IsBackground)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"job_expired",
			sdk.NewAttribute("job_id", jobID),
		),
	)
}

func (k Keeper) OnJobSolved(ctx sdk.Context, jobID string, solverAddr string, solutionIpfsCid string) {
	job, found := k.GetJob(ctx, jobID)
	if !found {
		return
	}

	solveTime := ctx.BlockTime().Unix() - job.CreatedAt

	job.Status = types.JobStatusCompleted
	k.SetJob(ctx, job)
	k.DecrementActiveJobCount(ctx)

	if k.GetCurrentJobID(ctx) == jobID {
		k.SetCurrentJobID(ctx, "")
	}

	// Publish solution for public submissions (not system-generated synthetic jobs)
	if job.IsBackground && job.Customer != BackgroundJobCustomer {
		solutionHash := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", jobID, job.BestEnergy)))
		k.PublishSolution(ctx, PublishedSolution{
			JobID:            jobID,
			SolutionHash:     hex.EncodeToString(solutionHash[:]),
			EnergyAchieved:   job.BestEnergy,
			SolverAddress:    solverAddr,
			SolveTimeSeconds: solveTime,
			SolutionIpfsCid:  solutionIpfsCid,
			Timestamp:        ctx.BlockTime().Unix(),
		})
	}

	// Difficulty adjustment for all background jobs
	if job.IsBackground {
		k.IncrementBackgroundJobCount(ctx)
		if k.AddSolveTime(ctx, solveTime) {
			k.AdjustDifficulty(ctx)
		}
	}

	k.Logger(ctx).Info("Job solved",
		"job_id", jobID,
		"solve_time", solveTime,
		"best_energy", job.BestEnergy,
		"solver", solverAddr,
	)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"job_solved",
			sdk.NewAttribute("job_id", jobID),
			sdk.NewAttribute("solve_time", fmt.Sprintf("%d", solveTime)),
			sdk.NewAttribute("best_energy", fmt.Sprintf("%d", job.BestEnergy)),
			sdk.NewAttribute("solver", solverAddr),
		),
	)
}

// ============================================
// PAID JOB QUEUE (Priority Fee Sorted)
// ============================================

var PaidJobQueueKey = []byte("paid_job_queue")

// PaidJobEntry stores job ID with its priority fee for sorting
type PaidJobEntry struct {
	JobID       string
	PriorityFee int64
	SubmitTime  int64
}

// GetPaidJobQueue returns the paid job queue (sorted by priority fee desc, then time asc)
func (k Keeper) GetPaidJobQueue(ctx sdk.Context) []PaidJobEntry {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(PaidJobQueueKey)
	if bz == nil {
		return []PaidJobEntry{}
	}

	// Decode: each entry is 8 bytes priorityFee + 8 bytes submitTime + 4 bytes jobID length + jobID
	queue := []PaidJobEntry{}
	offset := 0
	for offset < len(bz) {
		if offset+20 > len(bz) {
			break
		}
		priorityFee := int64(binary.BigEndian.Uint64(bz[offset : offset+8]))
		offset += 8
		submitTime := int64(binary.BigEndian.Uint64(bz[offset : offset+8]))
		offset += 8
		idLen := int(binary.BigEndian.Uint32(bz[offset : offset+4]))
		offset += 4
		if offset+idLen > len(bz) {
			break
		}
		jobID := string(bz[offset : offset+idLen])
		offset += idLen
		queue = append(queue, PaidJobEntry{JobID: jobID, PriorityFee: priorityFee, SubmitTime: submitTime})
	}
	return queue
}

// SetPaidJobQueue saves the paid job queue
func (k Keeper) SetPaidJobQueue(ctx sdk.Context, queue []PaidJobEntry) {
	store := ctx.KVStore(k.storeKey)

	totalSize := 0
	for _, entry := range queue {
		totalSize += 8 + 8 + 4 + len(entry.JobID)
	}

	bz := make([]byte, totalSize)
	offset := 0
	for _, entry := range queue {
		binary.BigEndian.PutUint64(bz[offset:offset+8], uint64(entry.PriorityFee))
		offset += 8
		binary.BigEndian.PutUint64(bz[offset:offset+8], uint64(entry.SubmitTime))
		offset += 8
		binary.BigEndian.PutUint32(bz[offset:offset+4], uint32(len(entry.JobID)))
		offset += 4
		copy(bz[offset:], entry.JobID)
		offset += len(entry.JobID)
	}

	store.Set(PaidJobQueueKey, bz)
}

// AddToPaidJobQueue adds a job to the paid queue in sorted position
func (k Keeper) AddToPaidJobQueue(ctx sdk.Context, jobID string, priorityFee int64) int64 {
	queue := k.GetPaidJobQueue(ctx)
	submitTime := ctx.BlockTime().Unix()

	newEntry := PaidJobEntry{JobID: jobID, PriorityFee: priorityFee, SubmitTime: submitTime}

	// Find insertion point (sorted by priority fee desc, then submit time asc)
	insertIdx := len(queue)
	for i, entry := range queue {
		if priorityFee > entry.PriorityFee {
			insertIdx = i
			break
		} else if priorityFee == entry.PriorityFee && submitTime < entry.SubmitTime {
			insertIdx = i
			break
		}
	}

	// Insert at position
	queue = append(queue[:insertIdx], append([]PaidJobEntry{newEntry}, queue[insertIdx:]...)...)
	k.SetPaidJobQueue(ctx, queue)

	return int64(insertIdx + 1) // 1-indexed position
}

// PopFromPaidJobQueue removes and returns the highest priority job
func (k Keeper) PopFromPaidJobQueue(ctx sdk.Context) string {
	queue := k.GetPaidJobQueue(ctx)
	if len(queue) == 0 {
		return ""
	}

	jobID := queue[0].JobID
	k.SetPaidJobQueue(ctx, queue[1:])
	return jobID
}

// GetPaidJobQueueLength returns the number of paid jobs in queue
func (k Keeper) GetPaidJobQueueLength(ctx sdk.Context) int {
	return len(k.GetPaidJobQueue(ctx))
}

// RemoveFromPaidJobQueue removes a specific job from the queue (for cancellation)
func (k Keeper) RemoveFromPaidJobQueue(ctx sdk.Context, jobID string) bool {
	queue := k.GetPaidJobQueue(ctx)
	for i, entry := range queue {
		if entry.JobID == jobID {
			queue = append(queue[:i], queue[i+1:]...)
			k.SetPaidJobQueue(ctx, queue)
			return true
		}
	}
	return false
}

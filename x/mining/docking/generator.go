package docking

import (
	"fmt"
	"time"
)

// JobGenerator creates background docking jobs
// Uses: AlphaFold DB (CC-BY-4.0) + PubChem (public domain) + AutoDock Vina (Apache 2.0)
type JobGenerator struct {
	targetIndex   int
	ligandBatchIdx int
	ligandsPerJob int64
}

func NewJobGenerator(ligandsPerJob int64) *JobGenerator {
	return &JobGenerator{
		targetIndex:   0,
		ligandBatchIdx: 0,
		ligandsPerJob: ligandsPerJob,
	}
}

// GenerateNextJob creates the next background job
func (g *JobGenerator) GenerateNextJob() (*Job, *FetchedProtein, error) {
	target := GetNextBackgroundTarget(g.targetIndex)
	g.targetIndex++

	protein, err := FetchProtein(target)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch %s: %w", target.UniprotID, err)
	}

	job := CreateBackgroundJob(protein, g.ligandsPerJob)
	
	// Track which ligand batch this job uses
	g.ligandBatchIdx++

	return job, protein, nil
}

// JobStatus tracks active jobs
type JobStatus struct {
	Job          *Job
	Protein      *FetchedProtein
	ResultsCount int64
	HitsCount    int
	StartedAt    time.Time
}

// BackgroundJobManager manages the background job queue
type BackgroundJobManager struct {
	generator    *JobGenerator
	activeJobs   map[string]*JobStatus
	hitThreshold float64
}

func NewBackgroundJobManager() *BackgroundJobManager {
	return &BackgroundJobManager{
		generator:    NewJobGenerator(10000),
		activeJobs:   make(map[string]*JobStatus),
		hitThreshold: -7.0, // kcal/mol
	}
}

func (m *BackgroundJobManager) StartNewJob() (*Job, error) {
	job, protein, err := m.generator.GenerateNextJob()
	if err != nil {
		return nil, err
	}

	m.activeJobs[job.ID] = &JobStatus{
		Job:       job,
		Protein:   protein,
		StartedAt: time.Now(),
	}

	return job, nil
}

func (m *BackgroundJobManager) RecordResult(result *Result) error {
	status, exists := m.activeJobs[result.JobID]
	if !exists {
		return fmt.Errorf("job %s not found", result.JobID)
	}

	status.ResultsCount++
	status.Job.DockedCount++

	if result.Score < m.hitThreshold {
		status.HitsCount++
	}

	return nil
}

func (m *BackgroundJobManager) GetJobProgress(jobID string) (docked, total int64, hits int) {
	status, exists := m.activeJobs[jobID]
	if !exists {
		return 0, 0, 0
	}
	return status.Job.DockedCount, status.Job.TotalLigands, status.HitsCount
}

func (m *BackgroundJobManager) IsJobComplete(jobID string) bool {
	status, exists := m.activeJobs[jobID]
	if !exists {
		return false
	}
	return status.Job.DockedCount >= status.Job.TotalLigands
}

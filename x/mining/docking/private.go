package docking

// PrivateJob - target is secret, results are public
type PrivateJob struct {
	ID           string `json:"id"`
	Customer     string `json:"customer"`
	TargetHash   string `json:"target_hash"`   // Hash of protein - customer proves later if needed
	LigandBatch  string `json:"ligand_batch"`  // Public ZINC batch
	Reward       int64  `json:"reward"`
	Deadline     int64  `json:"deadline"`
	Status       int    `json:"status"`
}

// Customer sends actual protein PDB to assigned miners off-chain
// Miners post public results: "ZINC000012345 scores -9.2 kcal/mol for job X"
// World sees the drug hits, but not what disease/protein they target
// Only customer knows the protein identity

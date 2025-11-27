package docking

// Job is a docking job (public or private target)
type Job struct {
	ID            string `json:"id"`
	
	// Target - public jobs show ProteinID, private only TargetHash
	ProteinID     string `json:"protein_id,omitempty"`
	TargetHash    string `json:"target_hash"`
	BindingSite   Site   `json:"binding_site"`
	
	// Ligands to screen
	LigandTranche string `json:"ligand_tranche"`
	TotalLigands  int64  `json:"total_ligands"`
	DockedCount   int64  `json:"docked_count"`
	
	// Payment (for paid jobs)
	Customer      string `json:"customer,omitempty"`
	Reward        int64  `json:"reward,omitempty"`
	IsBackground  bool   `json:"is_background"`
	
	CreatedAt     int64  `json:"created_at"`
	Deadline      int64  `json:"deadline"`
}

// Site defines the docking search box
type Site struct {
	CenterX float64 `json:"center_x"`
	CenterY float64 `json:"center_y"`
	CenterZ float64 `json:"center_z"`
	SizeX   float64 `json:"size_x"`
	SizeY   float64 `json:"size_y"`
	SizeZ   float64 `json:"size_z"`
}

// Result is ONE ligand docked - smallest unit of work
type Result struct {
	JobID          string  `json:"job_id"`
	LigandID       string  `json:"ligand_id"`
	LigandSMILES   string  `json:"ligand_smiles"`
	RotatableBonds int     `json:"rotatable_bonds"`
	Score          float64 `json:"score"`
	Miner          string  `json:"miner"`
	Reward         int64   `json:"reward"`
	BlockHeight    int64   `json:"block_height"`
}

// Hit is a promising drug candidate (score below threshold)
type Hit struct {
	JobID        string  `json:"job_id"`
	ProteinID    string  `json:"protein_id"`
	LigandID     string  `json:"ligand_id"`
	LigandSMILES string  `json:"ligand_smiles"`
	Score        float64 `json:"score"`
	Rank         int     `json:"rank"`
	Miner        string  `json:"miner"`
	BlockHeight  int64   `json:"block_height"`
}

// MinerStats tracks a miner's contributions
type MinerStats struct {
	Address       string `json:"address"`
	LigandsDocked int64  `json:"ligands_docked"`
	TotalReward   int64  `json:"total_reward"`
	HitsFound     int    `json:"hits_found"`
	BestScore     float64 `json:"best_score"`
}

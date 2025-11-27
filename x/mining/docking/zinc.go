package docking

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ZINC15 database URLs
const (
	ZINC15BaseURL     = "https://zinc15.docking.org"
	ZINC20BaseURL     = "https://zinc20.docking.org"  // Newer version
	ZINCTranchesURL   = "https://zinc15.docking.org/tranches/all3D.txt"
	ZINCSubstancesAPI = "https://zinc15.docking.org/substances"
)

// ZINCSubstance represents a compound from ZINC database
type ZINCSubstance struct {
	ZINCID       string  `json:"zinc_id"`
	SMILES       string  `json:"smiles"`
	InChIKey     string  `json:"inchikey"`
	MolWeight    float64 `json:"mwt"`
	LogP         float64 `json:"logp"`
	HBondDonors  int     `json:"hbd"`
	HBondAccept  int     `json:"hba"`
	RotBonds     int     `json:"rotatable_bonds"`
	TPSA         float64 `json:"tpsa"`  // Topological polar surface area
	Purchasable  string  `json:"purchasability"`
	Reactivity   string  `json:"reactivity"`
}

// ZINCTranche represents a subset of ZINC compounds organized by properties
type ZINCTranche struct {
	ID          string `json:"id"`           // e.g., "AAAA"
	LogP        string `json:"logp_range"`   // e.g., "M" for medium
	MWT         string `json:"mwt_range"`    // e.g., "250-300"
	Count       int64  `json:"count"`
	DownloadURL string `json:"download_url"`
}

// DrugLikeTranches returns tranches that satisfy Lipinski's Rule of Five
// These are most likely to be viable drug candidates
var DrugLikeTranches = []string{
	// Format: [LogP bin][MWT bin][Reactivity][Purchasability]
	// LogP: A(-1 to 0), B(0-1), C(1-2), D(2-3), E(3-4), F(4-5)
	// MWT: A(200-250), B(250-300), C(300-325), D(325-350), E(350-375), F(375-400), G(400-425), H(425-450), I(450-500)
	"BAAB", "BABB", "BACB", "BADB", "BAEB",
	"CAAB", "CABB", "CACB", "CADB", "CAEB",
	"DAAB", "DABB", "DACB", "DADB", "DAEB",
	"EAAB", "EABB", "EACB", "EADB", "EAEB",
}

var zincClient = &http.Client{
	Timeout: 60 * time.Second,
}

// FetchZINCTranches gets available tranches from ZINC
func FetchZINCTranches() ([]ZINCTranche, error) {
	resp, err := zincClient.Get(ZINCTranchesURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tranches: %w", err)
	}
	defer resp.Body.Close()
	
	var tranches []ZINCTranche
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			tranches = append(tranches, ZINCTranche{
				ID:          parts[0],
				DownloadURL: fmt.Sprintf("%s/tranches/download/%s.smi.gz", ZINC15BaseURL, parts[0]),
			})
		}
	}
	
	return tranches, scanner.Err()
}

// FetchLigandBatch downloads a batch of ligands from ZINC
func FetchLigandBatch(trancheID string, startIdx, count int) ([]ZINCSubstance, error) {
	// Use ZINC API for substance lookup
	url := fmt.Sprintf("%s.json?tranche=%s&start=%d&count=%d", 
		ZINCSubstancesAPI, trancheID, startIdx, count)
	
	resp, err := zincClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ligands: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ZINC API returned status %d", resp.StatusCode)
	}
	
	var substances []ZINCSubstance
	if err := json.NewDecoder(resp.Body).Decode(&substances); err != nil {
		return nil, fmt.Errorf("failed to decode substances: %w", err)
	}
	
	return substances, nil
}

// FetchSMILESBatch downloads SMILES strings for a tranche
func FetchSMILESBatch(trancheURL string, maxCount int) ([]ZINCSubstance, error) {
	resp, err := zincClient.Get(trancheURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch SMILES: %w", err)
	}
	defer resp.Body.Close()
	
	var reader io.Reader = resp.Body
	
	// Handle gzip compression
	if strings.HasSuffix(trancheURL, ".gz") {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}
	
	var substances []ZINCSubstance
	scanner := bufio.NewScanner(reader)
	count := 0
	
	for scanner.Scan() && count < maxCount {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			substances = append(substances, ZINCSubstance{
				SMILES: parts[0],
				ZINCID: parts[1],
			})
			count++
		}
	}
	
	return substances, scanner.Err()
}

// ValidateDrugLikeness checks Lipinski's Rule of Five
func (s *ZINCSubstance) ValidateDrugLikeness() bool {
	violations := 0
	
	if s.MolWeight > 500 {
		violations++
	}
	if s.LogP > 5 {
		violations++
	}
	if s.HBondDonors > 5 {
		violations++
	}
	if s.HBondAccept > 10 {
		violations++
	}
	
	// Allow 1 violation (common in modern drug discovery)
	return violations <= 1
}

// LigandBatchSpec specifies which ligands to dock
type LigandBatchSpec struct {
	TrancheID   string   `json:"tranche_id"`
	ZINCIDs     []string `json:"zinc_ids"`
	SMILESURL   string   `json:"smiles_url"`    // Direct download URL
	StartIndex  int64    `json:"start_index"`
	BatchSize   int      `json:"batch_size"`
	BatchHash   string   `json:"batch_hash"`    // SHA256 of ligand list
}

// CreateLigandBatchSpec creates a specification for a ligand batch
func CreateLigandBatchSpec(trancheID string, startIdx int64, batchSize int) LigandBatchSpec {
	return LigandBatchSpec{
		TrancheID:  trancheID,
		StartIndex: startIdx,
		BatchSize:  batchSize,
		SMILESURL:  fmt.Sprintf("%s/tranches/download/%s.smi.gz", ZINC15BaseURL, trancheID),
	}
}

// EstimateTotalLigands returns approximate number of drug-like compounds in ZINC
func EstimateTotalLigands() int64 {
	// ZINC15 contains ~230 million purchasable compounds
	// ~750 million total, but we focus on drug-like subset
	return 230_000_000
}

// GetDrugLikeSubset returns URL for pre-filtered drug-like compounds
func GetDrugLikeSubset() string {
	// ZINC provides pre-filtered "in-stock" drug-like compounds
	return "https://zinc15.docking.org/tranches/download/drug-like-in-stock.smi.gz"
}

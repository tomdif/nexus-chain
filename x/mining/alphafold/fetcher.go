package alphafold

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

// AlphaFoldPrediction represents the API response
type AlphaFoldPrediction struct {
	UniprotAccession   string  `json:"uniprotAccession"`
	UniprotID          string  `json:"uniprotId"`
	UniprotDescription string  `json:"uniprotDescription"`
	Gene               string  `json:"gene"`
	Sequence           string  `json:"sequence"`
	PdbURL             string  `json:"pdbUrl"`
	GlobalMetricValue  float64 `json:"globalMetricValue"`
	OrganismScientific string  `json:"organismScientificName"`
}

// Atom represents a C-alpha atom from PDB
type Atom struct {
	ResNum  int
	Residue string
	X, Y, Z float64
}

// ProteinJob represents a job derived from a protein structure
type ProteinJob struct {
	ProblemType  string `json:"problem_type"`
	ProblemData  string `json:"problem_data"` // base64 encoded
	ProblemHash  string `json:"problem_hash"`
	ProblemSize  int    `json:"problem_size"`
	Threshold    int64  `json:"threshold"`
	ProteinName  string `json:"protein_name"`
	UniprotID    string `json:"uniprot_id"`
	Description  string `json:"description"`
	NumContacts  int    `json:"num_contacts"`
	NumResidues  int    `json:"num_residues"`
}

// MedicallyImportantProteins - curated list for background jobs
var MedicallyImportantProteins = []struct {
	UniprotID   string
	Name        string
	Description string
}{
	{"P04637", "TP53", "Tumor protein p53 - cancer suppressor"},
	{"P00533", "EGFR", "Epidermal growth factor receptor - cancer"},
	{"P38398", "BRCA1", "BRCA1 DNA repair - breast cancer"},
	{"P01308", "INS", "Insulin - diabetes"},
	{"P05067", "APP", "Amyloid precursor protein - Alzheimer's"},
	{"P04062", "GBA", "Glucocerebrosidase - Parkinson's"},
	{"Q99720", "SIGMAR1", "Sigma-1 receptor - neurodegeneration"},
	{"P02768", "ALB", "Serum albumin - drug binding"},
	{"P68871", "HBB", "Hemoglobin beta - sickle cell"},
	{"P00558", "PGK1", "Phosphoglycerate kinase - metabolism"},
	{"P60709", "ACTB", "Beta-actin - cytoskeleton"},
	{"P0DP23", "CALM1", "Calmodulin - calcium signaling"},
	{"P69905", "HBA1", "Hemoglobin alpha - oxygen transport"},
	{"P01375", "TNF", "Tumor necrosis factor - inflammation"},
	{"P01579", "IFNG", "Interferon gamma - immune response"},
}

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

// FetchProteinInfo gets protein metadata from AlphaFold API
func FetchProteinInfo(uniprotID string) (*AlphaFoldPrediction, error) {
	url := fmt.Sprintf("https://alphafold.ebi.ac.uk/api/prediction/%s", uniprotID)
	
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch protein info: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	
	var predictions []AlphaFoldPrediction
	if err := json.NewDecoder(resp.Body).Decode(&predictions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if len(predictions) == 0 {
		return nil, fmt.Errorf("no predictions found for %s", uniprotID)
	}
	
	return &predictions[0], nil
}

// FetchPDBStructure downloads the PDB file and parses C-alpha atoms
func FetchPDBStructure(pdbURL string) ([]Atom, error) {
	resp, err := httpClient.Get(pdbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PDB: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PDB fetch returned status %d", resp.StatusCode)
	}
	
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDB: %w", err)
	}
	
	return ParsePDB(string(data))
}

// ParsePDB extracts C-alpha atoms from PDB content
func ParsePDB(content string) ([]Atom, error) {
	var atoms []Atom
	
	for _, line := range strings.Split(content, "\n") {
		if !strings.HasPrefix(line, "ATOM") {
			continue
		}
		if len(line) < 54 {
			continue
		}
		
		// Check if it's a C-alpha atom
		atomName := strings.TrimSpace(line[12:16])
		if atomName != "CA" {
			continue
		}
		
		resNum := 0
		fmt.Sscanf(strings.TrimSpace(line[22:26]), "%d", &resNum)
		
		var x, y, z float64
		fmt.Sscanf(line[30:38], "%f", &x)
		fmt.Sscanf(line[38:46], "%f", &y)
		fmt.Sscanf(line[46:54], "%f", &z)
		
		atoms = append(atoms, Atom{
			ResNum:  resNum,
			Residue: strings.TrimSpace(line[17:20]),
			X:       x,
			Y:       y,
			Z:       z,
		})
	}
	
	return atoms, nil
}

// CreateProteinJob converts a protein structure to an Ising optimization problem
func CreateProteinJob(info *AlphaFoldPrediction, atoms []Atom, maxResidues int) (*ProteinJob, error) {
	if len(atoms) == 0 {
		return nil, fmt.Errorf("no atoms found")
	}
	
	// Limit problem size
	n := len(atoms)
	if n > maxResidues {
		n = maxResidues
		atoms = atoms[:n]
	}
	
	// Create contact map (residues within 8 Angstroms, separated by >= 3 in sequence)
	type Contact struct {
		I, J     int
		Distance float64
	}
	var contacts []Contact
	
	cutoff := 8.0
	minSeqSep := 3
	
	for i := 0; i < n; i++ {
		for j := i + minSeqSep; j < n; j++ {
			dx := atoms[i].X - atoms[j].X
			dy := atoms[i].Y - atoms[j].Y
			dz := atoms[i].Z - atoms[j].Z
			dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
			
			if dist < cutoff {
				contacts = append(contacts, Contact{I: i, J: j, Distance: dist})
			}
		}
	}
	
	// Build binary problem data
	// Format: [n (4 bytes)][num_couplings (4 bytes)][couplings...][fields...]
	data := make([]byte, 0, 8+len(contacts)*8+n*4)
	
	// Problem size
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(n))
	data = append(data, buf...)
	
	// Number of couplings
	binary.LittleEndian.PutUint32(buf, uint32(len(contacts)))
	data = append(data, buf...)
	
	// Couplings: i (2 bytes), j (2 bytes), strength (4 bytes float32)
	for _, c := range contacts {
		binary.LittleEndian.PutUint16(buf[:2], uint16(c.I))
		data = append(data, buf[:2]...)
		binary.LittleEndian.PutUint16(buf[:2], uint16(c.J))
		data = append(data, buf[:2]...)
		
		// Coupling strength: negative (ferromagnetic) inversely proportional to distance
		strength := float32(-1.0 / (c.Distance + 0.1))
		binary.LittleEndian.PutUint32(buf, math.Float32bits(strength))
		data = append(data, buf...)
	}
	
	// Local fields (all zero for now)
	for i := 0; i < n; i++ {
		binary.LittleEndian.PutUint32(buf, 0)
		data = append(data, buf...)
	}
	
	// Compute hash
	hash := sha256.Sum256(data)
	
	// Threshold: 80% of maximum possible energy reduction
	threshold := int64(-float64(len(contacts)) * 0.8)
	
	return &ProteinJob{
		ProblemType: "protein_folding",
		ProblemData: base64.StdEncoding.EncodeToString(data),
		ProblemHash: fmt.Sprintf("%x", hash),
		ProblemSize: n,
		Threshold:   threshold,
		ProteinName: info.Gene,
		UniprotID:   info.UniprotAccession,
		Description: info.UniprotDescription,
		NumContacts: len(contacts),
		NumResidues: n,
	}, nil
}

// FetchAndCreateJob is a convenience function to fetch and create a job in one call
func FetchAndCreateJob(uniprotID string, maxResidues int) (*ProteinJob, error) {
	info, err := FetchProteinInfo(uniprotID)
	if err != nil {
		return nil, fmt.Errorf("fetch info failed: %w", err)
	}
	
	atoms, err := FetchPDBStructure(info.PdbURL)
	if err != nil {
		return nil, fmt.Errorf("fetch PDB failed: %w", err)
	}
	
	return CreateProteinJob(info, atoms, maxResidues)
}

// GetRandomProtein returns a random protein from the curated list
func GetRandomProtein(index int) (uniprotID, name, description string) {
	idx := index % len(MedicallyImportantProteins)
	p := MedicallyImportantProteins[idx]
	return p.UniprotID, p.Name, p.Description
}

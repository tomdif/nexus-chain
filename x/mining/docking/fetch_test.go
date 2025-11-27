package docking

import (
	"fmt"
	"testing"
)

func TestFetchAlphaFold(t *testing.T) {
	pdb, err := FetchFromAlphaFold("P01308")
	if err != nil {
		t.Fatalf("Failed to fetch: %v", err)
	}
	
	fmt.Printf("Fetched %d bytes\n", len(pdb))
	fmt.Printf("Hash: %s\n", HashPDB(pdb)[:16])
	
	site := ParseBindingSite(pdb)
	fmt.Printf("Binding site center: %.1f, %.1f, %.1f\n", site.CenterX, site.CenterY, site.CenterZ)
}

func TestGenerateJob(t *testing.T) {
	gen := NewJobGenerator(1000)
	job, protein, err := gen.GenerateNextJob()
	if err != nil {
		t.Fatalf("Failed to generate: %v", err)
	}
	
	fmt.Printf("Job ID: %s\n", job.ID)
	fmt.Printf("Protein: %s (%s)\n", protein.Name, protein.Disease)
	fmt.Printf("License: %s\n", protein.License)
	fmt.Printf("Total ligands: %d\n", job.TotalLigands)
}

func TestFetchPubChem(t *testing.T) {
	ligands, err := FetchFromPubChem(2244, 5)
	if err != nil {
		t.Fatalf("Failed to fetch: %v", err)
	}
	
	fmt.Printf("Fetched %d ligands:\n", len(ligands))
	for _, l := range ligands {
		smiles := l.SMILES
		if len(smiles) > 40 {
			smiles = smiles[:40] + "..."
		}
		fmt.Printf("  %s: %s (bonds: %d)\n", l.ZINCID, smiles, l.RotatableBonds)
	}
}

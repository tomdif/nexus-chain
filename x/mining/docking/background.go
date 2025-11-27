package docking

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// LEGAL STATUS - ALL SOURCES PERMIT COMMERCIAL USE:
// =================================================
// 
// 1. AlphaFold Database (pre-computed structures)
//    License: CC-BY 4.0
//    Source: https://alphafold.ebi.ac.uk/
//    "Data is available for academic and commercial use"
//    Attribution: Cite Jumper et al. 2021
//
// 2. PubChem (compound data)
//    License: Public Domain (US Government work)
//    Source: https://pubchem.ncbi.nlm.nih.gov/
//    "NCBI places no restrictions on use or distribution"
//    "No restrictions on commercial purposes"
//
// 3. AutoDock Vina (docking engine)
//    License: Apache 2.0
//    Commercial use permitted
//
// We do NOT run AlphaFold model - only download pre-computed structures.

const AlphaFoldDBURL = "https://alphafold.ebi.ac.uk/files"

var BackgroundTargets = []Target{
	{UniprotID: "P04637", Name: "TP53", Disease: "Cancer"},
	{UniprotID: "P00533", Name: "EGFR", Disease: "Cancer"},
	{UniprotID: "P38398", Name: "BRCA1", Disease: "Breast Cancer"},
	{UniprotID: "P01308", Name: "INS", Disease: "Diabetes"},
	{UniprotID: "P05067", Name: "APP", Disease: "Alzheimer's"},
	{UniprotID: "P04062", Name: "GBA", Disease: "Parkinson's"},
	{UniprotID: "P02768", Name: "ALB", Disease: "Drug Binding"},
	{UniprotID: "P68871", Name: "HBB", Disease: "Sickle Cell"},
	{UniprotID: "P00558", Name: "PGK1", Disease: "Metabolism"},
	{UniprotID: "P60709", Name: "ACTB", Disease: "Cytoskeleton"},
	{UniprotID: "P01375", Name: "TNF", Disease: "Inflammation"},
	{UniprotID: "P01579", Name: "IFNG", Disease: "Immune"},
	{UniprotID: "P15056", Name: "BRAF", Disease: "Melanoma"},
	{UniprotID: "P00519", Name: "ABL1", Disease: "Leukemia"},
	{UniprotID: "P04406", Name: "GAPDH", Disease: "Metabolism"},
}

type Target struct {
	UniprotID string
	Name      string
	Disease   string
}

type FetchedProtein struct {
	UniprotID   string
	Name        string
	Disease     string
	PDBContent  string
	PDBHash     string
	BindingSite Site
	FetchedAt   int64
	License     string
}

var httpClient = &http.Client{Timeout: 30 * time.Second}

// FetchFromAlphaFold downloads pre-computed structure from AlphaFold Database
func FetchFromAlphaFold(uniprotID string) (string, error) {
	// Current version is v6, fallback to older versions
	for _, ver := range []string{"v6", "v4", "v3", "v2"} {
		url := fmt.Sprintf("%s/AF-%s-F1-model_%s.pdb", AlphaFoldDBURL, uniprotID, ver)
		
		resp, err := httpClient.Get(url)
		if err != nil {
			continue
		}
		
		if resp.StatusCode == 200 {
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				return "", err
			}
			return string(body), nil
		}
		resp.Body.Close()
	}
	return "", fmt.Errorf("failed to fetch %s from AlphaFold DB", uniprotID)
}

func ParseBindingSite(pdbContent string) Site {
	var sumX, sumY, sumZ float64
	var count int

	for _, line := range strings.Split(pdbContent, "\n") {
		if !strings.HasPrefix(line, "ATOM") || len(line) < 54 {
			continue
		}
		if strings.TrimSpace(line[12:16]) != "CA" {
			continue
		}
		var x, y, z float64
		fmt.Sscanf(line[30:38], "%f", &x)
		fmt.Sscanf(line[38:46], "%f", &y)
		fmt.Sscanf(line[46:54], "%f", &z)
		sumX += x
		sumY += y
		sumZ += z
		count++
	}

	if count == 0 {
		return Site{SizeX: 30, SizeY: 30, SizeZ: 30}
	}
	return Site{
		CenterX: sumX / float64(count),
		CenterY: sumY / float64(count),
		CenterZ: sumZ / float64(count),
		SizeX:   30, SizeY: 30, SizeZ: 30,
	}
}

func HashPDB(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

func FetchProtein(target Target) (*FetchedProtein, error) {
	pdb, err := FetchFromAlphaFold(target.UniprotID)
	if err != nil {
		return nil, err
	}
	return &FetchedProtein{
		UniprotID:   target.UniprotID,
		Name:        target.Name,
		Disease:     target.Disease,
		PDBContent:  pdb,
		PDBHash:     HashPDB(pdb),
		BindingSite: ParseBindingSite(pdb),
		FetchedAt:   time.Now().Unix(),
		License:     "CC-BY-4.0 (AlphaFold DB)",
	}, nil
}

func CreateBackgroundJob(protein *FetchedProtein, totalLigands int64) *Job {
	return &Job{
		ID:           fmt.Sprintf("bg_%s_%d", protein.UniprotID, time.Now().Unix()),
		ProteinID:    protein.UniprotID,
		TargetHash:   protein.PDBHash,
		BindingSite:  protein.BindingSite,
		TotalLigands: totalLigands,
		IsBackground: true,
		CreatedAt:    time.Now().Unix(),
		Deadline:     time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
}

func GetNextBackgroundTarget(index int) Target {
	return BackgroundTargets[index%len(BackgroundTargets)]
}

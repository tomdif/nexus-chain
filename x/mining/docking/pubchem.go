package docking

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const PubChemAPI = "https://pubchem.ncbi.nlm.nih.gov/rest/pug"

var pubchemClient = &http.Client{Timeout: 30 * time.Second}

type PubChemResponse struct {
	PropertyTable struct {
		Properties []PubChemCompound `json:"Properties"`
	} `json:"PropertyTable"`
}

type PubChemCompound struct {
	CID                int    `json:"CID"`
	MolecularWeight    string `json:"MolecularWeight"`
	SMILES             string `json:"ConnectivitySMILES"`
	RotatableBondCount int    `json:"RotatableBondCount"`
}

// FetchFromPubChem gets compounds by CID range
func FetchFromPubChem(startCID, count int) ([]Ligand, error) {
	cids := make([]string, count)
	for i := 0; i < count; i++ {
		cids[i] = strconv.Itoa(startCID + i)
	}
	cidList := strings.Join(cids, ",")

	url := fmt.Sprintf("%s/compound/cid/%s/property/CanonicalSMILES,MolecularWeight,RotatableBondCount/JSON",
		PubChemAPI, cidList)

	resp, err := pubchemClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("PubChem fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("PubChem returned status %d", resp.StatusCode)
	}

	var result PubChemResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode: %w", err)
	}

	ligands := make([]Ligand, 0, len(result.PropertyTable.Properties))
	for _, c := range result.PropertyTable.Properties {
		ligands = append(ligands, Ligand{
			ZINCID:         fmt.Sprintf("CID%d", c.CID),
			SMILES:         c.SMILES,
			RotatableBonds: c.RotatableBondCount,
		})
	}

	return ligands, nil
}

// FetchDrugLike gets compounds from known drug CID ranges
func FetchDrugLike(batchIndex, count int) ([]Ligand, error) {
	startCID := GetDrugLikeCID(batchIndex, count)
	return FetchFromPubChem(startCID, count)
}

// GetDrugLikeCID returns starting CID for drug-like compounds
func GetDrugLikeCID(batchIndex, batchSize int) int {
	ranges := []int{2000, 3000, 4000, 5000, 10000, 20000, 50000}
	rangeIdx := batchIndex % len(ranges)
	offset := (batchIndex / len(ranges)) * batchSize
	return ranges[rangeIdx] + offset
}

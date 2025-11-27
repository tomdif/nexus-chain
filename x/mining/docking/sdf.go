package docking

import (
	"fmt"
	"strings"
	"time"
)

// SDF property keys (standard for chemistry software)
const (
	PropZINCID     = "ZINC_ID"
	PropSMILES     = "SMILES"
	PropScore      = "BINDING_AFFINITY_KCAL_MOL"
	PropTarget     = "TARGET_PROTEIN"
	PropRank       = "RANK"
	PropMiner      = "DISCOVERED_BY"
	PropBlock      = "BLOCK_HEIGHT"
	PropTimestamp  = "TIMESTAMP"
)

// HitToSDF converts a hit to standard SDF format
func HitToSDF(h Hit) string {
	var sb strings.Builder
	
	// Molecule name
	sb.WriteString(fmt.Sprintf("%s\n", h.LigandID))
	
	// Program line
	sb.WriteString(fmt.Sprintf("  NEXUS     %s\n", time.Now().Format("01021506")))
	
	// Comment
	sb.WriteString(fmt.Sprintf("Target: %s  Score: %.2f kcal/mol\n", h.ProteinID, h.Score))
	
	// Counts line (no atoms - just metadata)
	sb.WriteString("  0  0  0  0  0  0  0  0  0  0999 V2000\n")
	sb.WriteString("M  END\n")
	
	// Properties
	writeProperty(&sb, PropZINCID, h.LigandID)
	writeProperty(&sb, PropSMILES, h.LigandSMILES)
	writeProperty(&sb, PropScore, fmt.Sprintf("%.2f", h.Score))
	writeProperty(&sb, PropTarget, h.ProteinID)
	writeProperty(&sb, PropRank, fmt.Sprintf("%d", h.Rank))
	writeProperty(&sb, PropMiner, h.Miner)
	writeProperty(&sb, PropBlock, fmt.Sprintf("%d", h.BlockHeight))
	writeProperty(&sb, PropTimestamp, time.Now().UTC().Format(time.RFC3339))
	
	sb.WriteString("$$$$\n")
	return sb.String()
}

func writeProperty(sb *strings.Builder, name, value string) {
	if value == "" {
		return
	}
	sb.WriteString(fmt.Sprintf(">  <%s>\n%s\n\n", name, value))
}

// HitsToSDF converts multiple hits to SDF format
func HitsToSDF(hits []Hit, jobID string) string {
	var sb strings.Builder
	
	// Header comment
	sb.WriteString(fmt.Sprintf("# NEXUS Drug Discovery Results\n"))
	sb.WriteString(fmt.Sprintf("# Job: %s\n", jobID))
	sb.WriteString(fmt.Sprintf("# Hits: %d\n", len(hits)))
	sb.WriteString(fmt.Sprintf("# Generated: %s\n\n", time.Now().UTC().Format(time.RFC3339)))
	
	for _, h := range hits {
		sb.WriteString(HitToSDF(h))
	}
	
	return sb.String()
}

// HitsToCSV exports hits as CSV
func HitsToCSV(hits []Hit) string {
	var sb strings.Builder
	sb.WriteString("rank,zinc_id,smiles,score_kcal_mol,target,miner,block\n")
	
	for _, h := range hits {
		sb.WriteString(fmt.Sprintf("%d,%s,%s,%.2f,%s,%s,%d\n",
			h.Rank, h.LigandID, h.LigandSMILES, h.Score, h.ProteinID, h.Miner, h.BlockHeight))
	}
	
	return sb.String()
}

#!/usr/bin/env python3
"""
Fetch protein structures from AlphaFold for NEXUS mining jobs.
Focus on medically important proteins.
"""

import subprocess
import json
import os

# Medically important proteins with their UniProt IDs
PROTEINS = [
    ("P04637", "TP53_human", "Tumor protein p53 - cancer"),
    ("P00533", "EGFR_human", "Epidermal growth factor receptor - cancer"),
    ("P38398", "BRCA1_human", "BRCA1 DNA repair - breast cancer"),
    ("P01308", "INS_human", "Insulin - diabetes"),
    ("P05067", "APP_human", "Amyloid precursor protein - Alzheimer's"),
    ("P10636", "TAU_human", "Microtubule-associated protein tau - Alzheimer's"),
    ("P04062", "GBA_human", "Glucocerebrosidase - Parkinson's"),
    ("Q99720", "SIGM1_human", "Sigma-1 receptor - neurodegeneration"),
    ("P02768", "ALB_human", "Serum albumin - drug binding"),
    ("P68871", "HBB_human", "Hemoglobin beta - sickle cell"),
]

def fetch_protein(uniprot_id, name):
    """Download protein structure from AlphaFold."""
    url = f"https://alphafold.ebi.ac.uk/files/AF-{uniprot_id}-F1-model_v4.pdb"
    output = f"{name}.pdb"
    
    if os.path.exists(output):
        print(f"  Already exists: {output}")
        return output
    
    result = subprocess.run(
        ["curl", "-s", "-f", url, "-o", output],
        capture_output=True
    )
    
    if result.returncode != 0:
        # Try v6 instead
        url = f"https://alphafold.ebi.ac.uk/files/AF-{uniprot_id}-F1-model_v6.pdb"
        result = subprocess.run(
            ["curl", "-s", "-f", url, "-o", output],
            capture_output=True
        )
    
    if result.returncode == 0:
        print(f"  Downloaded: {output}")
        return output
    else:
        print(f"  Failed to download {uniprot_id}")
        return None

def main():
    print("Fetching medically important proteins from AlphaFold...\n")
    
    downloaded = []
    for uniprot_id, name, description in PROTEINS:
        print(f"{name}: {description}")
        pdb_file = fetch_protein(uniprot_id, name)
        if pdb_file:
            downloaded.append((pdb_file, name, description))
    
    print(f"\n✓ Downloaded {len(downloaded)} proteins")
    
    # Create jobs for each
    print("\nCreating NEXUS mining jobs...\n")
    
    jobs = []
    for pdb_file, name, description in downloaded:
        result = subprocess.run(
            ["python3", "extract_protein_problem.py", pdb_file, "128"],
            capture_output=True,
            text=True
        )
        if result.returncode == 0:
            job_file = pdb_file.replace('.pdb', '_job.json')
            if os.path.exists(job_file):
                with open(job_file) as f:
                    job = json.load(f)
                job['description'] = description
                job['uniprot_id'] = pdb_file.split('_')[0].upper()
                jobs.append(job)
                print(f"  ✓ {name}: {job['num_contacts']} contacts, threshold={job['threshold']}")
    
    # Save job queue
    with open('protein_job_queue.json', 'w') as f:
        json.dump(jobs, f, indent=2)
    
    print(f"\n✓ Created protein_job_queue.json with {len(jobs)} jobs")

if __name__ == '__main__':
    main()

#!/usr/bin/env python3
"""
Extract protein structure and convert to NEXUS mining optimization problem.

The problem is formulated as minimizing the energy of a coarse-grained 
protein model where:
- Nodes = C-alpha atoms (backbone)
- Edges = contacts between residues (distance < 8 Angstroms)
- Energy = sum of (distance - ideal_distance)^2 for contacts

Miners optimize spin configurations that represent conformational states.
"""

import sys
import json
import hashlib
import base64
import struct
from math import sqrt

def parse_pdb(filename):
    """Extract C-alpha coordinates from PDB file."""
    atoms = []
    with open(filename, 'r') as f:
        for line in f:
            if line.startswith('ATOM') and line[12:16].strip() == 'CA':
                x = float(line[30:38])
                y = float(line[38:46])
                z = float(line[46:54])
                residue = line[17:20].strip()
                res_num = int(line[22:26])
                atoms.append({
                    'residue': residue,
                    'res_num': res_num,
                    'x': x,
                    'y': y,
                    'z': z
                })
    return atoms

def compute_distance(a1, a2):
    """Euclidean distance between two atoms."""
    dx = a1['x'] - a2['x']
    dy = a1['y'] - a2['y']
    dz = a1['z'] - a2['z']
    return sqrt(dx*dx + dy*dy + dz*dz)

def create_contact_map(atoms, cutoff=8.0, min_seq_sep=3):
    """
    Create contact map from protein structure.
    Only includes contacts between residues separated by at least min_seq_sep.
    """
    n = len(atoms)
    contacts = []
    
    for i in range(n):
        for j in range(i + min_seq_sep, n):
            dist = compute_distance(atoms[i], atoms[j])
            if dist < cutoff:
                contacts.append({
                    'i': i,
                    'j': j,
                    'distance': dist
                })
    
    return contacts

def create_ising_problem(atoms, contacts, problem_size=None):
    """
    Convert protein contact map to Ising model problem.
    
    The Ising model energy is:
    E = -sum_ij J_ij * s_i * s_j - sum_i h_i * s_i
    
    We encode:
    - J_ij = contact strength (inverse of distance)
    - h_i = local field (bias towards native contacts)
    """
    n = len(atoms)
    
    # Limit problem size for tractability
    if problem_size and n > problem_size:
        # Take a subset (e.g., a domain)
        atoms = atoms[:problem_size]
        n = problem_size
        contacts = [c for c in contacts if c['i'] < n and c['j'] < n]
    
    # Create coupling matrix J (sparse representation)
    J = {}
    for c in contacts:
        # Coupling strength inversely proportional to distance
        # Negative J favors aligned spins (native contact)
        strength = -1.0 / (c['distance'] + 0.1)
        J[(c['i'], c['j'])] = strength
    
    # Create local field h (slight bias)
    h = [0.0] * n
    
    # Pack into binary format for NEXUS
    # Format: [n (4 bytes)][num_couplings (4 bytes)][couplings...][fields...]
    data = struct.pack('<I', n)  # problem size
    data += struct.pack('<I', len(J))  # number of couplings
    
    for (i, j), strength in J.items():
        data += struct.pack('<HHf', i, j, strength)  # i, j as uint16, strength as float32
    
    for field in h:
        data += struct.pack('<f', field)  # local fields as float32
    
    return data, n, len(J)

def main():
    if len(sys.argv) < 2:
        print("Usage: python extract_protein_problem.py <pdb_file> [max_residues]")
        sys.exit(1)
    
    pdb_file = sys.argv[1]
    max_residues = int(sys.argv[2]) if len(sys.argv) > 2 else 100
    
    print(f"Processing: {pdb_file}")
    print(f"Max residues: {max_residues}")
    
    # Parse PDB
    atoms = parse_pdb(pdb_file)
    print(f"Found {len(atoms)} C-alpha atoms")
    
    # Create contact map
    contacts = create_contact_map(atoms)
    print(f"Found {len(contacts)} contacts (cutoff=8Å, seq_sep>=3)")
    
    # Create Ising problem
    problem_data, n, num_couplings = create_ising_problem(atoms, contacts, max_residues)
    
    # Encode as base64 for JSON
    problem_b64 = base64.b64encode(problem_data).decode('ascii')
    
    # Compute hash
    problem_hash = hashlib.sha256(problem_data).hexdigest()
    
    # Create NEXUS job specification
    job = {
        "problem_type": "protein_folding",
        "source": "alphafold",
        "pdb_file": pdb_file,
        "protein_name": pdb_file.replace('.pdb', ''),
        "num_residues": min(len(atoms), max_residues),
        "num_contacts": num_couplings,
        "problem_size": n,
        "problem_data": problem_b64,
        "problem_hash": problem_hash,
        "threshold": -int(num_couplings * 0.8),  # 80% of max possible energy reduction
        "metadata": {
            "total_residues": len(atoms),
            "total_contacts": len(contacts),
            "cutoff_angstroms": 8.0,
            "min_sequence_separation": 3
        }
    }
    
    # Output JSON
    output_file = pdb_file.replace('.pdb', '_job.json')
    with open(output_file, 'w') as f:
        json.dump(job, f, indent=2)
    
    print(f"\nJob created: {output_file}")
    print(f"Problem size: {n} spins")
    print(f"Couplings: {num_couplings}")
    print(f"Threshold: {job['threshold']}")
    print(f"Hash: {problem_hash[:16]}...")

if __name__ == '__main__':
    main()

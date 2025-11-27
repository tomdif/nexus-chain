# NEXUS API Schema

## Overview

The NEXUS API provides REST and GraphQL interfaces for labs and researchers to:
- Submit bounty jobs
- Query results
- Download datasets
- Verify proofs

Base URL: `https://api.nexus.network/v1`

---

## REST Endpoints

### Jobs

#### List Jobs
```
GET /jobs
```

Query Parameters:
| Parameter | Type | Description |
|-----------|------|-------------|
| status | string | Filter by status |
| job_type | string | Filter by type |
| protein_id | string | Filter by UniProt ID |
| page | int | Page number |
| limit | int | Results per page |

#### Get Job
```
GET /jobs/{job_id}
```

#### Create Bounty Job
```
POST /jobs
```

Request Body:
```json
{
  "job_type": "protein_stability",
  "protein_id": "P00533",
  "pdb_reference": "4HJO",
  "mutations": ["L858R", "T790M"],
  "budget": "50000000000unexus",
  "mev_fee": "1000000000",
  "min_miners": 100,
  "deadline_hours": 24,
  "public_results": true
}
```

---

### Proteins

#### Get Protein
```
GET /proteins/{uniprot_id}
```

Response:
```json
{
  "uniprot_id": "P00533",
  "gene_name": "EGFR",
  "protein_name": "Epidermal growth factor receptor",
  "organism": "Homo sapiens",
  "length": 1210,
  "mutations_computed": 21847,
  "completeness_percent": 95.0,
  "landscape_available": true
}
```

#### Get Mutation
```
GET /proteins/{uniprot_id}/mutations/{mutation}
```

Example: `GET /proteins/P00533/mutations/L858R`

Response:
```json
{
  "protein_id": "P00533",
  "mutation": "L858R",
  "hgvs_notation": "p.Leu858Arg",
  "position": 858,
  "result": {
    "ddg_kcal_mol": -2.31,
    "ddg_std": 0.15,
    "confidence": 0.95,
    "classification": "stabilizing"
  },
  "computation": {
    "job_id": "job_2025_12345",
    "num_miners": 284,
    "total_steps": 28400000
  },
  "verification": {
    "chain": "nexus-mainnet",
    "block_height": 1250500,
    "all_proofs_valid": true
  }
}
```

#### Get Landscape
```
GET /proteins/{uniprot_id}/landscape?format=csv
```

---

### Verification

#### Verify Proof
```
GET /verify/{proof_hash}
```

Response:
```json
{
  "proof_hash": "0x...",
  "valid": true,
  "job_id": "job_2025_12345",
  "miner": "nexus1abc...",
  "steps_proven": 100000,
  "block_height": 1250000
}
```

---

### Network Stats
```
GET /stats/network
```

Response:
```json
{
  "total_miners": 12543,
  "active_miners_24h": 8721,
  "total_jobs_completed": 1543876,
  "proteins_with_data": 18432,
  "mutations_computed": 349821543
}
```

---

## GraphQL Schema
```graphql
type Query {
  job(id: ID!): Job
  jobs(status: JobStatus, jobType: JobType, first: Int): JobConnection!
  protein(uniprotId: String!): Protein
  proteins(organism: String, hasLandscape: Boolean): ProteinConnection!
  mutation(proteinId: String!, mutation: String!): MutationResult
  verifyProof(proofHash: String!): ProofVerification!
  networkStats: NetworkStats!
}

type Job {
  id: ID!
  jobType: JobType!
  source: JobSource!
  status: JobStatus!
  protein: Protein
  mutations: [String!]
  budget: String!
  mevFee: String!
  minersParticipated: Int!
  stepsCompleted: BigInt!
  results: JobResults
  createdAt: DateTime!
  completedAt: DateTime
}

type Protein {
  uniprotId: String!
  geneName: String!
  proteinName: String!
  organism: String!
  sequence: String!
  length: Int!
  landscape: Landscape
  mutationsComputed: Int!
  completenessPercent: Float!
}

type MutationResult {
  protein: Protein!
  mutation: String!
  hgvsNotation: String!
  position: Int!
  ddgKcalMol: Float!
  ddgStd: Float!
  confidence: Float!
  classification: MutationClassification!
  numMiners: Int!
  proofHashes: [String!]!
  verifiedOnChain: Boolean!
}

type Landscape {
  proteinId: String!
  version: Int!
  completenessPercent: Float!
  matrixIpfsCid: String!
  csvDownloadUrl: String!
  heatmapUrl: String!
  hotspots: [Hotspot!]!
}

enum JobType {
  ISING_OPTIMIZATION
  PROTEIN_STABILITY
  MOLECULAR_DOCKING
  MUTATIONAL_SCAN
}

enum JobStatus {
  PENDING
  ACTIVE
  AGGREGATING
  COMPLETED
  EXPIRED
}

enum MutationClassification {
  STABILIZING
  NEUTRAL
  DESTABILIZING
  HIGHLY_DESTABILIZING
}
```

---

## Authentication

For bounty job submission:
```
Authorization: Bearer <api_key>
```

## Rate Limits

| Endpoint Type | Rate Limit |
|---------------|------------|
| Read (GET) | 1000/hour |
| Write (POST) | 100/hour |
| Export | 10/hour |

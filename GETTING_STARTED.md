# NEXUS Chain - Complete Beginner's Guide

This guide assumes you're starting from scratch on Ubuntu/WSL. Every command is copy-paste ready.

---

## Part 1: Install Prerequisites

### Step 1.1: Update Your System

Open your terminal and run:

```bash
sudo apt update && sudo apt upgrade -y
```

Wait for it to complete (may take a few minutes).

---

### Step 1.2: Install Basic Tools

```bash
sudo apt install -y build-essential git curl wget
```

---

### Step 1.3: Install Go 1.22

Cosmos SDK requires Go. Run these commands ONE AT A TIME:

```bash
# Download Go
cd ~
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
```

```bash
# Remove any old Go installation and extract new one
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
```

```bash
# Add Go to your PATH (this makes 'go' command available)
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
source ~/.bashrc
```

```bash
# Verify Go is installed
go version
```

You should see: `go version go1.22.0 linux/amd64`

If you don't see this, close your terminal and open a new one, then try `go version` again.

---

### Step 1.4: Create Go Workspace

```bash
mkdir -p ~/go/src ~/go/bin ~/go/pkg
```

---

## Part 2: Get the NEXUS Chain Code

### Step 2.1: Create Project Directory

```bash
mkdir -p ~/projects
cd ~/projects
```

---

### Step 2.2: Download the Code

**Option A: If you downloaded the tarball from Claude**

1. The file `nexus-chain.tar.gz` should be in your Downloads folder
2. Move it to your projects folder:

```bash
# If on WSL, your Windows downloads are usually at:
cp /mnt/c/Users/YOUR_USERNAME/Downloads/nexus-chain.tar.gz ~/projects/

# Replace YOUR_USERNAME with your actual Windows username
# For example: cp /mnt/c/Users/Tom/Downloads/nexus-chain.tar.gz ~/projects/
```

3. Extract it:

```bash
cd ~/projects
tar -xzvf nexus-chain.tar.gz
cd nexus-chain
```

**Option B: If Option A doesn't work**

Tell me and I'll help you get the files another way.

---

### Step 2.3: Verify You're in the Right Place

```bash
pwd
```

Should show: `/home/YOUR_USERNAME/projects/nexus-chain`

```bash
ls -la
```

Should show files like: `Makefile`, `go.mod`, `README.md`, and folders like `app/`, `x/`, `cmd/`

---

## Part 3: Build the Chain

### Step 3.1: Download Dependencies

This downloads all the libraries the code needs:

```bash
cd ~/projects/nexus-chain
go mod tidy
```

This will take 2-5 minutes and download a lot of packages. You'll see output scrolling by - that's normal.

If you see errors about "module not found", that's okay for now - we'll fix them.

---

### Step 3.2: Try Building

```bash
make build
```

**Expected: You will likely see errors.** That's because the code I generated needs some fixes to compile. This is normal in development.

---

## Part 4: Fix Compilation Errors

The code needs some adjustments. Let me walk you through fixing them.

### Step 4.1: Check What Errors You See

Run:

```bash
go build ./... 2>&1 | head -50
```

This shows the first 50 lines of errors. 

**Copy and paste the error output back to me** and I'll tell you exactly what to fix.

---

## Part 5: What Happens After It Builds

Once we fix the errors and it builds successfully, here's what you'll do:

### Step 5.1: Initialize a Local Testnet

```bash
chmod +x scripts/init.sh
./scripts/init.sh
```

This creates:
- A validator key (like a wallet)
- Genesis file (the starting state of the blockchain)
- Configuration files

### Step 5.2: Start the Chain

```bash
./build/nexusd start
```

You'll see blocks being produced every 2 seconds:
```
INF committed state height=1 ...
INF committed state height=2 ...
INF committed state height=3 ...
```

Press `Ctrl+C` to stop.

### Step 5.3: Interact with the Chain

In a NEW terminal (keep the chain running in the first one):

```bash
# Check chain status
./build/nexusd status

# Check your validator account balance
./build/nexusd query bank balances $(./build/nexusd keys show validator -a --keyring-backend test)
```

---

## Part 6: Understanding the Project Structure

```
nexus-chain/
│
├── app/                      # The main application
│   └── app.go               # Wires all modules together
│
├── cmd/nexusd/              # The binary (executable)
│   └── main.go              # Entry point
│
├── x/mining/                # YOUR CUSTOM MODULE (the interesting part!)
│   │
│   ├── keeper/              # Business logic
│   │   ├── keeper.go        # Core functions (PostJob, SubmitProof, etc.)
│   │   ├── abci.go          # Block-level logic (checkpoints)
│   │   └── msg_server.go    # Handles transactions
│   │
│   ├── types/               # Data structures
│   │   ├── types.go         # Job, Proof, Checkpoint definitions
│   │   ├── params.go        # Economic parameters (80/20 split, etc.)
│   │   └── msgs.go          # Transaction message types
│   │
│   └── module.go            # Registers module with Cosmos SDK
│
├── scripts/
│   └── init.sh              # Initializes local testnet
│
├── go.mod                   # Dependencies
└── Makefile                 # Build commands
```

---

## Part 7: Key Concepts

### What is a Cosmos SDK Module?

Think of it like a plugin. Cosmos SDK provides the base blockchain (accounts, tokens, staking), and you add custom modules for your specific functionality.

Your `x/mining` module adds:
- Job posting (customers submit optimization problems)
- Proof submission (miners submit ZK proofs)
- Share calculation (Universal Share Formula)
- Checkpoints (every ~10 minutes)

### What is a Keeper?

The Keeper is where the business logic lives. It's like the "backend" of your module. It reads and writes to the blockchain's database.

### What is a Message?

A Message (Msg) is a transaction type. When someone wants to do something (post a job, submit a proof), they send a Message.

### What are ABCI Hooks?

ABCI = Application Blockchain Interface. BeginBlocker and EndBlocker run at the start and end of every block. This is where checkpoints get created.

---

## Part 8: The Mining Module Flow

```
1. Customer posts job (MsgPostJob)
   └── Reward held in escrow

2. Miners solve problem
   └── Generate ZK proof with Nova prover

3. Miner submits proof (MsgSubmitProof)
   └── VerifyProof() checks ZK proof (currently placeholder)
   └── Shares calculated: max(0, previous_best - new_energy)
   └── Shares recorded

4. Job expires or threshold met
   └── EndBlocker marks job complete

5. Miners claim rewards (MsgClaimRewards)
   └── Proportional to shares earned
   └── 80% to miners, 20% to validators
```

---

## Next Steps After Building

1. **Get it building** - Fix any compilation errors
2. **Run local testnet** - See blocks being produced
3. **Understand the code** - Read through keeper.go
4. **Connect your ZK prover** - Replace placeholder VerifyProof()
5. **Add CLI commands** - Make it user-friendly
6. **Write tests** - Ensure everything works
7. **Deploy testnet** - Multiple nodes

---

## Common Issues

### "go: command not found"
Close terminal and open new one. If still broken:
```bash
export PATH=$PATH:/usr/local/go/bin
```

### "permission denied"
Add `sudo` before the command, or:
```bash
chmod +x <filename>
```

### "module not found" errors
```bash
go mod tidy
```

### Build takes forever
Normal for first build. Subsequent builds are faster.

---

## Getting Help

When you hit an error:
1. Copy the FULL error message
2. Note what command you ran
3. Paste both to me

I'll tell you exactly what to fix.

---

Ready? Start at Part 1, Step 1.1 and work through sequentially. Let me know when you hit any issues!

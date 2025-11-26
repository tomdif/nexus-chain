# NEXUS Validator Guide

## Hardware Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| CPU | 4 cores | 8 cores |
| RAM | 8 GB | 16 GB |
| Storage | 100 GB SSD | 500 GB NVMe |
| Network | 100 Mbps | 1 Gbps |

## Setup

### 1. Build from Source
```bash
git clone https://github.com/tomdif/nexus-chain.git
cd nexus-chain
go build -o nexusd ./cmd/nexusd/
sudo mv nexusd /usr/local/bin/
```

### 2. Initialize Node
```bash
nexusd init <your-moniker>
```

This creates:
- `~/.nexus/config/config.toml` - Node configuration
- `~/.nexus/config/genesis.json` - Chain genesis
- `~/.nexus/config/priv_validator_key.json` - Validator key (BACKUP THIS!)
- `~/.nexus/config/node_key.json` - P2P identity

### 3. Configure

Edit `~/.nexus/config/config.toml`:
```toml
# Moniker (your validator name)
moniker = "my-awesome-validator"

# P2P settings
[p2p]
laddr = "tcp://0.0.0.0:26656"
persistent_peers = "<node-id>@<ip>:26656,..."

# RPC settings  
[rpc]
laddr = "tcp://127.0.0.1:26657"
```

### 4. Get Genesis

For testnet:
```bash
curl -o ~/.nexus/config/genesis.json https://raw.githubusercontent.com/tomdif/nexus-chain/main/genesis.json
```

### 5. Start Node
```bash
nexusd start
```

Or with systemd:
```bash
sudo tee /etc/systemd/system/nexusd.service > /dev/null <<SERVICE
[Unit]
Description=NEXUS Node
After=network.target

[Service]
Type=simple
User=$USER
ExecStart=/usr/local/bin/nexusd start
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
SERVICE

sudo systemctl enable nexusd
sudo systemctl start nexusd
```

## Monitoring

### Check Sync Status
```bash
curl -s localhost:26657/status | jq '.result.sync_info'
```

### View Logs
```bash
journalctl -u nexusd -f
```

### Check Validator Status
```bash
nexusd query staking validator <validator-address>
```

## Security

### Firewall
```bash
# Allow P2P
sudo ufw allow 26656/tcp

# Restrict RPC to localhost only
sudo ufw deny 26657/tcp
```

### Key Backup

**Critical files to backup:**
- `~/.nexus/config/priv_validator_key.json`
- `~/.nexus/config/node_key.json`
- Your mnemonic phrase

### Sentry Node Architecture

For production, run sentry nodes:
```
Internet
    │
    ▼
┌─────────┐     ┌─────────┐
│ Sentry 1│◄───►│ Sentry 2│
└────┬────┘     └────┬────┘
     │               │
     └───────┬───────┘
             │
             ▼
      ┌─────────────┐
      │  Validator  │
      │  (private)  │
      └─────────────┘
```

## Troubleshooting

### Node Won't Sync

1. Check peers: `curl localhost:26657/net_info`
2. Reset state: `nexusd tendermint unsafe-reset-all`
3. Re-download genesis

### Out of Memory

Increase swap or reduce `max-open-connections` in config.

### Missed Blocks

Check system time is synced:
```bash
timedatectl status
```

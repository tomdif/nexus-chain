#!/bin/bash
# Generate genesis for NEXUS testnet

set -e

CHAIN_ID="${CHAIN_ID:-nexus-testnet-1}"
HOME_DIR="${HOME_DIR:-$HOME/.nexus}"
MONIKER="${1:-nexus-validator}"

echo "=== NEXUS Genesis Generator ==="
echo "Chain ID: $CHAIN_ID"
echo "Home: $HOME_DIR"
echo "Moniker: $MONIKER"
echo ""

# Initialize
./nexusd init $MONIKER --home $HOME_DIR

echo ""
echo "Genesis generated at: $HOME_DIR/config/genesis.json"
echo ""
echo "=== NEXUS Tokenomics ==="
echo ""
echo "Token: NEX (1 NEX = 1,000,000 unexus)"
echo ""
echo "Initial Allocation (25B NEX):"
echo "  - Founders/Team:     10B NEX (10%)"
echo "  - Treasury:          10B NEX (10%)"
echo "  - Initial Validators: 5B NEX (5%)"
echo ""
echo "Mining Emissions (75B NEX over 20+ years):"
echo "  Year 1-2:   35,950 NEX/min = 18.9B/year (100%)"
echo "  Year 3-4:   17,975 NEX/min =  9.4B/year (50%)"
echo "  Year 5-6:    8,988 NEX/min =  4.7B/year (25%)"
echo "  Year 7-8:    4,494 NEX/min =  2.4B/year (12.5%)"
echo "  Year 9-10:   2,229 NEX/min =  1.2B/year (6.2%)"
echo "  Year 11-12:  1,114 NEX/min =  590M/year (3.1%)"
echo "  Year 13-14:    539 NEX/min =  283M/year (1.5%)"
echo "  Year 15+:      539 NEX/min =  283M/year (perpetual)"
echo ""
echo "Deflationary Mechanisms:"
echo "  - 2% job posting fee burned"
echo "  - 50% transaction fees burned"
echo "  - Priority fees burned"
echo ""
echo "To start: ./nexusd start --home $HOME_DIR"

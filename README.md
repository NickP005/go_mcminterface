# MCMInterface v1.2.1

A production-ready Go library for interfacing with the Mochimo Network through the native socket/TCP protocol.
Written by [NickP005](https://github.com/NickP005)

## Installation

```bash
go get github.com/NickP005/go_mcminterface
```

## Features

- **Node Discovery**: Automatically find and connect to Mochimo network nodes
- **Consensus Mechanism**: Query multiple nodes to ensure data integrity
- **Balance Queries**: Look up address balances across the network
- **Transaction Management**: Create and submit transactions to the network
- **Block Information**: Access block data, trailers, and chain history
- **Address Utilities**: Work with Mochimo addresses including WOTS+ conversions
- **Performance Optimization**: Intelligent node selection based on response time
- **Ledger Operations**: Load, save and manipulate ledger data

## Configuration

### Settings Structure

The library uses a central settings structure for configuration:

```go
type SettingsType struct {
    StartIPs           []string     // Initial seed nodes
    IPs                []string     // Known network nodes
    Nodes              []RemoteNode // Benchmarked nodes with performance data
    IPExpandDepth      int          // How many levels deep to search for new nodes
    ForceQueryStartIPs bool         // Always include seed nodes in queries
    QuerySize          int          // Number of nodes to query for consensus
    QueryTimeout       int          // Single query timeout in seconds
    MaxQueryAttempts   int          // Maximum number of retries per query
    DefaultPort        int          // Default port for node connections
}
```

Example `settings.json`:
```json
{
    "StartIPs": ["seed1.mochimo.org", "seed2.mochimo.org"],
    "IPs": [],
    "Nodes": [],
    "IPExpandDepth": 3,
    "ForceQueryStartIPs": false,
    "QuerySize": 5,
    "QueryTimeout": 5,
    "MaxQueryAttempts": 3,
    "DefaultPort": 2095
}
```

## Core API Reference

### Network Management

#### Loading and Saving Settings

```go
// Load settings from a JSON file
func LoadSettings(path string) SettingsType

// Load default settings (embedded in the library)
func LoadDefaultSettings()

// Save current settings to a file
func SaveSettings(settings SettingsType)
```

#### Node Discovery and Benchmarking

```go
// Discover new nodes by querying known nodes for their peers
func ExpandIPs()

// Benchmark nodes to determine connection quality
func BenchmarkNodes(concurrent int)

// Pick best nodes for querying based on response time
func PickNodes(count int) []RemoteNode

// Connect to a specific node
func ConnectToNode(ip string) SocketData
```

### Balance and Address Operations

```go
// Query balance of an address with consensus verification
func QueryBalance(wots_address string) (uint64, error)

// Resolve a tag to a full address
func QueryTagResolve(tag []byte) (WotsAddress, error)
func QueryTagResolveHex(tag_hex string) (WotsAddress, error)

// Convert between address formats
func WotsAddressFromBytes(bytes []byte) WotsAddress
func WotsAddressFromHex(wots_hex string) WotsAddress
func AddrFromWots(wots []byte) []byte
func AddrFromImplicit(tag []byte) []byte
```

### Block and Blockchain Operations

```go
// Query the latest block number with consensus
func QueryLatestBlockNumber() (uint64, error)

// Query block hash with consensus
func QueryBlockHash(block_num uint64) ([HASHLEN]byte, error)

// Get a specific block's data with consensus verification
func QueryBlockBytes(block_num uint64) ([]byte, error)
func QueryBlockFromNumber(block_num uint64) (Block, error)

// Get block trailers for a range of blocks
func QueryBTrailers(start_block uint32, count uint32) ([]BTRAILER, error)

// Submit a transaction to the network
func SubmitTransaction(tx TXENTRY) error
```

### Ledger Operations

```go
// Load a ledger from file
func LoadLedgerFromFile(filepath string) (*Ledger, error)

// Save a ledger to file
func SaveLedgerToFile(ledger *Ledger, filepath string) error

// Get a partition of the ledger
func (l *Ledger) GetLedgerPartition(startIndex, endIndex uint64) (*Ledger, error)
```

## Ledger Management

The `Ledger` type provides ways to store and manipulate blockchain ledger data:

```go
type Ledger struct {
    Size            uint64
    Entries         []LedgerEntry
    IsBalanceSorted bool
    IsAddressSorted bool
}
```

### Ledger Methods

```go
// Add entry to the ledger (maintains sorting if enabled)
func (l *Ledger) AddEntry(entry LedgerEntry)

// Sort ledger entries
func (l *Ledger) SortBalances()
func (l *Ledger) SortAddresses()

// Search functions
func (l *Ledger) SearchByAddressPrefix(prefix []byte) []LedgerEntry
func (l *Ledger) SearchByAddress(address []byte) *LedgerEntry

// Get ledger partition
func (l *Ledger) GetLedgerPartition(startIndex, endIndex uint64) (*Ledger, error)
```

## Transaction Building

### Creating and Managing Transactions

```go
// Create a new transaction
tx := NewTXENTRY()

// Set transaction properties
tx.SetSourceAddress(srcAddr)
tx.SetChangeAddress(changeAddr)
tx.SetSendTotal(amount)
tx.SetFee(fee)
tx.SetBlockToLive(blockNum + 1000)  // Expiration

// Add destination
dst := NewDSTFromString(tagHex, reference, amount)
tx.AddDestination(dst)

// Submit transaction
err := SubmitTransaction(tx)
```

## Examples

### Basic Network Connection

```go
package main

import (
    "fmt"
    "github.com/NickP005/go_mcminterface"
)

func main() {
    // Initialize network connection
    go_mcminterface.LoadSettings("settings.json")
    go_mcminterface.ExpandIPs()
    go_mcminterface.BenchmarkNodes(5)
    
    // Save discovered nodes for future use
    go_mcminterface.SaveSettings(go_mcminterface.Settings)
}
```

### Querying Balance

```go
package main

import (
    "fmt"
    "github.com/NickP005/go_mcminterface"
)

func main() {
    go_mcminterface.LoadSettings("settings.json")
    
    // Query balance by address
    address := "0f8213c50de73ee326009d6a1475d1dba1105777" // Example address
    balance, err := go_mcminterface.QueryBalance(address)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    
    // Convert from nanoMCM to MCM
    mcmBalance := float64(balance) / 1000000000
    fmt.Printf("Balance: %.9f MCM\n", mcmBalance)
}
```

### Working with Blocks

```go
package main

import (
    "encoding/hex"
    "fmt"
    "github.com/NickP005/go_mcminterface"
)

func main() {
    go_mcminterface.LoadSettings("settings.json")
    
    // Get latest block number
    blockNum, err := go_mcminterface.QueryLatestBlockNumber()
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Printf("Latest block: %d\n", blockNum)
    
    // Get block data
    block, err := go_mcminterface.QueryBlockFromNumber(blockNum)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    
    // Display block information
    fmt.Printf("Block hash: %x\n", block.Trailer.Bhash)
    fmt.Printf("Block transactions: %d\n", len(block.Body))
}
```

### Ledger Operations

```go
package main

import (
    "fmt"
    "github.com/NickP005/go_mcminterface"
)

func main() {
    // Load ledger from file
    ledger, err := go_mcminterface.LoadLedgerFromFile("ledger.dat")
    if err != nil {
        fmt.Println("Error loading ledger:", err)
        return
    }
    
    // Sort ledger by address for efficient searching
    ledger.SortAddresses()
    
    // Search for address
    prefix := []byte{0x01, 0x23, 0x45} // Example address prefix
    results := ledger.SearchByAddressPrefix(prefix)
    fmt.Printf("Found %d addresses starting with prefix\n", len(results))
    
    // Get a partition of the ledger
    partition, err := ledger.GetLedgerPartition(1000, 2000)
    if err != nil {
        fmt.Println("Error getting partition:", err)
        return
    }
    fmt.Printf("Partition contains %d entries\n", partition.Size)
    
    // Save modified ledger
    err = go_mcminterface.SaveLedgerToFile(ledger, "ledger_sorted.dat")
    if err != nil {
        fmt.Println("Error saving ledger:", err)
    }
}
```

## Low-Level Socket API

For advanced users, the library provides direct access to the Mochimo network protocol:

```go
// Socket operations
func (m *SocketData) Connect()
func (m *SocketData) SendOP(op uint16) error
func (m *SocketData) Hello() error

// Direct network queries
func (m *SocketData) GetIPList() ([]string, error)
func (m *SocketData) ResolveTag(tag []byte) (WotsAddress, error)
func (m *SocketData) GetBalance(wots_addr WotsAddress) (uint64, error)
func (m *SocketData) GetBlockBytes(block_num uint64) ([]byte, error)
func (m *SocketData) GetBlockHash(block_num uint64) ([HASHLEN]byte, error)
func (m *SocketData) GetTrailersBytes(block_num uint32, count uint32) ([]byte, error)
func (m *SocketData) SubmitTransaction(tx TXENTRY) error
```

## Implementation Notes

- All query functions use a consensus mechanism requiring >50% agreement among nodes
- Node selection is weighted by ping performance for optimal reliability
- Built-in retry mechanisms handle temporary network failures
- Sorting enables efficient binary searches when working with large ledgers
- Address operations support both legacy WOTS+ addresses and hash-based addresses
- Documentation reflects version 1.2.0 feature set

## Support & Community

Join our communities for support and discussions:

<div align="center">

[![NickP005 Development Server](https://img.shields.io/discord/709417966881472572?color=7289da&label=NickP005%20Development%20Server&logo=discord&logoColor=white)](https://discord.gg/Q5jM8HJhNT)   
[![Mochimo Official](https://img.shields.io/discord/460867662977695765?color=7289da&label=Mochimo%20Official&logo=discord&logoColor=white)](https://discord.gg/SvdXdr2j3Y)

</div>

- **NickP005 Development Server**: Technical support and development discussions
- **Mochimo Official**: General Mochimo blockchain discussions and community
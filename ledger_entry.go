package go_mcminterface

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"
)

type Ledger struct {
	Size            uint64
	Entries         []LedgerEntry
	IsBalanceSorted bool
	IsAddressSorted bool
}

func (l *Ledger) GetSize() uint64 {
	return l.Size
}

func (l *Ledger) GetEntries() []LedgerEntry {
	return l.Entries
}

func (l *Ledger) AddEntry(entry LedgerEntry) {
	if l.IsBalanceSorted {
		// Find the position to insert while maintaining balance sorting
		idx := sort.Search(len(l.Entries), func(i int) bool {
			return l.Entries[i].Balance >= entry.Balance
		})
		// Insert entry at the correct position
		l.Entries = append(l.Entries, LedgerEntry{})
		copy(l.Entries[idx+1:], l.Entries[idx:])
		l.Entries[idx] = entry
	} else if l.IsAddressSorted {
		// Find the position to insert while maintaining address sorting
		idx := sort.Search(len(l.Entries), func(i int) bool {
			return bytes.Compare(l.Entries[i].Address[:], entry.Address[:]) >= 0
		})
		// Insert entry at the correct position
		l.Entries = append(l.Entries, LedgerEntry{})
		copy(l.Entries[idx+1:], l.Entries[idx:])
		l.Entries[idx] = entry
	} else {
		// Simply append if not sorted
		l.Entries = append(l.Entries, entry)
	}
	l.Size++
}

// SortBalances sorts the ledger entries by balance in ascending order
func (l *Ledger) SortBalances() {
	sort.Slice(l.Entries, func(i, j int) bool {
		return l.Entries[i].Balance < l.Entries[j].Balance
	})
	l.IsBalanceSorted = true
	l.IsAddressSorted = false
}

// SortAddresses sorts the ledger entries by address in ascending order
func (l *Ledger) SortAddresses() {
	sort.Slice(l.Entries, func(i, j int) bool {
		return bytes.Compare(l.Entries[i].Address[:], l.Entries[j].Address[:]) < 0
	})
	l.IsAddressSorted = true
	l.IsBalanceSorted = false
}

// SearchByAddressPrefix searches for entries with an address starting with the given prefix
// prefix can be any length from 1 to ADDR_LEN bytes
// Returns a new Ledger containing only the matching entries
func (l *Ledger) SearchByAddressPrefix(prefix []byte) *Ledger {
	if len(prefix) == 0 || len(prefix) > ADDR_LEN {
		return &Ledger{Size: 0, Entries: []LedgerEntry{}, IsBalanceSorted: l.IsBalanceSorted, IsAddressSorted: l.IsAddressSorted}
	}

	result := &Ledger{
		IsBalanceSorted: l.IsBalanceSorted,
		IsAddressSorted: l.IsAddressSorted,
	}

	// If address sorted, we can do a binary search to find the first match
	if l.IsAddressSorted {
		// Find the first potential match
		idx := sort.Search(len(l.Entries), func(i int) bool {
			return bytes.Compare(l.Entries[i].Address[:len(prefix)], prefix) >= 0
		})

		// Collect all matches
		for i := idx; i < len(l.Entries); i++ {
			if bytes.HasPrefix(l.Entries[i].Address[:], prefix) {
				result.Entries = append(result.Entries, l.Entries[i])
			} else if bytes.Compare(l.Entries[i].Address[:len(prefix)], prefix) > 0 {
				// We've moved past potential matches
				break
			}
		}
	} else {
		// Linear search if not sorted
		for _, entry := range l.Entries {
			if bytes.HasPrefix(entry.Address[:], prefix) {
				result.Entries = append(result.Entries, entry)
			}
		}
	}

	result.Size = uint64(len(result.Entries))
	return result
}

// FilterBy returns a new Ledger with entries that meet the balance criteria
// minBalance is the minimum balance required
// maxBalance is the maximum balance allowed (-1 means no upper limit)
func (l *Ledger) FilterBy(minBalance, maxBalance uint64) *Ledger {
	result := &Ledger{
		IsBalanceSorted: l.IsBalanceSorted,
		IsAddressSorted: l.IsAddressSorted,
	}

	noUpperLimit := maxBalance == ^uint64(0) // Using max uint64 as the indicator for "no upper limit"

	// If balance sorted, we can optimize the filtering
	if l.IsBalanceSorted {
		// Find the first entry with balance >= minBalance
		startIdx := sort.Search(len(l.Entries), func(i int) bool {
			return l.Entries[i].Balance >= minBalance
		})

		// If no upper limit, include all entries from startIdx
		if noUpperLimit {
			result.Entries = append(result.Entries, l.Entries[startIdx:]...)
		} else {
			// Find the first entry with balance > maxBalance
			endIdx := sort.Search(len(l.Entries), func(i int) bool {
				return l.Entries[i].Balance > maxBalance
			})

			// Include all entries from startIdx to endIdx-1
			result.Entries = append(result.Entries, l.Entries[startIdx:endIdx]...)
		}
	} else {
		// Linear search if not sorted by balance
		for _, entry := range l.Entries {
			if entry.Balance >= minBalance && (noUpperLimit || entry.Balance <= maxBalance) {
				result.Entries = append(result.Entries, entry)
			}
		}
	}

	result.Size = uint64(len(result.Entries))
	return result
}

// SearchByAddress searches for an entry with the exact address
func (l *Ledger) SearchByAddress(address []byte) *LedgerEntry {
	if len(address) != ADDR_LEN {
		return nil
	}

	// If address sorted, we can do a binary search
	if l.IsAddressSorted {
		idx := sort.Search(len(l.Entries), func(i int) bool {
			return bytes.Compare(l.Entries[i].Address[:], address) >= 0
		})

		if idx < len(l.Entries) && bytes.Equal(l.Entries[idx].Address[:], address) {
			return &l.Entries[idx]
		}
	} else {
		// Linear search if not sorted
		for i := range l.Entries {
			if bytes.Equal(l.Entries[i].Address[:], address) {
				return &l.Entries[i]
			}
		}
	}

	return nil
}

// LoadLedgerFromFile loads a ledger from a file
func LoadLedgerFromFile(filepath string) (*Ledger, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open ledger file: %w", err)
	}
	defer file.Close()

	// Create a new ledger
	ledger := &Ledger{}

	// Read the ledger header (size)
	var size uint64
	err = binary.Read(file, binary.LittleEndian, &size)
	if err != nil {
		return nil, fmt.Errorf("failed to read ledger size: %w", err)
	}
	ledger.Size = size

	// Read each ledger entry
	ledger.Entries = make([]LedgerEntry, size)
	for i := uint64(0); i < size; i++ {
		var entry LedgerEntry
		// Read address (ADDR_LEN bytes)
		if _, err := io.ReadFull(file, entry.Address[:]); err != nil {
			return nil, fmt.Errorf("failed to read address at entry %d: %w", i, err)
		}
		// Read balance (8 bytes)
		err = binary.Read(file, binary.LittleEndian, &entry.Balance)
		if err != nil {
			return nil, fmt.Errorf("failed to read balance at entry %d: %w", i, err)
		}
		ledger.Entries[i] = entry
	}

	// Ledger is not sorted initially
	ledger.IsBalanceSorted = false
	ledger.IsAddressSorted = false

	return ledger, nil
}

// SaveToFile saves the ledger to a file
func (l *Ledger) SaveToFile(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create ledger file: %w", err)
	}
	defer file.Close()

	// Write the ledger size
	err = binary.Write(file, binary.LittleEndian, l.Size)
	if err != nil {
		return fmt.Errorf("failed to write ledger size: %w", err)
	}

	// Write each ledger entry
	for _, entry := range l.Entries {
		// Write address (ADDR_LEN bytes)
		if _, err := file.Write(entry.Address[:]); err != nil {
			return fmt.Errorf("failed to write address: %w", err)
		}
		// Write balance (8 bytes)
		err = binary.Write(file, binary.LittleEndian, entry.Balance)
		if err != nil {
			return fmt.Errorf("failed to write balance: %w", err)
		}
	}

	return nil
}

// GetPartition returns a subset of the ledger from startIndex to endIndex (inclusive)
func (l *Ledger) GetPartition(startIndex, endIndex uint64) (*Ledger, error) {
	// Validate indices
	if startIndex > endIndex || endIndex >= l.Size {
		return nil, fmt.Errorf("invalid partition range: start=%d, end=%d, size=%d",
			startIndex, endIndex, l.Size)
	}

	// Create a new ledger with the subset of entries
	partition := &Ledger{
		Size:            endIndex - startIndex + 1,
		IsBalanceSorted: l.IsBalanceSorted,
		IsAddressSorted: l.IsAddressSorted,
	}

	// Copy the specified entries
	partition.Entries = make([]LedgerEntry, partition.Size)
	copy(partition.Entries, l.Entries[startIndex:endIndex+1])

	return partition, nil
}

type LedgerEntry struct {
	Address [ADDR_LEN]byte
	Balance uint64
}

func (le *LedgerEntry) GetAddress() WotsAddress {
	var wots WotsAddress = WotsAddressFromBytes(le.Address[:])
	return wots
}

func (le *LedgerEntry) GetBalance() uint64 {
	return le.Balance
}

func (le *LedgerEntry) SetAddress(address []byte) {
	copy(le.Address[:], address)
}

func (le *LedgerEntry) SetBalance(balance uint64) {
	le.Balance = balance
}

// ToWotsAddress converts the LedgerEntry to a WotsAddress
func (le *LedgerEntry) ToWotsAddress() WotsAddress {
	var wots WotsAddress
	copy(wots.Address[:], le.Address[:])
	wots.Amount = le.Balance
	return wots
}

// NewLedgerEntryFromWotsAddress creates a new LedgerEntry from a WotsAddress
func NewLedgerEntryFromWotsAddress(wots WotsAddress) LedgerEntry {
	var entry LedgerEntry
	copy(entry.Address[:], wots.Address[:])
	entry.Balance = wots.Amount
	return entry
}

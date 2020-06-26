package rawdb

import (
	"context"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
)

type ExternalFreezerRemoteAPI interface {
	HasAncient(ctx context.Context, kind string, number uint64) (bool, error)
	Ancient(ctx context.Context, kind string, number uint64) (string, error)
	Ancients(ctx context.Context) (uint64, error)
	AncientSize(ctx context.Context, kind string) (uint64, error)

	AppendAncient(ctx context.Context, number uint64, hash, header, body, receipt, td string)
	TruncateAncients(ctx context.Context, n uint64) error
	Sync(ctx context.Context) error
	repair() error
}

// FreezerRemoteAPI exposes a JSONRPC related API
type FreezerRemoteAPI struct {
	freezer *freezerRemote
}

// NewFreezerRemoteAPI exposes an endpoint to create a remote service
func NewFreezerRemoteAPI(freezerStr string, namespace string) (*FreezerRemoteAPI, error) {
	log.Info("constructing new freezer")
	f, err := newFreezerRemote(freezerStr, namespace, "")
	if err != nil {
		return nil, err
	}

	freezerAPI := FreezerRemoteAPI{
		freezer: f,
	}
	return &freezerAPI, nil
}

func (freezerRemoteAPI *FreezerRemoteAPI) pingVersion() string {
	return "version 1"
}

// Close terminates the chain freezer, unmapping all the data files.
func (freezerRemoteAPI *FreezerRemoteAPI) Close() error {
	return freezerRemoteAPI.freezer.Close()
}

// HasAncient returns an indicator whether the specified ancient data exists
// in the freezer.
func (freezerRemoteAPI *FreezerRemoteAPI) HasAncient(kind string, number uint64) (bool, error) {
	return freezerRemoteAPI.freezer.HasAncient(kind, number)
}

// Ancient retrieves an ancient binary blob from the append-only immutable files.
func (freezerRemoteAPI *FreezerRemoteAPI) Ancient(kind string, number uint64) (string, error) {
	ancient, err := freezerRemoteAPI.freezer.Ancient(kind, number)
	if err != nil {
		return "0x", err
	}
	return hexutil.Encode(ancient), err
}

// Ancients returns the length of the frozen items.
func (freezerRemoteAPI *FreezerRemoteAPI) Ancients() (uint64, error) {
	numAncients, err := freezerRemoteAPI.freezer.Ancients()
	if err != nil {
		return 0, err
	}
	return numAncients, err
}

// AncientSize returns the ancient size of the specified category.
func (freezerRemoteAPI *FreezerRemoteAPI) AncientSize(kind string) (uint64, error) {
	size, err := freezerRemoteAPI.freezer.AncientSize(kind)
	if err != nil {
		return 0, err
	}
	return size, err
}

// AppendAncient injects all binary blobs belong to block at the end of the
// append-only immutable table files.
//
// Notably, this function is lock free but kind of thread-safe. All out-of-order
// injection will be rejected. But if two injections with same number happen at
// the same time, we can get into the trouble.
//
// Note that the frozen marker is updated outside of the service calls.
func (freezerRemoteAPI *FreezerRemoteAPI) AppendAncient(number uint64, hash, header, body, receipts, td string) (err error) {
	var bHash, bHeader, bBody, bReceipts, bTd []byte
	bHash, err = hexutil.Decode(hash)
	if err != nil {
		return err
	}
	bHeader, err = hexutil.Decode(header)
	if err != nil {
		return err
	}
	bBody, err = hexutil.Decode(body)
	if err != nil {
		return err
	}
	bReceipts, err = hexutil.Decode(receipts)
	if err != nil {
		return err
	}
	bTd, err = hexutil.Decode(td)
	return freezerRemoteAPI.freezer.AppendAncient(number, bHash, bHeader, bBody, bReceipts, bTd)
}

// Truncate discards any recent data above the provided threshold number.
func (freezerRemoteAPI *FreezerRemoteAPI) TruncateAncients(items uint64) error {
	return freezerRemoteAPI.freezer.TruncateAncients(items)
}

// sync flushes all data tables to disk.
func (freezerRemoteAPI *FreezerRemoteAPI) Sync() error {
	return freezerRemoteAPI.freezer.Sync()
}

// repair truncates all data tables to the same length.
func (freezerRemoteAPI *FreezerRemoteAPI) repair() error {
	/*min := uint64(math.MaxUint64)
	for _, table := range f.tables {
		items := atomic.LoadUint64(&table.items)
		if min > items {
			min = items
		}
	}
	for _, table := range f.tables {
		if err := table.truncate(min); err != nil {
			return err
		}
	}
	atomic.StoreUint64(&f.frozen, min)
	*/
	return nil
}

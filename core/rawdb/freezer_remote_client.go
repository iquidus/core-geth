package rawdb

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

// FreezerRemoteClient is an RPC client implementing the interface of ethdb.AncientStore.
// Rather than implementing storage business logic, the struct's methods delegate
// the business logic to an external server that is responsible for managing the actual
// ancient store logic.
type FreezerRemoteClient struct {
	client *rpc.Client
	status string
	quit   chan struct{}
}

// newFreezerRemoteClient constructs a rpc client to connect to a remote freezer
func newFreezerRemoteClient(endpoint string, ipc bool) (*FreezerRemoteClient, error) {
	client, err := rpc.Dial(endpoint)
	if err != nil {
		return nil, err
	}

	extfreezer := &FreezerRemoteClient{
		client: client,
	}

	// Check if reachable
	version, err := extfreezer.pingVersion()
	if err != nil {
		return nil, err
	}
	extfreezer.status = fmt.Sprintf("ok [version=%v]", version)
	return extfreezer, nil
}

func (api *FreezerRemoteClient) pingVersion() (string, error) {

	return "version 1", nil
}

// Close terminates the chain freezer, unmapping all the data files.
func (api *FreezerRemoteClient) Close() error {
	return api.client.Call(nil, "freezer_close")
}

// HasAncient returns an indicator whether the specified ancient data exists
// in the freezer.
func (api *FreezerRemoteClient) HasAncient(kind string, number uint64) (bool, error) {
	var res bool
	err := api.client.Call(&res, "freezer_hasAncient", kind, number)
	return res, err
}

// Ancient retrieves an ancient binary blob from the append-only immutable files.
func (api *FreezerRemoteClient) Ancient(kind string, number uint64) ([]byte, error) {
	var res string
	if err := api.client.Call(&res, "freezer_ancient", kind, number); err != nil {
		return nil, err
	}
	return hexutil.Decode(res)
}

// Ancients returns the length of the frozen items.
func (api *FreezerRemoteClient) Ancients() (uint64, error) {
	var res uint64
	err := api.client.Call(&res, "freezer_ancients")
	return res, err
}

// AncientSize returns the ancient size of the specified category.
func (api *FreezerRemoteClient) AncientSize(kind string) (uint64, error) {
	var res uint64
	err := api.client.Call(&res, "freezer_ancientSize", kind)
	return res, err
}

// AppendAncient injects all binary blobs belong to block at the end of the
// append-only immutable table files.
//
// Notably, this function is lock free but kind of thread-safe. All out-of-order
// injection will be rejected. But if two injections with same number happen at
// the same time, we can get into the trouble.
//
// Note that the frozen marker is updated outside of the service calls.
func (api *FreezerRemoteClient) AppendAncient(number uint64, hash, header, body, receipts, td []byte) (err error) {
	hexHash := hexutil.Encode(hash)
	hexHeader := hexutil.Encode(header)
	hexBody := hexutil.Encode(body)
	hexReceipts := hexutil.Encode(receipts)
	hexTd := hexutil.Encode(td)
	err = api.client.Call(nil, "freezer_appendAncient", number, hexHash, hexHeader, hexBody, hexReceipts, hexTd)
	return
}

// TruncateAncients discards any recent data above the provided threshold number.
func (api *FreezerRemoteClient) TruncateAncients(items uint64) error {
	return api.client.Call(nil, "freezer_truncateAncients", items)
}

// Sync flushes all data tables to disk.
func (api *FreezerRemoteClient) Sync() error {
	return api.client.Call(nil, "freezer_sync")
}

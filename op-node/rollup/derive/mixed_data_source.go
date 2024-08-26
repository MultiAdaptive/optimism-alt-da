package derive

import (
	"context"
	"errors"
	"fmt"
	plasma "github.com/ethereum-optimism/optimism/op-plasma"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"io"
)

type MixedDataSource struct {
	data         []blobOrCalldata
	ref          eth.L1BlockRef
	batcherAddr  common.Address
	dsCfg        DataSourceConfig
	fetcher      L1TransactionFetcher
	blobsFetcher L1BlobsFetcher
	log          log.Logger

	l1            L1Fetcher
	plasmaFetcher PlasmaInputFetcher
}

func NewMixedDataSource(ctx context.Context, log log.Logger, dsCfg DataSourceConfig, l1 L1Fetcher, blobsFetcher L1BlobsFetcher, plasmaFetcher PlasmaInputFetcher, ref eth.L1BlockRef, batcherAddr common.Address) DataIter {
	return &MixedDataSource{
		ref:           ref,
		batcherAddr:   batcherAddr,
		dsCfg:         dsCfg,
		blobsFetcher:  blobsFetcher,
		log:           log.New("origin", ref),
		l1:            l1,
		plasmaFetcher: plasmaFetcher,
	}
}

func (mds *MixedDataSource) Next(ctx context.Context) (eth.Data, error) {
	if mds.data == nil {
		var err error
		if mds.data, err = mds.open(ctx); err != nil {

			return nil, err
		}
	}

	if len(mds.data) == 0 {
		return nil, io.EOF
	}

	next := mds.data[0]
	mds.data = mds.data[1:]

	if next.blob != nil {
		data, err := next.blob.ToData()
		if err != nil {
			mds.log.Error("ignoring blob due to parse failure", "err", err)
			return mds.Next(ctx)
		}
		return data, nil
	}

	if next.calldata == nil {
		return nil, io.EOF
	}
	calldata := *next.calldata
	if len(calldata) == 0 {
		return nil, NotEnoughData
	}
	// If the tx data type is not plasma, we forward it downstream to let the next
	// steps validate and potentially parse it as L1 DA inputs.
	if calldata[0] != plasma.TxDataVersion1 {
		return calldata, nil
	}

	// validate batcher inbox data is a commitment.
	// strip the transaction data version byte from the data before decoding.
	comm, err := plasma.DecodeCommitmentData(calldata[1:])
	if err != nil {
		mds.log.Warn("invalid commitment", "commitment", calldata, "err", err)
		return nil, NotEnoughData
	}

	return mds.handleCommitment(ctx, comm)
}
func (mds *MixedDataSource) open(ctx context.Context) ([]blobOrCalldata, error) {
	_, txs, err := mds.l1.InfoAndTxsByHash(ctx, mds.ref.Hash)
	if err != nil {
		if errors.Is(err, ethereum.NotFound) {
			return nil, NewResetError(fmt.Errorf("failed to open mixed data ource source: %w", err))
		}
		return nil, NewTemporaryError(fmt.Errorf("failed to open mixed data source: %w", err))
	}

	blobData, callData, hashes := extractDataAndHashes(txs, &mds.dsCfg, mds.batcherAddr)

	if len(hashes) == 0 {
		// there are no blobs to fetch so we can return immediately
		return blobData, nil
	}

	// download the actual blob bodies corresponding to the indexed blob hashes
	blobs, err := mds.blobsFetcher.GetBlobs(ctx, mds.ref, hashes)
	if err != nil {
		return callData, nil
	}

	// go back over the data array and populate the blob pointers
	if err = fillBlobPointers(blobData, blobs); err != nil {
		// this shouldn't happen unless there is a bug in the blobs fetcher
		return nil, NewResetError(fmt.Errorf("mixed failed to fill blob pointers: %w", err))
	}
	return blobData, nil
}

func extractDataAndHashes(txs types.Transactions, config *DataSourceConfig, batcherAddr common.Address) ([]blobOrCalldata, []blobOrCalldata, []eth.IndexedBlobHash) {
	var blob []blobOrCalldata
	var data []blobOrCalldata
	var hashes []eth.IndexedBlobHash
	blobIndex := 0 // index of each blob in the block's blob sidecar
	for _, tx := range txs {
		// skip any non-batcher transactions
		if !isValidBatchTx(tx, config.l1Signer, config.batchInboxAddress, batcherAddr) {
			blobIndex += len(tx.BlobHashes())
			continue
		}

		// handle blob batcher transactions by extracting their blob hashes, ignoring any calldata.
		if len(tx.Data()) > 0 {
			calldata := eth.Data(tx.Data())
			data = append(data, blobOrCalldata{nil, &calldata}) // will fill in blob pointers after we download them below
		}
		for _, h := range tx.BlobHashes() {
			idh := eth.IndexedBlobHash{
				Index: uint64(blobIndex),
				Hash:  h,
			}
			hashes = append(hashes, idh)
			blob = append(blob, blobOrCalldata{nil, nil}) // will fill in blob pointers after we download them below
			blobIndex += 1
		}
	}
	return blob, data, hashes
}

func fillDataPointers(data []eth.Data, blobs []*eth.Blob) error {
	blobIndex := 0
	for i := range data {
		if blobIndex >= len(blobs) {
			return fmt.Errorf("didn't get enough blobs")
		}
		if blobs[blobIndex] == nil {
			return fmt.Errorf("found a nil blob")
		}
		blobData, err := blobs[blobIndex].ToData()
		if err != nil {
			return fmt.Errorf("blobData Error")
		}
		data[i] = blobData
		blobIndex++
	}
	if blobIndex != len(blobs) {
		return fmt.Errorf("got too many blobs")
	}
	return nil
}

func (mds *MixedDataSource) handleCommitment(ctx context.Context, comm plasma.CommitmentData) (eth.Data, error) {
	data, err := mds.plasmaFetcher.GetInput(ctx, mds.l1, comm, mds.ref.ID())
	if errors.Is(err, plasma.ErrReorgRequired) {
		return nil, NewResetError(err)
	} else if errors.Is(err, plasma.ErrExpiredChallenge) {
		mds.log.Warn("Challenge expired, skipping batch", "comm", comm)
		return mds.Next(ctx)
	} else if errors.Is(err, plasma.ErrMissingPastWindow) {
		return nil, NewCriticalError(fmt.Errorf("data for comm %x not available: %w", comm, err))
	} else if errors.Is(err, plasma.ErrPendingChallenge) {
		return nil, NotEnoughData
	} else if err != nil {
		return nil, NewTemporaryError(fmt.Errorf("failed to fetch input data with comm %x: %w", comm, err))
	}
	if comm.CommitmentType() == plasma.Keccak256CommitmentType && len(data) > plasma.MaxInputSize {
		mds.log.Warn("Input data exceeds max size", "size", len(data), "max", plasma.MaxInputSize)
		return mds.Next(ctx)
	}
	return data, nil
}

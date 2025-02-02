package mon

import (
	"context"
	"errors"
	"testing"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

var (
	mockRootClaim = common.HexToHash("0x10")
)

func TestDetector_CheckRootAgreement(t *testing.T) {
	t.Parallel()

	t.Run("OutputFetchFails", func(t *testing.T) {
		validator, rollup, metrics := setupOutputValidatorTest(t)
		rollup.err = errors.New("boom")
		agree, fetched, err := validator.CheckRootAgreement(context.Background(), 0, mockRootClaim)
		require.ErrorIs(t, err, rollup.err)
		require.Equal(t, common.Hash{}, fetched)
		require.False(t, agree)
		require.Zero(t, metrics.fetchTime)
	})

	t.Run("OutputMismatch", func(t *testing.T) {
		validator, _, metrics := setupOutputValidatorTest(t)
		agree, fetched, err := validator.CheckRootAgreement(context.Background(), 0, common.Hash{})
		require.NoError(t, err)
		require.Equal(t, mockRootClaim, fetched)
		require.False(t, agree)
		require.NotZero(t, metrics.fetchTime)
	})

	t.Run("OutputMatches", func(t *testing.T) {
		validator, _, metrics := setupOutputValidatorTest(t)
		agree, fetched, err := validator.CheckRootAgreement(context.Background(), 0, mockRootClaim)
		require.NoError(t, err)
		require.Equal(t, mockRootClaim, fetched)
		require.True(t, agree)
		require.NotZero(t, metrics.fetchTime)
	})

	t.Run("OutputNotFound", func(t *testing.T) {
		validator, rollup, metrics := setupOutputValidatorTest(t)
		// This crazy error is what we actually get back from the API
		rollup.err = errors.New("failed to get L2 block ref with sync status: failed to determine L2BlockRef of height 42984924, could not get payload: not found")
		agree, fetched, err := validator.CheckRootAgreement(context.Background(), 42984924, mockRootClaim)
		require.NoError(t, err)
		require.Equal(t, common.Hash{}, fetched)
		require.False(t, agree)
		require.Zero(t, metrics.fetchTime)
	})
}

func setupOutputValidatorTest(t *testing.T) (*outputValidator, *stubRollupClient, *stubOutputMetrics) {
	client := &stubRollupClient{}
	metrics := &stubOutputMetrics{}
	validator := newOutputValidator(metrics, client)
	return validator, client, metrics
}

type stubOutputMetrics struct {
	fetchTime float64
}

func (s *stubOutputMetrics) RecordOutputFetchTime(fetchTime float64) {
	s.fetchTime = fetchTime
}

type stubRollupClient struct {
	blockNum uint64
	err      error
}

func (s *stubRollupClient) OutputAtBlock(ctx context.Context, blockNum uint64) (*eth.OutputResponse, error) {
	s.blockNum = blockNum
	return &eth.OutputResponse{OutputRoot: eth.Bytes32(mockRootClaim)}, s.err
}

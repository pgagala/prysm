package cache

import (
	"testing"

	types "github.com/prysmaticlabs/eth2-types"
	"github.com/prysmaticlabs/prysm/beacon-chain/state"
	v1 "github.com/prysmaticlabs/prysm/beacon-chain/state/v1"
	ethpb "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/shared/bytesutil"
	"github.com/prysmaticlabs/prysm/shared/params"
	"github.com/prysmaticlabs/prysm/shared/testutil/assert"
	"github.com/prysmaticlabs/prysm/shared/testutil/require"
	"google.golang.org/protobuf/proto"
)

func TestCheckpointStateCache_StateByCheckpoint(t *testing.T) {
	cache := NewCheckpointStateCache()

	cp1 := &ethpb.Checkpoint{Epoch: 1, Root: bytesutil.PadTo([]byte{'A'}, 32)}
	st, err := v1.InitializeFromProto(&ethpb.BeaconState{
		GenesisValidatorsRoot: params.BeaconConfig().ZeroHash[:],
		Slot:                  64,
	})
	require.NoError(t, err)

	s, err := cache.StateByCheckpoint(cp1)
	require.NoError(t, err)
	assert.Equal(t, state.BeaconState(nil), s, "Expected state not to exist in empty cache")

	require.NoError(t, cache.AddCheckpointState(cp1, st))

	s, err = cache.StateByCheckpoint(cp1)
	require.NoError(t, err)

	pbState1, err := v1.ProtobufBeaconState(s.InnerStateUnsafe())
	require.NoError(t, err)
	pbstate, err := v1.ProtobufBeaconState(st.InnerStateUnsafe())
	require.NoError(t, err)
	if !proto.Equal(pbState1, pbstate) {
		t.Error("incorrectly cached state")
	}

	cp2 := &ethpb.Checkpoint{Epoch: 2, Root: bytesutil.PadTo([]byte{'B'}, 32)}
	st2, err := v1.InitializeFromProto(&ethpb.BeaconState{
		Slot: 128,
	})
	require.NoError(t, err)
	require.NoError(t, cache.AddCheckpointState(cp2, st2))

	s, err = cache.StateByCheckpoint(cp2)
	require.NoError(t, err)
	assert.DeepEqual(t, st2.CloneInnerState(), s.CloneInnerState(), "incorrectly cached state")

	s, err = cache.StateByCheckpoint(cp1)
	require.NoError(t, err)
	assert.DeepEqual(t, st.CloneInnerState(), s.CloneInnerState(), "incorrectly cached state")
}

func TestCheckpointStateCache_MaxSize(t *testing.T) {
	c := NewCheckpointStateCache()
	st, err := v1.InitializeFromProto(&ethpb.BeaconState{
		Slot: 0,
	})
	require.NoError(t, err)

	for i := uint64(0); i < uint64(maxCheckpointStateSize+100); i++ {
		require.NoError(t, st.SetSlot(types.Slot(i)))
		require.NoError(t, c.AddCheckpointState(&ethpb.Checkpoint{Epoch: types.Epoch(i), Root: make([]byte, 32)}, st))
	}

	assert.Equal(t, maxCheckpointStateSize, len(c.cache.Keys()))
}

package set

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	s := New("a", "b", "c")
	assert.Len(t, s, 3, "expected set length 3")
	assert.Contains(t, s, "a")
	assert.Contains(t, s, "b")
	assert.Contains(t, s, "c")
}

func TestAdd(t *testing.T) {
	s := New[string]()
	s.Add("a")
	assert.Contains(t, s, "a", "Add failed")
	assert.Len(t, s, 1, "expected set length 1")
	// Adding duplicate should be no-op
	s.Add("a")
	assert.Len(t, s, 1, "expected set length 1 after duplicate add")
}

func TestRemove(t *testing.T) {
	s := New("a", "b", "c")
	s.Remove("b")
	assert.NotContains(t, s, "b", "Remove failed")
	assert.Len(t, s, 2, "expected set length 2")
	// Removing non-existent should be no-op
	s.Remove("x")
	assert.Len(t, s, 2, "expected set length 2 after removing non-existent")
}

func TestHas(t *testing.T) {
	s := New("a", "b")
	assert.Contains(t, s, "a", "Has returned false for existing member")
	assert.NotContains(t, s, "c", "Has returned true for non-existent member")
}

func TestMembers(t *testing.T) {
	s := New("a", "b", "c")
	members := s.Members()
	assert.Len(t, members, 3, "expected 3 members")
	// Check all expected members are present
	assert.ElementsMatch(
		t,
		members,
		[]string{"a", "b", "c"},
		"Members missing expected values",
	)
}

func TestMarshalJSON(t *testing.T) {
	s := New("a", "b", "c")
	_, err := json.Marshal(s)
	require.NoError(t, err, "MarshalJSON failed")
}

func TestUnmarshalJSON(t *testing.T) {
	data := []byte(`["a","b","c"]`)
	var s Set[string]
	err := json.Unmarshal(data, &s)
	require.NoError(t, err, "UnmarshalJSON failed")
	assert.Len(t, s, 3, "expected set length 3")
	assert.Contains(t, s, "a")
	assert.Contains(t, s, "b")
	assert.Contains(t, s, "c")
}

func TestRoundTrip(t *testing.T) {
	original := New("x", "y", "z")
	data, err := json.Marshal(original)
	require.NoError(t, err, "Marshal failed")

	var roundTrip Set[string]
	err = json.Unmarshal(data, &roundTrip)
	require.NoError(t, err, "Unmarshal failed")
	assert.Len(t, roundTrip, len(original), "round trip length mismatch")
	for k := range original {
		assert.Contains(t, roundTrip, k, "round trip missing member %q", k)
	}
}

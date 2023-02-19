package cpngo_test

import (
	"testing"

	"github.com/adamlouis/cpngo/cpngo"
	"github.com/stretchr/testify/require"
)

func TestCPNEmpty(t *testing.T) {
	rnr, err := cpngo.NewRunner(&cpngo.Net{
		[]cpngo.Place{},
		[]cpngo.Transition{},
		[]cpngo.InputArc{},
		[]cpngo.OutputArc{},
		[]cpngo.Token{},
	})

	require.NoError(t, err)
	ts := rnr.Enabled()
	require.Len(t, ts, 0)
	require.Error(t, rnr.FireAny())
}

func TestExampleCPN(t *testing.T) {
	rnr, err := cpngo.NewRunner(&cpngo.Net{
		[]cpngo.Place{
			{ID: "p1"},
			{ID: "p2"},
			{ID: "p3"},
			{ID: "p4"},
			{ID: "p5"},
		},
		[]cpngo.Transition{
			{ID: "t1"},
			{ID: "t2"},
			{ID: "t3"},
			{ID: "t4"},
		},
		[]cpngo.InputArc{
			{ID: "p1t1", FromID: "p1", ToID: "t1"},
			{ID: "p2t2", FromID: "p2", ToID: "t2"},
			{ID: "p3t3", FromID: "p3", ToID: "t3"},
			{ID: "p4t4", FromID: "p4", ToID: "t4"},
		},
		[]cpngo.OutputArc{
			{ID: "t1p2", FromID: "t1", ToID: "p2"},
			{ID: "t1p3", FromID: "t1", ToID: "p3"},
			{ID: "t2p4", FromID: "t2", ToID: "p4"},
			{ID: "t3p4", FromID: "t3", ToID: "p4"},
			{ID: "t4p5", FromID: "t4", ToID: "p5"},
		},
		[]cpngo.Token{
			{ID: "t1", PlaceID: "p1", Color: "foobar"},
		},
	})

	require.NoError(t, err)
	require.Len(t, rnr.Enabled(), 1, "expected 1 enabled transition")
	tks := rnr.Tokens()
	require.Len(t, tks, 1, "expected 1 tokens")
	require.Equal(t, tks[0].PlaceID, "p1")

	require.NoError(t, rnr.FireAny())
	require.Len(t, rnr.Enabled(), 2, "expected 2 enabled transitions")
	tks = rnr.Tokens()
	require.Len(t, tks, 2, "expected 2 tokens")

	require.NoError(t, rnr.FireAny())
	require.Len(t, rnr.Enabled(), 2, "expected 2 enabled transitions")
	tks = rnr.Tokens()
	require.Len(t, tks, 2, "expected 2 tokens")

	require.NoError(t, rnr.FireAny())
	require.Len(t, rnr.Enabled(), 1, "expected 1 enabled transition")
	tks = rnr.Tokens()
	require.Len(t, tks, 2, "expected 2 tokens")

	require.NoError(t, rnr.FireAny())
	require.Len(t, rnr.Enabled(), 1, "expected 1 enabled transition")
	tks = rnr.Tokens()
	require.Len(t, tks, 2, "expected 2 token")

	require.NoError(t, rnr.FireAny())
	require.Len(t, rnr.Enabled(), 0, "expected 0 enabled transitions")
	tokens := rnr.Tokens()
	require.Len(t, tokens, 2, "expected 2 tokens")
	require.Equal(t, "p5", tokens[0].PlaceID, "expected token in p5")
	require.Equal(t, "p5", tokens[1].PlaceID, "expected token in p5")
}

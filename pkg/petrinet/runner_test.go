package petrinet_test

import (
	"testing"

	"github.com/adamlouis/cpngo/pkg/petrinet"
	"github.com/stretchr/testify/require"
)

func TestCPNEmpty(t *testing.T) {
	rnr, err := petrinet.NewRunner(&petrinet.Net{
		[]petrinet.Place{},
		[]petrinet.Transition{},
		[]petrinet.InputArc{},
		[]petrinet.OutputArc{},
		[]petrinet.Token{},
	})

	require.NoError(t, err)
	ts := rnr.Enabled()
	require.Len(t, ts, 0)
	_, err = rnr.FireAny()
	require.Error(t, err)
}

func TestExampleCPN(t *testing.T) {
	rnr, err := petrinet.NewRunner(&petrinet.Net{
		[]petrinet.Place{
			{ID: "p1"},
			{ID: "p2"},
			{ID: "p3"},
			{ID: "p4"},
			{ID: "p5"},
		},
		[]petrinet.Transition{
			{ID: "t1"},
			{ID: "t2"},
			{ID: "t3"},
			{ID: "t4"},
		},
		[]petrinet.InputArc{
			{FromPlaceID: "p1", ToTransitionID: "t1"},
			{FromPlaceID: "p2", ToTransitionID: "t2"},
			{FromPlaceID: "p3", ToTransitionID: "t3"},
			{FromPlaceID: "p4", ToTransitionID: "t4"},
		},
		[]petrinet.OutputArc{
			{FromTransitionID: "t1", ToPlaceID: "p2"},
			{FromTransitionID: "t1", ToPlaceID: "p3"},
			{FromTransitionID: "t2", ToPlaceID: "p4"},
			{FromTransitionID: "t3", ToPlaceID: "p4"},
			{FromTransitionID: "t4", ToPlaceID: "p5"},
		},
		[]petrinet.Token{
			{ID: "t1", OnPlaceID: "p1", Color: "foobar"},
		},
	})

	require.NoError(t, err)
	require.Len(t, rnr.Enabled(), 1, "expected 1 enabled transition")
	tks := rnr.Tokens()
	require.Len(t, tks, 1, "expected 1 tokens")
	require.Equal(t, tks[0].OnPlaceID, petrinet.PlaceID("p1"))

	_, err = rnr.FireAny()
	require.NoError(t, err)
	require.Len(t, rnr.Enabled(), 2, "expected 2 enabled transitions")
	tks = rnr.Tokens()
	require.Len(t, tks, 2, "expected 2 tokens")

	_, err = rnr.FireAny()
	require.NoError(t, err)
	require.Len(t, rnr.Enabled(), 2, "expected 2 enabled transitions")
	tks = rnr.Tokens()
	require.Len(t, tks, 2, "expected 2 tokens")

	_, err = rnr.FireAny()
	require.NoError(t, err)
	require.Len(t, rnr.Enabled(), 1, "expected 1 enabled transition")
	tks = rnr.Tokens()
	require.Len(t, tks, 2, "expected 2 tokens")

	_, err = rnr.FireAny()
	require.NoError(t, err)
	require.Len(t, rnr.Enabled(), 1, "expected 1 enabled transition")
	tks = rnr.Tokens()
	require.Len(t, tks, 2, "expected 2 token")

	_, err = rnr.FireAny()
	require.NoError(t, err)
	require.Len(t, rnr.Enabled(), 0, "expected 0 enabled transitions")
	tokens := rnr.Tokens()

	require.Len(t, tokens, 2, "expected 2 tokens")
	require.Equal(t, petrinet.PlaceID("p5"), tokens[0].OnPlaceID, "expected token in p5")
	require.Equal(t, petrinet.PlaceID("p5"), tokens[1].OnPlaceID, "expected token in p5")
}

func TestExpr(t *testing.T) {
	rnr, err := petrinet.NewRunner(&petrinet.Net{
		[]petrinet.Place{
			{ID: "p1"},
			{ID: "p2"},
			{ID: "p3"},
		},
		[]petrinet.Transition{
			{ID: "t1"},
			{ID: "t2"},
		},
		[]petrinet.InputArc{
			{FromPlaceID: "p1", ToTransitionID: "t1", Expr: "color == \"foobar\""},
			{FromPlaceID: "p2", ToTransitionID: "t2"},
		},
		[]petrinet.OutputArc{
			{FromTransitionID: "t1", ToPlaceID: "p2", Expr: "42"},
			{FromTransitionID: "t2", ToPlaceID: "p3", Expr: "colors[0] + 42"},
		},
		[]petrinet.Token{
			{ID: "tk1", OnPlaceID: "p1", Color: "foobar"},
			{ID: "tk2", OnPlaceID: "p1", Color: "buzz"},
		},
	})
	require.NoError(t, err)

	require.NoError(t, err)
	require.Len(t, rnr.Enabled(), 1, "expected 1 enabled transition")
	tks := rnr.Tokens()
	require.Len(t, tks, 2, "expected 2 tokens")

	_, err = rnr.FireAny()
	require.NoError(t, err)
	require.Len(t, rnr.Enabled(), 1, "expected 1 enabled transitions")
	tks = rnr.Tokens()
	require.Len(t, tks, 2, "expected 2 tokens")

	_, err = rnr.FireAny()
	require.NoError(t, err)
	require.Len(t, rnr.Enabled(), 0, "expected 0 enabled transitions")
	tks = rnr.Tokens()
	require.Len(t, tks, 2, "expected 2 tokens")

	for _, tk := range tks {
		if tk.OnPlaceID == "p1" {
			require.Equal(t, "buzz", tk.Color)
		}
		if tk.OnPlaceID == "p3" {
			require.Equal(t, 84, tk.Color)
		}
	}
}

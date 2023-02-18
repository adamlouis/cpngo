package cpngo

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
)

type Net struct {
	placesByID      map[string]*place
	transitionsByID map[string]*transition
	inputArcsByID   map[string]*inputArc
	outputArcsByID  map[string]*outputArc
	tokensByID      map[string]*token
}

type Summary struct {
	Places      []Place      `json:"places"`
	Transitions []Transition `json:"transitions"`
	InputArcs   []InputArc   `json:"input_arcs"`
	OutputArcs  []OutputArc  `json:"output_arcs"`
	Tokens      []Token      `json:"tokens"`
}

type Place struct {
	ID string `json:"id"`
}
type Transition struct {
	ID string `json:"id"`
}
type InputArc struct {
	ID     string `json:"id"`
	FromID string `json:"from_id"`
	ToID   string `json:"to_id"`
}
type OutputArc struct {
	ID     string `json:"id"`
	FromID string `json:"from_id"`
	ToID   string `json:"to_id"`
}
type Token struct {
	ID      string `json:"id"`
	PlaceID string `json:"place_id"`
	Color   any    `json:"color"`
}

type place struct {
	Place
	tokensByID map[string]*token
}
type transition struct {
	Transition
	inputPlaces  []*place
	outputPlaces []*place
}
type inputArc struct {
	InputArc
	from *place
	to   *transition
}
type outputArc struct {
	OutputArc
	from *transition
	to   *place
}
type token struct {
	Token
	place *place
}

func NewNet(
	Places []Place,
	Transitions []Transition,
	InputArcs []InputArc,
	OutputArcs []OutputArc,
	Tokens []Token,
) (*Net, error) {
	ret := &Net{
		placesByID:      map[string]*place{},
		transitionsByID: map[string]*transition{},
		inputArcsByID:   map[string]*inputArc{},
		outputArcsByID:  map[string]*outputArc{},
		tokensByID:      map[string]*token{},
	}

	for _, p := range Places {
		ret.placesByID[p.ID] = &place{
			Place:      p,
			tokensByID: map[string]*token{},
		}
	}
	for _, t := range Transitions {
		ret.transitionsByID[t.ID] = &transition{Transition: t}
	}
	for _, a := range InputArcs {
		ret.inputArcsByID[a.ID] = &inputArc{InputArc: a}
	}
	for _, a := range OutputArcs {
		ret.outputArcsByID[a.ID] = &outputArc{OutputArc: a}
	}
	for _, t := range Tokens {
		ret.tokensByID[t.ID] = &token{Token: t}
	}

	if err := ret.connectPointers(); err != nil {
		return nil, err
	}

	return ret, nil
}

func (n *Net) Places() []Place {
	ret := []Place{}
	for _, p := range n.placesByID {
		ret = append(ret, p.Place)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].ID < ret[j].ID })
	return ret
}
func (n *Net) Transitions() []Transition {
	ret := []Transition{}
	for _, t := range n.transitionsByID {
		ret = append(ret, t.Transition)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].ID < ret[j].ID })
	return ret
}
func (n *Net) InputArcs() []InputArc {
	ret := []InputArc{}
	for _, a := range n.inputArcsByID {
		ret = append(ret, a.InputArc)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].ID < ret[j].ID })
	return ret
}
func (n *Net) OutputArcs() []OutputArc {
	ret := []OutputArc{}
	for _, a := range n.outputArcsByID {
		ret = append(ret, a.OutputArc)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].ID < ret[j].ID })
	return ret
}
func (n *Net) Tokens() []Token {
	ret := []Token{}
	for _, t := range n.tokensByID {
		ret = append(ret, t.Token)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].ID < ret[j].ID })
	return ret
}
func (n *Net) Summary() Summary {
	return Summary{
		Places:      n.Places(),
		Transitions: n.Transitions(),
		InputArcs:   n.InputArcs(),
		OutputArcs:  n.OutputArcs(),
		Tokens:      n.Tokens(),
	}
}

func (n *Net) Enabled() []*Transition {
	ret := []*Transition{}
	for _, t := range n.transitionsByID {
		enabled, err := n.isTransitionEnabled(t.ID)
		if err != nil {
			panic("unreachable: corrupted net: " + err.Error())
		}
		if enabled {
			ret = append(ret, &t.Transition)
		}
	}
	return ret
}

func (n *Net) isTransitionEnabled(id string) (bool, error) {
	t, ok := n.transitionsByID[id]
	if !ok {
		return false, fmt.Errorf("transition %q does not exist", id)
	}

	if len(t.inputPlaces) == 0 {
		panic("unreachable: corrupted net: transition has no input places")
	}

	for _, p := range t.inputPlaces {
		if len(p.tokensByID) < 1 {
			return false, nil
		}
	}
	return true, nil
}

func (n *Net) FireAny() error {
	ts := n.Enabled()
	if len(ts) < 1 {
		return fmt.Errorf("no transitions are enabled")
	}
	sort.Slice(ts, func(i, j int) bool {
		return ts[i].ID < ts[j].ID
	})
	return n.Fire(ts[0].ID)
}

func (n *Net) Fire(id string) error {
	// TODO(adamlouis): implement execution polices
	// TODO(adamlouis): implement color
	// TODO(adamlouis): implement guards
	// TODO(adamlouis): implement arc expressions

	t, ok := n.transitionsByID[id]
	if !ok {
		return fmt.Errorf("transition %q does not exist", id)
	}

	enabled, err := n.isTransitionEnabled(id)
	if err != nil {
		return err
	}
	if !enabled {
		return fmt.Errorf("transition %q is not enabled", id)
	}

	for _, p := range t.inputPlaces {
		tkp, err := oneV(p.tokensByID)
		if err != nil {
			return fmt.Errorf("transition %q failed to consume token: %w", id, err)
		}
		if err := n.consumeToken((*tkp).ID); err != nil {
			return fmt.Errorf("transition %q failed to consume token: %w", id, err)
		}
	}

	for _, p := range t.outputPlaces {
		tk := &token{
			Token: Token{
				ID:      uuid.New().String(),
				PlaceID: p.ID,
				Color:   nil,
			},
			place: p,
		}
		p.tokensByID[tk.ID] = tk
		n.tokensByID[tk.ID] = tk
	}

	return nil
}

func (n *Net) consumeToken(id string) error {
	t, ok := n.tokensByID[id]
	if !ok {
		return fmt.Errorf("token %q does not exist", id)
	}
	if t.place == nil {
		return fmt.Errorf("token %q is not in a place", id)
	}
	delete(n.tokensByID, id)
	delete(t.place.tokensByID, id)
	t.place = nil
	t.PlaceID = ""
	return nil
}

func (n *Net) connectPointers() error {
	for _, t := range n.tokensByID {
		p, ok := n.placesByID[t.PlaceID]
		if !ok {
			return fmt.Errorf("token %q has invalid place id %q", t.ID, t.PlaceID)
		}
		t.place = p
		p.tokensByID[t.ID] = t
	}

	for _, a := range n.inputArcsByID {
		fp, ok := n.placesByID[a.FromID]
		if !ok {
			return fmt.Errorf("input arc %q has invalid from id %q", a.ID, a.FromID)
		}
		tt, ok := n.transitionsByID[a.ToID]
		if !ok {
			return fmt.Errorf("input arc %q has invalid to id %q", a.ID, a.ToID)
		}

		a.from = fp
		a.to = tt
		tt.inputPlaces = append(tt.inputPlaces, fp)
	}

	for _, a := range n.outputArcsByID {
		ft, ok := n.transitionsByID[a.FromID]
		if !ok {
			return fmt.Errorf("output arc %q has invalid from id %q", a.ID, a.FromID)
		}
		tp, ok := n.placesByID[a.ToID]
		if !ok {
			return fmt.Errorf("output arc %q has invalid to id %q", a.ID, a.ToID)
		}

		a.from = ft
		a.to = tp
		ft.outputPlaces = append(ft.outputPlaces, tp)
	}

	return nil
}

func oneV[TK comparable, TV any](m map[TK]TV) (*TV, error) {
	for _, v := range m {
		return &v, nil
	}
	return nil, fmt.Errorf("map is empty")
}

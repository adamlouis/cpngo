package cpngo

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
)

type Runner struct {
	placesByID      map[string]*place
	transitionsByID map[string]*transition
	inputArcsByID   map[string]*inputArc
	outputArcsByID  map[string]*outputArc
	tokensByID      map[string]*token
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

func NewRunner(n *Net) (*Runner, error) {
	ret := &Runner{
		placesByID:      map[string]*place{},
		transitionsByID: map[string]*transition{},
		inputArcsByID:   map[string]*inputArc{},
		outputArcsByID:  map[string]*outputArc{},
		tokensByID:      map[string]*token{},
	}

	for _, p := range n.Places {
		ret.placesByID[p.ID] = &place{
			Place:      p,
			tokensByID: map[string]*token{},
		}
	}
	for _, t := range n.Transitions {
		ret.transitionsByID[t.ID] = &transition{Transition: t}
	}
	for _, a := range n.InputArcs {
		ret.inputArcsByID[a.ID] = &inputArc{InputArc: a}
	}
	for _, a := range n.OutputArcs {
		ret.outputArcsByID[a.ID] = &outputArc{OutputArc: a}
	}
	for _, t := range n.Tokens {
		ret.tokensByID[t.ID] = &token{Token: t}
	}

	if err := ret.connectPointers(); err != nil {
		return nil, err
	}

	return ret, nil
}

func (r *Runner) Places() []Place {
	ret := []Place{}
	for _, p := range r.placesByID {
		ret = append(ret, p.Place)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].ID < ret[j].ID })
	return ret
}
func (r *Runner) Transitions() []Transition {
	ret := []Transition{}
	for _, t := range r.transitionsByID {
		ret = append(ret, t.Transition)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].ID < ret[j].ID })
	return ret
}
func (r *Runner) InputArcs() []InputArc {
	ret := []InputArc{}
	for _, a := range r.inputArcsByID {
		ret = append(ret, a.InputArc)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].ID < ret[j].ID })
	return ret
}
func (r *Runner) OutputArcs() []OutputArc {
	ret := []OutputArc{}
	for _, a := range r.outputArcsByID {
		ret = append(ret, a.OutputArc)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].ID < ret[j].ID })
	return ret
}
func (r *Runner) Tokens() []Token {
	ret := []Token{}
	for _, t := range r.tokensByID {
		ret = append(ret, t.Token)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].ID < ret[j].ID })
	return ret
}
func (r *Runner) Net() Net {
	return Net{
		Places:      r.Places(),
		Transitions: r.Transitions(),
		InputArcs:   r.InputArcs(),
		OutputArcs:  r.OutputArcs(),
		Tokens:      r.Tokens(),
	}
}

func (r *Runner) Enabled() []*Transition {
	ret := []*Transition{}
	for _, t := range r.transitionsByID {
		enabled, err := r.isTransitionEnabled(t.ID)
		if err != nil {
			panic("unreachable: corrupted net: " + err.Error())
		}
		if enabled {
			ret = append(ret, &t.Transition)
		}
	}
	return ret
}

func (r *Runner) isTransitionEnabled(id string) (bool, error) {
	t, ok := r.transitionsByID[id]
	if !ok {
		return false, fmt.Errorf("transition %q does not exist", id)
	}

	if len(t.inputPlaces) == 0 {
		panic("unreachable: corrupted net: transition has no input places")
	}

	for _, p := range t.inputPlaces {
		// TODO(adam): check arc expression here
		if len(p.tokensByID) < 1 {
			return false, nil
		}
	}
	return true, nil
}

func (r *Runner) FireAny() error {
	ts := r.Enabled()
	if len(ts) < 1 {
		return fmt.Errorf("no transitions are enabled")
	}
	sort.Slice(ts, func(i, j int) bool {
		return ts[i].ID < ts[j].ID
	})
	return r.Fire(ts[0].ID)
}

func (r *Runner) Fire(id string) error {
	// TODO(adamlouis): implement execution polices
	// TODO(adamlouis): implement color
	// TODO(adamlouis): implement guards
	// TODO(adamlouis): implement arc expressions

	t, ok := r.transitionsByID[id]
	if !ok {
		return fmt.Errorf("transition %q does not exist", id)
	}

	enabled, err := r.isTransitionEnabled(id)
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
		if err := r.consumeToken((*tkp).ID); err != nil {
			return fmt.Errorf("transition %q failed to consume token: %w", id, err)
		}
	}

	// TODO(adam): use arc expression here
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
		r.tokensByID[tk.ID] = tk
	}

	return nil
}

func (r *Runner) consumeToken(id string) error {
	t, ok := r.tokensByID[id]
	if !ok {
		return fmt.Errorf("token %q does not exist", id)
	}
	if t.place == nil {
		return fmt.Errorf("token %q is not in a place", id)
	}
	delete(r.tokensByID, id)
	delete(t.place.tokensByID, id)
	t.place = nil
	t.PlaceID = ""
	return nil
}

func (r *Runner) connectPointers() error {
	for _, t := range r.tokensByID {
		p, ok := r.placesByID[t.PlaceID]
		if !ok {
			return fmt.Errorf("token %q has invalid place id %q", t.ID, t.PlaceID)
		}
		t.place = p
		p.tokensByID[t.ID] = t
	}

	for _, a := range r.inputArcsByID {
		fp, ok := r.placesByID[a.FromID]
		if !ok {
			return fmt.Errorf("input arc %q has invalid from id %q", a.ID, a.FromID)
		}
		tt, ok := r.transitionsByID[a.ToID]
		if !ok {
			return fmt.Errorf("input arc %q has invalid to id %q", a.ID, a.ToID)
		}

		a.from = fp
		a.to = tt
		tt.inputPlaces = append(tt.inputPlaces, fp)
	}

	for _, a := range r.outputArcsByID {
		ft, ok := r.transitionsByID[a.FromID]
		if !ok {
			return fmt.Errorf("output arc %q has invalid from id %q", a.ID, a.FromID)
		}
		tp, ok := r.placesByID[a.ToID]
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

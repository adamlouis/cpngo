package petrinet

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
)

// Runner is a petri net runnner.
// It is an in-memory structure that implements petrinet behaviors.
type Runner struct {
	placesByID      map[PlaceID]*place
	transitionsByID map[TransitionID]*transition
	inputArcs       []*inputArc
	outputArcs      []*outputArc
	tokensByID      map[TokenID]*token
}

// runner is a private interface to help document public methods on the Runner struct
type runner interface {
	// Net returns the current Net state.
	Net() Net

	// Places returns a list of Places in the net.
	// This value should never change since Net topology may not change.
	Places() []Place

	Transitions() []Transition

	InputArcs() []InputArc

	OutputArcs() []OutputArc

	Tokens() []Token

	Enabled() []EnabledTransition

	Fire(fire EnabledTransition) (*Result, error)
	FireAny() (*Result, error)

	FireAsync(fire EnabledTransition) (*Result, error)
	FireResolve(id FireID) (*Result, error)
	FireReject(id FireID) (*Result, error)
}

// verify that Runner implements the private runner interface
var _ runner = (*Runner)(nil)

func NewRunner(n *Net) (*Runner, error) {
	ret := &Runner{
		placesByID:      map[PlaceID]*place{},
		transitionsByID: map[TransitionID]*transition{},
		inputArcs:       []*inputArc{},
		outputArcs:      []*outputArc{},
		tokensByID:      map[TokenID]*token{},
	}

	for _, p := range n.Places {
		ret.placesByID[p.ID] = &place{
			Place:      p,
			tokensByID: map[TokenID]*token{},
		}
	}

	for _, t := range n.Transitions {
		ret.transitionsByID[t.ID] = &transition{Transition: t}
	}

	for _, a := range n.InputArcs {
		ret.inputArcs = append(ret.inputArcs, &inputArc{InputArc: a})
	}

	for _, a := range n.OutputArcs {
		ret.outputArcs = append(ret.outputArcs, &outputArc{OutputArc: a})
	}

	for _, t := range n.Tokens {
		ret.tokensByID[t.ID] = &token{Token: t}
	}

	if err := ret.connectPointers(); err != nil {
		return nil, err
	}

	return ret, nil
}

type place struct {
	Place
	tokensByID map[TokenID]*token
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
	for _, a := range r.inputArcs {
		ret = append(ret, a.InputArc)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].String() < ret[j].String() })
	return ret
}

func (r *Runner) OutputArcs() []OutputArc {
	ret := []OutputArc{}
	for _, a := range r.outputArcs {
		ret = append(ret, a.OutputArc)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].String() < ret[j].String() })
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

func (r *Runner) Enabled() []EnabledTransition {
	ret := []EnabledTransition{}
	for _, t := range r.transitionsByID {
		enabled, err := r.isTransitionEnabled(t.ID)
		if err != nil {
			panic("unreachable: corrupted net: " + err.Error())
		}

		if enabled {
			// ret = append(ret, t.Transition)
		}
	}
	return ret
}

func (r *Runner) FireAsync(fire EnabledTransition) (*Result, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (r *Runner) FireResolve(id FireID) (*Result, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (r *Runner) FireReject(id FireID) (*Result, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (r *Runner) Fire(id EnabledTransition) (*Result, error) {

	// // TODO(adamlouis): implement execution polices
	// // TODO(adamlouis): implement guards

	// t, ok := r.transitionsByID[id]
	// if !ok {
	// 	return fmt.Errorf("transition %q does not exist", id)
	// }

	// enabled, err := r.isTransitionEnabled(id)
	// if err != nil {
	// 	return err
	// }
	// if !enabled {
	// 	return fmt.Errorf("transition %q is not enabled", id)
	// }

	// // TODO(adamlouis): no tokens may be enabled if arc expression is non-deterministic (e.g. rand())
	// consumeTokens := []*token{}
	// for _, p := range t.inputPlaces {
	// 	inputArc, err := r.getInputArc(p.ID, t.ID)
	// 	if err != nil {
	// 		return fmt.Errorf("transition %q failed to consume token: %w", id, err)
	// 	}
	// 	foundToken := false
	// 	for _, tk := range p.tokensByID {
	// 		ok, err := r.inputTokenOk(tk, inputArc)
	// 		if err != nil {
	// 			return fmt.Errorf("transition %q failed to consume token: %w", id, err)
	// 		}
	// 		if ok {
	// 			consumeTokens = append(consumeTokens, tk)
	// 			foundToken = true
	// 			break
	// 		}
	// 	}
	// 	if !foundToken {
	// 		return fmt.Errorf("transition %q failed to consume token", id)
	// 	}
	// }

	// consumedColors := []any{}
	// for _, tk := range consumeTokens {
	// 	if err := r.consumeToken(tk.ID); err != nil {
	// 		// TODO(adamlouis): if any fails, the net is corrupted & we should roll back
	// 		return fmt.Errorf("transition %q failed to consume token: %w", tk.ID, err)
	// 	}
	// 	consumedColors = append(consumedColors, tk.Color)
	// }

	// for _, p := range t.outputPlaces {
	// 	oa, err := r.getOutputArc(t.ID, p.ID)
	// 	if err != nil {
	// 		return fmt.Errorf("transition %q failed to produce token: %w", id, err)
	// 	}
	// 	color, err := nextColor(consumedColors, oa)
	// 	if err != nil {
	// 		return fmt.Errorf("transition %q failed to produce token when getting next color: %w", id, err)
	// 	}
	// 	tk := &token{
	// 		Token: Token{
	// 			ID:        TokenID(uuid.New().String()),
	// 			OnPlaceID: p.ID,
	// 			Color:     color,
	// 		},
	// 		place: p,
	// 	}
	// 	p.tokensByID[tk.ID] = tk
	// 	r.tokensByID[tk.ID] = tk
	// }

	// return nil

	return nil, fmt.Errorf("unimplemented")
}

func (r *Runner) FireAny() (*Result, error) {
	ts := r.Enabled()
	if len(ts) < 1 {
		return nil, fmt.Errorf("no transitions are enabled")
	}
	sort.Slice(ts, func(i, j int) bool {
		return ts[i].ID < ts[j].ID
	})
	return r.Fire(ts[0])
}

func (r *Runner) getInputArc(fromPlaceID PlaceID, toTransitionID TransitionID) (*inputArc, error) {
	for _, a := range r.inputArcs {
		if a.FromPlaceID == fromPlaceID && a.ToTransitionID == toTransitionID {
			return a, nil
		}
	}
	return nil, fmt.Errorf("input arc from %q to %q does not exist", fromPlaceID, toTransitionID)
}

func (r *Runner) getOutputArc(fromTransitionID TransitionID, toPlaceID PlaceID) (*outputArc, error) {
	for _, a := range r.outputArcs {
		if a.FromTransitionID == fromTransitionID && a.ToPlaceID == toPlaceID {
			return a, nil
		}
	}
	return nil, fmt.Errorf("output arc from %q to %q does not exist", fromTransitionID, toPlaceID)
}

func (r *Runner) isTransitionEnabled(id TransitionID) (bool, error) {
	t, ok := r.transitionsByID[id]
	if !ok {
		return false, fmt.Errorf("transition %q does not exist", id)
	}

	if len(t.inputPlaces) == 0 {
		panic("unreachable: corrupted net: transition has no input places")
	}

	for _, p := range t.inputPlaces {
		arc, err := r.getInputArc(p.ID, t.ID)
		if err != nil {
			return false, err
		}

		anyOk := false
		for _, tk := range p.tokensByID {
			ok, err := r.inputTokenOk(tk, arc)
			if err != nil {
				return false, err
			}
			if ok {
				anyOk = true
				break
			}
		}
		if !anyOk {
			return false, nil
		}
	}
	return true, nil
}

func compile(code string) (*vm.Program, error) {
	rand := expr.Function("rand", func(params ...any) (any, error) {
		return rand.Float64(), nil
	})
	return expr.Compile(code, rand)
}

func (r *Runner) inputTokenOk(tk *token, a *inputArc) (bool, error) {
	if a.Expr == "" {
		return true, nil
	}

	prog, err := compile(a.Expr)
	if err != nil {
		return false, fmt.Errorf("failed to compile expression %q: %w", a.Expr, err)
	}

	v, err := expr.Run(prog, map[string]any{
		"color": tk.Color,
	})
	if err != nil {
		return false, fmt.Errorf("failed to run expression %q: %w", a.Expr, err)
	}
	b, ok := v.(bool)
	if !ok {
		return false, fmt.Errorf("expression %q did not return a bool", a.Expr)
	}
	return b, nil
}

func nextColor(colors []any, a *outputArc) (any, error) {
	if a.Expr == "" {
		return nil, nil
	}

	prog, err := compile(a.Expr)
	if err != nil {
		return nil, fmt.Errorf("failed to compile expression %q: %w", a.Expr, err)
	}

	v, err := expr.Run(prog, map[string]any{
		"colors": colors,
	})
	fmt.Println("colors", colors, "expr", a.Expr, "v", v, "err", err)
	if err != nil {
		return nil, fmt.Errorf("failed to run expression %q: %w", a.Expr, err)
	}
	return v, nil
}

func (r *Runner) consumeToken(id TokenID) error {
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
	t.OnPlaceID = ""
	return nil
}

func (r *Runner) connectPointers() error {
	// add place & token pointers derived from token placement
	for _, token := range r.tokensByID {
		place, ok := r.placesByID[token.OnPlaceID]
		if !ok {
			return fmt.Errorf("token %q has invalid place id %q", token.ID, token.OnPlaceID)
		}
		token.place = place
		place.tokensByID[token.ID] = token
	}

	// add place & transition pointers derived from input arcs
	for _, arc := range r.inputArcs {
		place, ok := r.placesByID[arc.FromPlaceID]
		if !ok {
			return fmt.Errorf("input arc %q has invalid from id %q", arc.String(), arc.FromPlaceID)
		}

		transition, ok := r.transitionsByID[arc.ToTransitionID]
		if !ok {
			return fmt.Errorf("input arc %q has invalid to id %q", arc.String(), arc.ToTransitionID)
		}

		arc.from = place
		arc.to = transition
		transition.inputPlaces = append(transition.inputPlaces, place)
	}

	// add place & transition pointers derived from output arcs
	for _, arc := range r.outputArcs {
		place, ok := r.placesByID[arc.ToPlaceID]
		if !ok {
			return fmt.Errorf("output arc %q has invalid to id %q", arc.String(), arc.ToPlaceID)
		}

		transition, ok := r.transitionsByID[arc.FromTransitionID]
		if !ok {
			return fmt.Errorf("output arc %q has invalid from id %q", arc.String(), arc.FromTransitionID)
		}

		arc.from = transition
		arc.to = place
		transition.outputPlaces = append(transition.outputPlaces, place)
	}

	return nil
}

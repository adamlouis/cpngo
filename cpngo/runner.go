package cpngo

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
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

func (r *Runner) getInputArc(fromPlaceID, toTransitionID string) (*inputArc, error) {
	for _, a := range r.inputArcsByID {
		if a.FromID == fromPlaceID && a.ToID == toTransitionID {
			return a, nil
		}
	}
	return nil, fmt.Errorf("input arc from %q to %q does not exist", fromPlaceID, toTransitionID)
}
func (r *Runner) getOutputArc(fromTransitionID, toPlaceID string) (*outputArc, error) {
	for _, a := range r.outputArcsByID {
		if a.FromID == fromTransitionID && a.ToID == toPlaceID {
			return a, nil
		}
	}
	return nil, fmt.Errorf("output arc from %q to %q does not exist", fromTransitionID, toPlaceID)
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

	// TODO(adamlouis): no tokens may be enabled if arc expression is non-deterministic (e.g. rand())
	consumeTokens := []*token{}
	for _, p := range t.inputPlaces {
		inputArc, err := r.getInputArc(p.ID, t.ID)
		if err != nil {
			return fmt.Errorf("transition %q failed to consume token: %w", id, err)
		}
		foundToken := false
		for _, tk := range p.tokensByID {
			ok, err := r.inputTokenOk(tk, inputArc)
			if err != nil {
				return fmt.Errorf("transition %q failed to consume token: %w", id, err)
			}
			if ok {
				consumeTokens = append(consumeTokens, tk)
				foundToken = true
				break
			}
		}
		if !foundToken {
			return fmt.Errorf("transition %q failed to consume token", id)
		}
	}

	consumedColors := []any{}
	for _, tk := range consumeTokens {
		if err := r.consumeToken(tk.ID); err != nil {
			// TODO(adamlouis): if any fails, the net is corrupted & we should roll back
			return fmt.Errorf("transition %q failed to consume token: %w", tk.ID, err)
		}
		consumedColors = append(consumedColors, tk.Color)
	}

	for _, p := range t.outputPlaces {
		oa, err := r.getOutputArc(t.ID, p.ID)
		if err != nil {
			return fmt.Errorf("transition %q failed to produce token: %w", id, err)
		}
		color, err := nextColor(consumedColors, oa)
		if err != nil {
			return fmt.Errorf("transition %q failed to produce token when getting next color: %w", id, err)
		}
		tk := &token{
			Token: Token{
				ID:      uuid.New().String(),
				PlaceID: p.ID,
				Color:   color,
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

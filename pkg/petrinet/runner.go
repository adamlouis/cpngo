package petrinet

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/google/uuid"
)

// Runner is a petri net runnner.
// It is an in-memory structure that implements Petri Net behaviors.
type Runner struct {
	placesByID      map[PlaceID]*place
	transitionsByID map[TransitionID]*transition
	inputArcs       []*inputArc
	outputArcs      []*outputArc
	tokensByID      map[TokenID]*token

	pending map[FireID]Result
}

// runner is a private interface to help document public methods on the Runner struct
type runner interface {
	// Net returns the current Net state.
	Net() Net

	// Places returns a list of Places in the net.
	// The return value should never change since Net topology may never change.
	Places() []Place

	// Transitions returns a list of Transitions in the net.
	// The return value should never change since Net topology may never change.
	Transitions() []Transition

	// IntputArcs returns a list of input arcs in the net.
	// The return value should never change since Net topology may never change.
	InputArcs() []InputArc

	// OutputArcs returns a list of output arcs in the net.
	// The return value should never change since Net topology may never change.
	OutputArcs() []OutputArc

	// Tokens returns a list of Tokens in the net - the "marking".
	// The return value may change any time any Fire-related function is called.
	Tokens() []Token

	// Enabled returns a list of enabled transitions.
	// The enabled transitions includes the transition ID with all combinations of valid tokens.
	Enabled() []EnabledTransition

	// Fire executes the provided transtion immediately.
	// Tokens are immediately consumed & produced.
	Fire(fire EnabledTransition) (*Result, error)

	// FireAsync *starts* the firing process for the provided transition.
	// It consumes, but does not produce tokens.
	// FireAsync must be paired with FireResolve or FireReject to complete the transaction.
	// Result contains a FireID which servers as an ID for the Fire transaction.
	FireAsync(fire EnabledTransition) (*Result, error)

	// FireResolve completes the asynchronous transaction by producing tokens on the output places.
	FireResolve(id FireID) (*Result, error)

	// FireReject fails the pending transaction. No tokens are produced on the output places.
	FireReject(id FireID) (*Result, error)

	// Pending returns a list of pending calls to FireAsync.
	// It represents the FireAsync transactions that have not yet been resolved or rejected.
	Pending() []Result
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
		pending:         map[FireID]Result{},
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
	inputArcs  []*inputArc
	outputArcs []*outputArc
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

func (t *token) String() string {
	return string(t.ID)
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

func (r *Runner) Pending() []Result {
	ret := []Result{}
	for _, r := range r.pending {
		ret = append(ret, r)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i].FireID < ret[j].FireID })
	return ret
}

func (r *Runner) Enabled() []EnabledTransition {
	ret := []EnabledTransition{}
	for _, t := range r.transitionsByID {
		enabled, err := r.listEnabled(t.ID)
		if err != nil {
			panic("unreachable: corrupted net: " + err.Error())
		}
		ret = append(ret, enabled...)
	}
	return ret
}

func (r *Runner) FireAsync(enabled EnabledTransition) (*Result, error) {

	t, ok := r.transitionsByID[enabled.ID]
	if !ok {
		return nil, fmt.Errorf("transition %q does not exist", enabled.ID)
	}

	// TODO: enabled check must validate all input arcs to have valid token assignment with valid place!
	if err := r.validTransition(enabled); err != nil {
		return nil, fmt.Errorf("transition %q is not enabled: %w", t.ID, err)
	}

	ret := &Result{
		FireID:       newFireID(),
		TransitionID: enabled.ID,
		Consumed:     nil,
		Produced:     nil,
	}

	consumeTokens := []*token{}
	for _, token := range enabled.TokenAssignments {
		t, ok := r.tokensByID[token.ID]
		if !ok {
			return nil, fmt.Errorf("invalid token ID %q", t.ID)
		}
		consumeTokens = append(consumeTokens, t)
	}

	for _, tk := range consumeTokens {
		ret.Consumed = append(ret.Consumed, tk.Token)
		if err := r.consumeToken(tk.ID); err != nil {
			panic("consume token failed after validation, net is corrupted")
		}
	}

	r.pending[ret.FireID] = *ret

	return ret, nil
}

func (r *Runner) FireResolve(id FireID) (*Result, error) {
	ret, ok := r.pending[id]
	if !ok {
		return nil, fmt.Errorf("no pending fire %q", ret.FireID)
	}

	t, ok := r.transitionsByID[ret.TransitionID]
	if !ok {
		return nil, fmt.Errorf("invalid transtion")
	}

	consumedColors := []any{}
	for _, tk := range ret.Consumed {
		consumedColors = append(consumedColors, tk.Color)
	}

	for _, p := range t.outputPlaces {
		for _, oa := range p.outputArcs {
			if oa.FromTransitionID != t.ID {
				continue
			}

			color, err := nextColor(consumedColors, oa)
			if err != nil {
				return nil, fmt.Errorf("transition %q failed to produce token when getting next color: %w", t.ID, err)
			}

			tk := &token{
				Token: Token{
					ID:        newTokenID(),
					OnPlaceID: p.ID,
					Color:     color,
				},
				place: p,
			}

			p.tokensByID[tk.ID] = tk
			r.tokensByID[tk.ID] = tk
		}
	}

	return &ret, nil
}

func (r *Runner) FireReject(id FireID) (*Result, error) {
	ret, ok := r.pending[id]
	if !ok {
		return nil, fmt.Errorf("no pending fire %q", ret.FireID)
	}

	if _, ok := r.transitionsByID[ret.TransitionID]; !ok {
		return nil, fmt.Errorf("invalid transtion")
	}

	return &ret, nil
}

func newFireID() FireID {
	return FireID(TokenID(uuid.New().String()))
}

func newTokenID() TokenID {
	return TokenID(TokenID(uuid.New().String()))
}

// TODO: Expose Fire and FireWithAssignments as separate functions

func (r *Runner) Fire(enabled EnabledTransition) (*Result, error) {
	res, err := r.FireAsync(enabled)
	if err != nil {
		return nil, err
	}
	return r.FireResolve(res.FireID)
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

// listEnabled returns the enabled transitions for the transition ID including the cartesian product of tokens
func (r *Runner) listEnabled(id TransitionID) ([]EnabledTransition, error) {
	t, ok := r.transitionsByID[id]
	if !ok {
		return nil, fmt.Errorf("transition %q does not exist", id)
	}

	if len(t.inputPlaces) == 0 {
		return nil, fmt.Errorf("unreachable: corrupted net: transition has no input places")
	}

	inputArcs := []*inputArc{}
	tokensForInputArcs := [][]*token{}

	for _, place := range t.inputPlaces {
		for _, arc := range place.inputArcs {

			tokensForArc := []*token{}

			for _, tk := range place.tokensByID {
				ok, err := r.tokenSatisifesInputArcExpr(tk, arc)
				if err != nil {
					return nil, err
				}
				if ok {
					tokensForArc = append(tokensForArc, tk)
				}
			}

			if len(tokensForArc) == 0 {
				return []EnabledTransition{}, nil
			}

			inputArcs = append(inputArcs, arc)
			tokensForInputArcs = append(tokensForInputArcs, tokensForArc)
		}
	}

	combinations := CartestianProduct(tokensForInputArcs)

	ret := []EnabledTransition{}

	for _, combo := range combinations {
		assignments := map[InputArc]Token{}

		for j, t := range combo {
			assignments[inputArcs[j].InputArc] = t.Token
		}

		ret = append(ret, EnabledTransition{
			Transition:       t.Transition,
			TokenAssignments: assignments,
		})
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].String() < ret[j].String()
	})

	return ret, nil
}

func (r *Runner) validTransition(enabled EnabledTransition) error {
	all, err := r.listEnabled(enabled.ID)
	if err != nil {
		return err
	}
	for _, v := range all {
		if enabled.Equal(v) {
			return nil
		}
	}
	return fmt.Errorf("not a valid transition")
}

func compileExpr(code string) (*vm.Program, error) {
	rand := expr.Function("rand", func(params ...any) (any, error) {
		return rand.Float64(), nil
	})
	return expr.Compile(code, rand)
}

func (r *Runner) tokenSatisifesInputArcExpr(tk *token, a *inputArc) (bool, error) {
	if a.Expr == "" {
		return true, nil
	}

	prog, err := compileExpr(a.Expr)
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

	prog, err := compileExpr(a.Expr)
	if err != nil {
		return nil, fmt.Errorf("failed to compile expression %q: %w", a.Expr, err)
	}

	// TODO: this key should not be here
	v, err := expr.Run(prog, map[string]any{
		"colors": colors,
	})

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

		place.inputArcs = append(place.inputArcs, arc)
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

		place.outputArcs = append(place.outputArcs, arc)
	}

	return nil
}

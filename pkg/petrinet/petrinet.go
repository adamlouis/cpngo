package petrinet

import "fmt"

type Net struct {
	Places      []Place      `json:"places"`
	Transitions []Transition `json:"transitions"`
	InputArcs   []InputArc   `json:"input_arcs"`
	OutputArcs  []OutputArc  `json:"output_arcs"`
	Tokens      []Token      `json:"tokens"`
}

type PlaceID string
type TransitionID string
type TokenID string

type Place struct {
	ID PlaceID `json:"id"`
}

type Transition struct {
	ID TransitionID `json:"id"`
}

type InputArc struct {
	FromPlaceID    PlaceID      `json:"from_id"`
	ToTransitionID TransitionID `json:"to_id"`
	Expr           string       `json:"expr"`
}

func (ia *InputArc) String() string {
	return fmt.Sprintf("%s -> %s", ia.FromPlaceID, ia.ToTransitionID)
}

type OutputArc struct {
	FromTransitionID TransitionID `json:"from_id"`
	ToPlaceID        PlaceID      `json:"to_id"`
	Expr             string       `json:"expr"`
}

func (oa *OutputArc) String() string {
	return fmt.Sprintf("%s -> %s", oa.FromTransitionID, oa.ToPlaceID)
}

type Token struct {
	ID        TokenID `json:"id"`
	OnPlaceID PlaceID `json:"place_id"`
	Color     any     `json:"color"`
}

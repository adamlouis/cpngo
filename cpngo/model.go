package cpngo

type Net struct {
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
	Expr   string `json:"expr"`
	FromID string `json:"from_id"`
	ToID   string `json:"to_id"`
}
type OutputArc struct {
	ID     string `json:"id"`
	Expr   string `json:"expr"`
	FromID string `json:"from_id"`
	ToID   string `json:"to_id"`
}
type Token struct {
	ID      string `json:"id"`
	PlaceID string `json:"place_id"`
	Color   any    `json:"color"`
}

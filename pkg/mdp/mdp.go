package mdp

// mdp maps a petrinet to markov decision process semantics

// MDP is a "Markov Decision Process".
type MDP[State any, Action any] interface {
	// State returns the current state of the system.
	State() State

	// ActionSpace returns an exhaustive list of legal actions that may be taken in the current state.
	ActionSpace() []Action

	// Execute executes the provided action.
	// Executing an action may change the values of State() and ActionSpace()
	Execute(action Action)

	// Pending returns a list of pending actions
	Pending() []Action

	// Next advances to the completion of the next pending action.
	Next()
}

// Policy is an interface for implementing MDP policies.
type Policy[Action any] interface {
	// Policy accepts a list of actions representing the legal action space and returns an index
	// to the action that should be taken. -1 indicates that no action should be taken.
	Policy(actions []Action) int
}

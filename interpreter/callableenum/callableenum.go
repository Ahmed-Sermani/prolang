package callableenum

// helps to track if the current scope inside a function or not
// to fix behavior where return statement or 'this' expression
// are allowed outside of a function or method :).
const (
	NONE = iota
	FUNCTION
	METHOD
	// to identify initlizers and disallow return statement in them.
	INITIALIZER
)

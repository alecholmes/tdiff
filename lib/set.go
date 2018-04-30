package lib

// StringSet represents an unordered set of strings.
type StringSet map[string]bool

// Add puts the given values into the set.
func (s StringSet) Add(values ...string) {
	for _, v := range values {
		s[v] = true
	}
}

// Contains returns true iff the set contains the given value.
// s.contains("foo") and s["foo"] are equivalent.
func (s StringSet) Contains(value string) bool {
	return s[value]
}

// Slice returns an arbitrarily ordered slice that contains each value in the set.
func (s StringSet) Slice() []string {
	values := make([]string, 0, len(s))
	for v := range s {
		values = append(values, v)
	}

	return values
}

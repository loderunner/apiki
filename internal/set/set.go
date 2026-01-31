package set

import "encoding/json"

// Set is a generic set using map[T]struct{} for efficiency.
type Set[T comparable] map[T]struct{}

// New creates a Set with the given members.
func New[T comparable](members ...T) Set[T] {
	s := make(Set[T], len(members))
	for _, m := range members {
		s[m] = struct{}{}
	}
	return s
}

// Add inserts a value into the set.
func (s Set[T]) Add(value T) {
	s[value] = struct{}{}
}

// Remove deletes a value from the set.
func (s Set[T]) Remove(value T) {
	delete(s, value)
}

// Has returns true if the value is in the set.
func (s Set[T]) Has(value T) bool {
	_, ok := s[value]
	return ok
}

// Members returns all values in the set as a slice.
func (s Set[T]) Members() []T {
	members := make([]T, 0, len(s))
	for k := range s {
		members = append(members, k)
	}
	return members
}

// MarshalJSON marshals the set as a JSON array.
func (s Set[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Members())
}

// UnmarshalJSON unmarshals a JSON array into the set.
func (s *Set[T]) UnmarshalJSON(data []byte) error {
	var members []T
	if err := json.Unmarshal(data, &members); err != nil {
		return err
	}
	*s = New(members...)
	return nil
}

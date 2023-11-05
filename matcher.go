package v

type Matcher struct {
	field    string
	label    string
	value    any
	branches []branch
	fallback func(valuer *Valuer) error
	compare  func(a, b any) bool
}

type branch struct {
	value  any
	handle func(valuer *Valuer) error
}

func Match(value any, field, label string) *Matcher {
	return &Matcher{
		field:    field,
		label:    label,
		value:    value,
		branches: []branch{},
		fallback: nil,
		compare:  func(a, b any) bool { return a == b },
	}
}

func (m *Matcher) Branch(value any, handle func(valuer *Valuer) error) *Matcher {
	m.branches = append(m.branches, branch{value: value, handle: handle})
	return m
}

func (m *Matcher) Fallback(handle func(valuer *Valuer) error) *Matcher {
	m.fallback = handle
	return m
}

func (m *Matcher) Validate() error {
	for _, b := range m.branches {
		if m.compare(m.value, b.value) {
			valuer := Value(m.value, m.field, m.label)
			return b.handle(valuer)
		}
	}

	if m.fallback != nil {
		valuer := Value(m.value, m.field, m.label)
		return m.fallback(valuer)
	}

	return nil
}

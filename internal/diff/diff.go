package diff

//go:generate go-oneof diff.go

type oneofDiff struct {
	Type struct {
		Left  string
		Right string
	}
	Value   struct{}
	Missing struct{}

	Fields struct {
		Fields map[string]*oneofDiff
	}
	Indices struct {
		Left  map[int]*oneofDiff
		Right map[int]*oneofDiff
	}
	Keys struct {
		Left  map[any]*oneofDiff
		Right map[any]*oneofDiff
	}
}

// Diff is an interface to limit available implementations to partially replicate discriminated union type functionality
type Diff interface {
	isDiff()
}

// Type branch of Diff
type Type struct {
	Left  string
	Right string
}

func (*Type) isDiff() {}

// Value branch of Diff
type Value struct{}

func (*Value) isDiff() {}

// Missing branch of Diff
type Missing struct{}

func (*Missing) isDiff() {}

// Fields branch of Diff
type Fields struct {
	Fields map[string]Diff
}

func (*Fields) isDiff() {}

// Indices branch of Diff
type Indices struct {
	Left  map[int]Diff
	Right map[int]Diff
}

func (*Indices) isDiff() {}

// Keys branch of Diff
type Keys struct {
	Left  map[any]Diff
	Right map[any]Diff
}

func (*Keys) isDiff() {}

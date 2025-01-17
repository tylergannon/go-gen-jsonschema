package testapp0_simple

type ParentStruct struct {
	Inline struct {
		Bar           *int
		Baz           *string
		Coolio        *bool
		Child         ChildStruct
		Children      []ChildStruct
		Nieces        []*ChildStruct
		GrandChildren [][]GrandChildStruct
	}
	ChildStruct
	GoodKid ChildStruct
	BadKid  *ChildStruct
}

type ChildStruct struct {
	Inline struct {
		Parent *ParentStruct
		Bar    *int
		Bark   *string
		Coolio *bool
		GrandChildStruct
	}
}

type GrandChildStruct struct {
	Inline struct {
		Child *ChildStruct
		Bar   *int
	}
}

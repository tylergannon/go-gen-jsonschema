package testapp0_simple

type ParentStruct struct {
	Inline struct {
		Bar    *int
		Baz    *string
		Coolio *bool
		Child  ChildStruct
		//Children []ChildStruct
		//Nieces   []*ChildStruct
	}
	ChildStruct
	GoodKid ChildStruct
	BadKid  *ChildStruct
}

type ChildStruct struct {
	Inline struct {
		*ParentStruct
		Bar    *int
		Bark   *string
		Coolio *bool
		GrandChildStruct
	}
}

type GrandChildStruct struct {
	Inline struct {
		*ChildStruct
		Bar *int
	}
}

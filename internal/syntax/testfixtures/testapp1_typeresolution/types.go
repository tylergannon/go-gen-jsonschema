// nolint
package testapp1_typeresolution

type TrivialNamedType int
type TrivialNamedTypeString string
type TrivialNamedTypePtrInt *int
type TrivialNamedTypePtrString *string

// Not allowed
type IllegalChan chan int
type IllegalFunc func()
type IllegalInterface interface {
	IllegalStuff() int
}

type (
	ArrayType            [3]TrivialNamedType
	SliceType            []TrivialNamedType
	SliceOfSlice         [][]TrivialNamedType
	PointersArrayType    *[3]*TrivialNamedType
	PointersSliceType    *[]*TrivialNamedType
	PointersSliceOfSlice *[]*[]*TrivialNamedType
)

type (
	StructTypeSimple    struct{}
	StructTypeWithStuff struct {
		Foo string
		Bar int
	}
	PointersStructTypeWithStuff struct {
		Foo *string
		Bar *int
	}
	PointersStructTypeWithJsonTabs struct {
		Foo *string `json:"foo,omitempty"`
		Bar *int    `json:"bar,omitempty"`
	}
	PointersStructTypeWithIgnoreFields struct {
		bap  *int
		Foo  *string `json:"foo,omitempty"`
		baz  *string `json:"-"`
		Bar  *int    `json:"bar,omitempty"`
		Cool *int    `json:"-"`
	}
	EmbeddedStruct struct {
		Foo string
		Bar int
	}
	StructWithEmbeddings struct {
		Bat string
		EmbeddedStruct
	}

	StructWithInline struct {
		Foo int
		Bar struct {
			Foo int
			Bar string
		}
	}

	StructWithInlineAndNamed struct {
		Foo ArrayType
		Bar struct {
			Foo *SliceOfSlice
			Bar *[10]SliceOfSlice
		}
	}

	StructWithInlineAndNamedAllCrazy struct {
		Foo ArrayType
		Bar struct {
			Foo *SliceOfSlice
			Bar *[10]SliceOfSlice
			Baz *[]*[]*struct {
				Bark    SliceOfSlice
				Bite    int
				Recurse *StructWithInlineAndNamedAllCrazy
				StructWithInlineAndNamed
				Again struct {
					Foo int `json:"bat"`
				} `json:"boop"`
			} `json:"__nobody__"`
			nope *[]*[]*struct { // Note skipped d/t private field name
				Bark SliceOfSlice
			}
			Nope *[]*[]*struct { // Note this is going to be skipped d/t the json tag name
				Bark SliceOfSlice
				Foo  func()
				Bar  chan int
			} `json:"-"`
			StructWithInlineAndNamed
		}
	}

	IllegalStructWithInlineAndNamedAllCrazy struct {
		Foo ArrayType
		Bar struct {
			Foo *SliceOfSlice
			Bar *[10]SliceOfSlice
			Baz *[]*[]*struct {
				Bark    SliceOfSlice
				Bite    int
				BadBoiz chan struct{}
				Recurse *StructWithInlineAndNamedAllCrazy
				StructWithInlineAndNamed
				Again struct {
					Foo int `json:"bat"`
				} `json:"boop"`
			} `json:"__nobody__"`
			nope *[]*[]*struct { // Note skipped d/t private field name
				Bark SliceOfSlice
			}
			Nope *[]*[]*struct { // Note this is going to be skipped d/t the json tag name
				Bark SliceOfSlice
			} `json:"-"`
			StructWithInlineAndNamed
		}
	}
)

func (i PointersStructTypeWithIgnoreFields) GetBap() *int {
	return i.bap
}

func (i PointersStructTypeWithIgnoreFields) GetBaz() *string {
	return i.baz
}

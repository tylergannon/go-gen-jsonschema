package syntax

import "slices"

type SeenTypes []TypeID

func (s SeenTypes) Seen(t TypeID) bool {
	return slices.Contains(s, t.Concrete())
}

func (s SeenTypes) See(t TypeID) SeenTypes {
	return append(SeenTypes{t.Concrete()}, s...)
}

func (s SeenTypes) Add(t TypeID) (SeenTypes, bool) {
	if s.Seen(t) {
		return nil, false
	}
	return s.See(t), true
}

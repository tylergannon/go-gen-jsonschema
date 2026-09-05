package palette

type Level uint16

const (
	LevelLow  Level = 3
	LevelHigh Level = 8
)

func (Level) String() string { return "not-the-wire-name" }

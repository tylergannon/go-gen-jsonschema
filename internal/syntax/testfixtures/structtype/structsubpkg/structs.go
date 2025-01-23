package structsubpkg

// Sounder represents something that can make a sound
type Sounder interface {
	MakeSound() string
	GetVolume() float64
}

// Mover represents something that can move
type Mover interface {
	GetPosition() Position
	SetPosition(pos Position)
	GetSpeed() float64
}

// Animal represents a basic animal interface
type Animal interface {
	Sounder
	Mover
	GetSpecies() string
	GetAge() int
}

// Position represents a 3D position
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// Measurement represents a single data point
type Measurement struct {
	Value     float64 `json:"value"`
	Unit      string  `json:"unit"`
	Timestamp int64   `json:"timestamp"`
	Valid     bool    `json:"valid"`
}

// Sensor represents a device that takes measurements
type Sensor struct {
	ID             string        `json:"id"`
	Model          string        `json:"model"`
	SerialNumber   *string       `json:"serialNumber,omitempty"`
	Measurements   []Measurement `json:"measurements"`
	Position       *Position     `json:"position,omitempty"`
	LastCalibrated *int64        `json:"lastCalibrated,omitempty"`
}

// ExperimentConfig holds configuration for an experiment
type ExperimentConfig struct {
	Name       string   `json:"name"`
	Duration   int64    `json:"duration"`
	SampleRate float64  `json:"sampleRate"`
	Threshold  *float64 `json:"threshold,omitempty"`
}

// SuperStructure combines various types into a complex structure
type SuperStructure struct {
	// Basic metadata
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Active      bool    `json:"active"`
	Version     int     `json:"version"`

	// Complex fields
	MainSensor    *Sensor          `json:"mainSensor,omitempty"`
	BackupSensors []*Sensor        `json:"backupSensors,omitempty"`
	Config        ExperimentConfig `json:"config"`

	// Interface fields
	Animals  []Animal           `json:"animals"`
	Sounders map[string]Sounder `json:"sounders"`

	// Nested collections
	Positions   []*Position      `json:"positions,omitempty"`
	RawReadings [][]*Measurement `json:"rawReadings,omitempty"`
}

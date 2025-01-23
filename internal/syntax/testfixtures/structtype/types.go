package structtype

import "github.com/tylergannon/go-gen-jsonschema/internal/syntax/testfixtures/structtype/structsubpkg"

type ArrayOfSuperStruct []*structsubpkg.SuperStructure

// ExperimentRun represents a single run of an experiment with all its data
type ExperimentRun struct {
	structsubpkg.SuperStructure
	// Basic metadata
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Notes      *string `json:"notes,omitempty"`
	IsComplete bool    `json:"isComplete"`

	// Configuration and setup
	Location *structsubpkg.Position        `json:"location,omitempty"`
	Config   structsubpkg.ExperimentConfig `json:"config"`

	// Data collection
	PrimaryData   []structsubpkg.Measurement    `json:"primaryData"`
	SecondaryData [][]*structsubpkg.Measurement `json:"secondaryData,omitempty"`

	// Monitoring entities
	Observers      []structsubpkg.Animal  `json:"observers"`
	SoundSources   []structsubpkg.Sounder `json:"soundSources"`
	MovingElements []structsubpkg.Mover   `json:"movingElements"`

	// Equipment
	Equipment     []*structsubpkg.Sensor          `json:"equipment,omitempty"`
	BackupDevices map[string]*structsubpkg.Sensor `json:"backupDevices,omitempty"`

	// Reference positions
	Waypoints     []*structsubpkg.Position         `json:"waypoints,omitempty"`
	ControlPoints map[string]structsubpkg.Position `json:"controlPoints"`
}

package builder

// Transform is a hook to mutate the in-memory build model before rendering.
// It enables extensions without invasive changes to scanning or rendering.
type Transform interface {
	Apply(*SchemaBuilder) error
}

var registeredTransforms []Transform

// RegisterTransform registers a transform to run before rendering.
// This is optional and currently unused by core.
func RegisterTransform(t Transform) {
	registeredTransforms = append(registeredTransforms, t)
}

func (s *SchemaBuilder) applyTransforms() error {
	for _, t := range registeredTransforms {
		if err := t.Apply(s); err != nil {
			return err
		}
	}
	return nil
}

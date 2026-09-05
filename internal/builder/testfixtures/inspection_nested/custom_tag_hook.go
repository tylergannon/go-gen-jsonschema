//go:build custom

package inspection_nested

func (SavedTagHookModel) MarshalJSON() ([]byte, error) {
	return []byte(`"custom"`), nil
}

package util

// ValidatableConfig defines the validatable Config
type ValidatableConfig interface {
	// Validate is used to validate config and returns error on failure
	Validate() error
}

package structs

// NewReference instantiates a Reference with the default file mode.
func NewReference() *Reference {
	return &Reference{
		Previous: make([]byte, 32),
		Data:     make([]byte, 32),
		Mode:     0644,
	}
}

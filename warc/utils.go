const tagName = "validate"

// A Decoder reads a property list from an input stream.
type Decoder struct {
	// the format of the most-recently-decoded property list
	Format int

	reader io.ReadSeeker
	lax    bool
}

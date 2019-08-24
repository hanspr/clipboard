// Package clipboard read/write on clipboard
package clipboard

// ReadFrom read string from clipboard
func ReadAll(location string) (string, error) {
	return readFrom(location)
}

// WriteTo write string to clipboard
func WriteAll(text, location string) error {
	return writeAll(text, location)
}

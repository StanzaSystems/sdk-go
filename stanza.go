package stanza

// Init initializes the SDK with options. The returned error is non-nil if
// options is invalid or if FlowController can't be reached.
func Init(options ClientOptions) error {
	_, err := NewClient(options)
	if err != nil {
		return err
	}

	return nil
}

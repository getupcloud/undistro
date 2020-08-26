package helm

// ChartUnavailableError is returned when the requested chart is
// unavailable, and the reason is known and finite.
type ChartUnavailableError struct {
	Err error
}

func (err ChartUnavailableError) Unwrap() error {
	return err.Err
}

func (err ChartUnavailableError) Error() string {
	return "chart unavailable: " + err.Err.Error()
}

// ChartNotReadyError is returned when the requested chart is
// unavailable at the moment, but may become at available a later stage
// without any interference from a human.
type ChartNotReadyError struct {
	Err error
}

func (err ChartNotReadyError) Unwrap() error {
	return err.Err
}

func (err ChartNotReadyError) Error() string {
	return "chart not ready: " + err.Err.Error()
}

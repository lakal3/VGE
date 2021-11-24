package vk

// APIContext handles error on most of Vulkan calls that should not fail.
// Instead of passing error values of most unlike errors (typically programming errors), call will take APIContext and signal errors though that
type apicontext interface {
	// Sets serious, unexpected error in API call. You should most likely at least panic/recover to try to continue running application
	setError(err error)

	// Test if context is valid. Can be set by SetError.
	// Some operations will not even run if context is invalid, however support is not still in all places and using invalid context will more or less likely
	// result in crash.
	isValid() bool

	// Optional method called whenever API call is transferred to C++ DLL.
	// Method can optionally return a function that will be called when call is finished.
	// This can be used to trace or time API calls.
	begin(callName string) (atEnd func())
}

type errContext struct {
	err error
}

func (e *errContext) setError(err error) {
	e.err = err
}

func (e *errContext) isValid() bool {
	return e.err == nil
}

func (e *errContext) begin(callName string) (atEnd func()) {
	return nil
}

type nullContext struct {
}

func (n nullContext) setError(err error) {

}

func (n nullContext) isValid() bool {
	return true
}

func (n nullContext) begin(callName string) (atEnd func()) {
	return nil
}

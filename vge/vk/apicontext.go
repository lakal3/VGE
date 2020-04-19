package vk

// APIContext handles error on most of Vulkan calls that should not fail.
// Instead of passing error values of most unlike errors (typically programming errors), call will take APIContext and signal errors though that
type APIContext interface {
	// Sets serious, unexpected error in API call. You should most likely at least panic/recover to try to continue running application
	SetError(err error)

	// Test if context is valid. Can be set by SetError.
	// Some operations will not even run if context is invalid, however support is not still in all places and using invalid context will more or less likely
	// result in crash.
	IsValid() bool

	// Optional method called whenever API call is transferred to C++ DLL.
	// Method can optionally return a function that will be called when call is finished.
	// This can be used to trace or time API calls.
	Begin(callName string) (atEnd func())
}

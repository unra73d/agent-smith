// Package logger provides a simple leveled logging facility with tag support.
// It allows enabling/disabling logging globally for DEBUG, WARN, and ERROR levels,
// and also provides per-logger instance control over these levels.
// The package includes helper functions for checking errors and potentially panicking.
package logger

import "fmt"

// DEBUG controls the global enablement of debug level logging.
// Set to 1 to enable, 0 to disable. Affects log printouts only.
// This setting is applied *after* per-logger tag level flags.
const DEBUG = 1

// WARN controls the global enablement of warning level logging.
// Set to 1 to enable, 0 to disable. Affects log printouts only.
// This setting is applied *after* per-logger tag level flags.
const WARN = 1

// ERROR controls the global enablement of error level logging.
// Set to 1 to enable, 0 to disable. Affects log printouts only.
// Panic behavior in CheckE/CheckMultiE functions is *not* affected by this flag.
// This setting is applied *after* per-logger tag level flags.
const ERROR = 1

// logger represents a logging instance with a specific tag and level controls.
type logger struct {
	tag     string // Tag prepended to log messages for this logger instance.
	d, w, e int    // Level enable flags (1 for enabled, 0 for disabled) for Debug, Warn, Error.
}

// Logger creates and returns a new logger instance.
// Parameters:
//   - tag: A string identifier prepended to messages logged by this instance (e.g., "[Database]").
//   - d: Set to 1 to enable Debug level logging for this instance, 0 to disable.
//   - w: Set to 1 to enable Warn level logging for this instance, 0 to disable.
//   - e: Set to 1 to enable Error level logging for this instance, 0 to disable.
//
// Note: Global DEBUG, WARN, ERROR flags must also be enabled for messages to be printed.
func Logger(tag string, d int, w int, e int) *logger {
	return &logger{tag: tag, d: d, w: w, e: e}
}

// D logs a debug message if debug logging is enabled for this logger instance
// and globally. The logger's tag is automatically prepended.
func (self *logger) D(v ...interface{}) {
	if self.d == 1 {
		D(append([]interface{}{fmt.Sprintf("[%s]", self.tag)}, v...)...)
	}
}

// D logs a global debug message if global debug logging (DEBUG constant) is enabled.
// Arguments are printed space-separated, followed by a newline.
func D(v ...interface{}) {
	if DEBUG == 1 {
		fmt.Println("[DEBUG]", v)
	}
}

// W logs a warning message if warning logging is enabled for this logger instance
// and globally. The logger's tag is automatically prepended.
func (self *logger) W(v ...interface{}) {
	if self.w == 1 {
		W(append([]interface{}{fmt.Sprintf("[%s]", self.tag)}, v...)...)
	}
}

// W logs a global warning message if global warning logging (WARN constant) is enabled.
// Arguments are printed space-separated, followed by a newline.
func W(v ...interface{}) {
	if WARN == 1 {
		fmt.Println("[WARN]", v)
	}
}

// E logs an error message if error logging is enabled for this logger instance
// and globally. The logger's tag is automatically prepended.
func (self *logger) E(v ...interface{}) {
	if self.e == 1 {
		E(append([]interface{}{fmt.Sprintf("[%s]", self.tag)}, v...)...)
	}
}

// E logs a global error message if global error logging (ERROR constant) is enabled.
// Arguments are printed space-separated, followed by a newline.
func E(v ...interface{}) {
	if ERROR == 1 {
		fmt.Println("[ERR]", v)
	}
}

// CheckW checks if the provided error `err` is non-nil. If it is, and if
// warning logging is enabled for this logger instance and globally, it logs
// the error along with the provided arguments `v` (prepended by the logger's tag).
// Returns true if `err` is non-nil, false otherwise.
func (self *logger) CheckW(err error, v ...interface{}) bool {
	if self.w == 1 {
		return CheckW(err, append([]interface{}{fmt.Sprintf("[%s]", self.tag)}, v...)...)
	}
	// Still return whether an error occurred, even if logging is disabled.
	return err != nil
}

// CheckW checks if the provided error `err` is non-nil. If it is, and if
// global warning logging (WARN constant) is enabled, it logs the error along
// with the provided arguments `v`.
// Returns true if `err` is non-nil, false otherwise.
func CheckW(err error, v ...interface{}) bool {
	if err != nil {
		W(append(v, err)...) // Append err to the message arguments
	}

	return err != nil
}

// CheckE checks if the provided error `err` is non-nil.
// If `err` is non-nil:
//  1. If error logging is enabled for this logger instance and globally, it logs
//     the error along with the provided arguments `v` (prepended by the logger's tag).
//  2. It then panics with the error `err`.
//
// If error logging is disabled for this instance but `err` is non-nil, it still panics.
func (self *logger) CheckE(err error, finalCall func(), v ...interface{}) {
	if err != nil { // Check for error first to ensure panic happens regardless of log level
		if self.e == 1 && ERROR == 1 { // Check both instance and global flags for logging
			// Use the global E function to handle the actual print logic
			E(append([]interface{}{fmt.Sprintf("[%s]", self.tag)}, append(v, err)...)...)
		}
		if finalCall != nil {
			finalCall()
		}
		panic(err) // Panic regardless of whether it was logged
	}
}

// CheckE checks if the provided error `err` is non-nil.
// If `err` is non-nil:
//  1. If global error logging (ERROR constant) is enabled, it logs the error
//     along with the provided arguments `v`.
//  2. It then panics with the error `err`.
func CheckE(err error, finalCall func(), v ...interface{}) {
	if err != nil {
		E(append(v, err)...) // Append err to the message arguments
		if finalCall != nil {
			finalCall()
		}
		panic(err)
	}
}

// CheckMultiE checks if the provided slice of errors `err` contains any non-nil errors.
// If it finds non-nil errors:
//  1. If error logging is enabled for this logger instance and globally, it logs
//     each non-nil error along with the provided arguments `v` (prepended by the logger's tag).
//  2. It then panics with the *first* non-nil error encountered in the slice.
//
// If error logging is disabled for this instance but non-nil errors exist, it still panics
// with the first non-nil error.
func (self *logger) CheckMultiE(errs []error, finalCall func(), v ...interface{}) {
	firstErr := findFirstError(errs)
	if firstErr != nil { // Check for error first
		if self.e == 1 && ERROR == 1 { // Check both instance and global flags for logging
			// Log each non-nil error individually
			prefix := fmt.Sprintf("[%s]", self.tag)
			for _, err := range errs {
				if err != nil {
					// Use the global E function to handle the actual print logic
					E(append([]interface{}{prefix}, append(v, err)...)...)
				}
			}
		}
		if finalCall != nil {
			finalCall()
		}
		panic(firstErr) // Panic with the first error found
	}
}

// CheckMultiE checks if the provided slice of errors `errs` contains any non-nil errors.
// If it finds non-nil errors:
//  1. If global error logging (ERROR constant) is enabled, it logs each non-nil
//     error along with the provided arguments `v`.
//  2. It then panics with the *first* non-nil error encountered in the slice.
func CheckMultiE(errs []error, finalCall func(), v ...interface{}) {
	firstErr := findFirstError(errs)
	if firstErr != nil {
		if ERROR == 1 {
			// Log each non-nil error individually
			for _, err := range errs {
				if err != nil {
					E(append(v, err)...) // Append err to the message arguments
				}
			}
		}
		if finalCall != nil {
			finalCall()
		}
		panic(firstErr) // Panic with the first error found
	}
}

// findFirstError is a helper function to get the first non-nil error from a slice.
// Returns nil if the slice is empty or contains only nil errors.
func findFirstError(errs []error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

// BreakOnError is intended to be used with `defer` in functions where panics
// generated by CheckE or CheckMultiE should be caught. It recovers from the
// panic and logs the recovered error using the global E function.
// This effectively stops the panicking flow and allows the calling function
// to return, preventing program termination.
// Example:
//
//	func myOperation() {
//	    log := log.Logger("MyOp", 1, 1, 1)
//	    defer logger.BreakOnError()
//
//	    _, err := potentiallyFailingCall()
//	    log.CheckE(err, "Failed during potentially failing call")
//
//	    // Code here won't execute if CheckE panics
//	}
func BreakOnError() {
	if r := recover(); r != nil {
		// Attempt to log the recovered value as an error.
		// This assumes the panic was likely caused by an error passed to CheckE/CheckMultiE.
		E("Recovered from panic:", r)
	}
}

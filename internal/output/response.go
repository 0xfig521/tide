package output

import (
	"encoding/json"
	"fmt"
	"os"
)

// ErrorCode is a stable machine-readable error identifier.
type ErrorCode string

const (
	CodeInvalidArgs       ErrorCode = "invalid_args"
	CodeDBOpenFailed      ErrorCode = "db_open_failed"
	CodeFeedAlreadyExists ErrorCode = "feed_already_exists"
	CodeFeedNotFound      ErrorCode = "feed_not_found"
	CodeEntryNotFound     ErrorCode = "entry_not_found"
	CodeFetchFailed       ErrorCode = "fetch_failed"
	CodeParseFailed       ErrorCode = "parse_failed"
	CodeCategoryNotFound  ErrorCode = "category_not_found"
	CodeInternalError     ErrorCode = "internal_error"
	CodeAlreadyExists     ErrorCode = "already_exists"
)

// Response is the standard JSON envelope for all CLI output.
// Only stdout receives JSON. Progress, logs, and diagnostics go to stderr.
type Response struct {
	OK    bool       `json:"ok"`
	Data  any        `json:"data"`
	Error *ErrorInfo `json:"error"`
	Meta  any        `json:"meta"`
}

// ErrorInfo is the structured error payload in a Response.
type ErrorInfo struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

// CmdError wraps an ErrorCode + message with a Cobra-compatible error interface.
// Commands using RunE should return this to signal failure with a non-zero exit code.
type CmdError struct {
	Code    ErrorCode
	Message string
}

func (e *CmdError) Error() string {
	return string(e.Code) + ": " + e.Message
}

// NewCmdError creates a CmdError for commands to return via RunE.
func NewCmdError(code ErrorCode, msg string) *CmdError {
	return &CmdError{Code: code, Message: msg}
}

// PrintSuccess writes a success Response to stdout.
// meta is optional (nil means omitted from JSON).
func PrintSuccess(data any, meta any) {
	resp := Response{
		OK:   true,
		Data: data,
		Meta: meta,
	}
	writeJSON(resp)
}

// PrintError writes an error Response to stdout AND returns a CmdError
// so that Cobra's RunE will cause a non-zero exit code.
// This is the single function commands should call on failure.
func PrintError(code ErrorCode, msg string) *CmdError {
	resp := Response{
		OK:    false,
		Error: &ErrorInfo{Code: code, Message: msg},
	}
	writeJSON(resp)
	return NewCmdError(code, msg)
}

// PrintTable writes a plain string to stdout (no JSON wrapping).
// Only used when --format table is explicitly requested.
func PrintTable(s string) {
	fmt.Println(s)
}

// PrintPlain writes a plain string to stdout.
// For commands that intentionally produce non-JSON output (e.g., schedule logs).
func PrintPlain(s string) {
	fmt.Println(s)
}

// FatalError writes an error message to stderr and returns a CmdError.
// Use this for pre-command initialization errors where JSON to stdout
// would be ambiguous (e.g., DB open failure before command execution).
func FatalError(msg string) *CmdError {
	fmt.Fprintln(os.Stderr, ErrorMsg(msg))
	return NewCmdError(CodeInternalError, msg)
}

func writeJSON(v any) {
	b, err := json.Marshal(v)
	if err != nil {
		// Fallback: write a minimal error envelope
		fallback := fmt.Sprintf(`{"ok":false,"error":{"code":"%s","message":"json marshal failed"}}`, CodeInternalError)
		fmt.Println(fallback)
		return
	}
	fmt.Println(string(b))
}

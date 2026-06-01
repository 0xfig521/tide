package output

import (
	"encoding/csv"
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

// PrintJSONL writes each item as a separate JSON line to stdout.
// Empty or nil slices produce no output.
// If an item fails to marshal, a minimal error line is written and processing continues.
func PrintJSONL(items []any) {
	for _, item := range items {
		b, err := json.Marshal(item)
		if err != nil {
			fmt.Fprintf(os.Stdout, `{"error":"json marshal failed: %s"}`+"\n", err.Error())
			continue
		}
		fmt.Fprintln(os.Stdout, string(b))
	}
}

// PrintJSONLItems is a generic version of PrintJSONL that accepts any slice type.
// Empty or nil slices produce no output.
func PrintJSONLItems[T any](items []T) {
	for _, item := range items {
		b, err := json.Marshal(item)
		if err != nil {
			fmt.Fprintf(os.Stdout, `{"error":"json marshal failed: %s"}`+"\n", err.Error())
			continue
		}
		fmt.Fprintln(os.Stdout, string(b))
	}
}

// PrintCSV writes CSV to stdout using encoding/csv.
// Writes headers first, then rows. Flushes and checks for errors.
// Empty headers with nil/empty rows produce no output.
func PrintCSV(headers []string, rows [][]string) {
	if len(headers) == 0 {
		return
	}
	w := csv.NewWriter(os.Stdout)
	_ = w.Write(headers)
	for _, row := range rows {
		_ = w.Write(row)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		fmt.Fprintf(os.Stderr, "csv write error: %v\n", err)
	}
}

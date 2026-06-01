package output

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStdout runs fn and returns what was written to stdout.
func captureStdout(fn func()) string {
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = orig

	out, _ := io.ReadAll(r)
	return string(out)
}

func TestPrintSuccess_JSONContract(t *testing.T) {
	type testData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	type testMeta struct {
		Page  int `json:"page"`
		Total int `json:"total"`
	}

	data := testData{Name: "hello", Value: 42}
	meta := testMeta{Page: 1, Total: 100}

	output := captureStdout(func() {
		PrintSuccess(data, meta)
	})

	var resp Response
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v\nraw output: %s", err, output)
	}

	if !resp.OK {
		t.Error("expected ok=true")
	}
	if resp.Error != nil {
		t.Errorf("expected error=null, got %+v", resp.Error)
	}
	if resp.Data == nil {
		t.Error("expected data to be non-nil")
	}
	if resp.Meta == nil {
		t.Error("expected meta to be non-nil")
	}

	// Verify data content by re-marshaling.
	dataJSON, _ := json.Marshal(resp.Data)
	var parsed testData
	if err := json.Unmarshal(dataJSON, &parsed); err != nil {
		t.Fatalf("failed to parse data: %v", err)
	}
	if parsed.Name != "hello" || parsed.Value != 42 {
		t.Errorf("data mismatch: %+v", parsed)
	}

	// Verify meta content.
	metaJSON, _ := json.Marshal(resp.Meta)
	var parsedMeta testMeta
	if err := json.Unmarshal(metaJSON, &parsedMeta); err != nil {
		t.Fatalf("failed to parse meta: %v", err)
	}
	if parsedMeta.Page != 1 || parsedMeta.Total != 100 {
		t.Errorf("meta mismatch: %+v", parsedMeta)
	}
}

func TestPrintSuccess_NilDataAndMeta(t *testing.T) {
	output := captureStdout(func() {
		PrintSuccess(nil, nil)
	})

	var resp Response
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	if !resp.OK {
		t.Error("expected ok=true")
	}
	if resp.Error != nil {
		t.Error("expected error to be null")
	}
	// data can be null when nil is passed.
	if resp.Meta != nil {
		t.Error("expected meta to be null")
	}
}

func TestPrintError_JSONContract(t *testing.T) {
	output := captureStdout(func() {
		PrintError(CodeFeedNotFound, "feed 42 not found")
	})

	var resp Response
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v\nraw output: %s", err, output)
	}

	if resp.OK {
		t.Error("expected ok=false")
	}
	if resp.Data != nil {
		t.Errorf("expected data=null, got %v", resp.Data)
	}
	if resp.Error == nil {
		t.Fatal("expected error to be non-nil")
	}
	if resp.Error.Code != CodeFeedNotFound {
		t.Errorf("error.code = %q, want %q", resp.Error.Code, CodeFeedNotFound)
	}
	if resp.Error.Message != "feed 42 not found" {
		t.Errorf("error.message = %q, want %q", resp.Error.Message, "feed 42 not found")
	}
	if resp.Meta != nil {
		t.Errorf("expected meta=null, got %v", resp.Meta)
	}
}

func TestPrintError_ReturnsCmdError(t *testing.T) {
	output := captureStdout(func() {
		PrintError(CodeInvalidArgs, "bad input")
	})

	// Verify stdout has JSON.
	var resp Response
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("stdout is not valid JSON: %v", err)
	}

	// PrintError returns *CmdError (concrete type), which implements error.
	cmdErr := PrintError(CodeInvalidArgs, "bad input")
	if cmdErr == nil {
		t.Fatal("PrintError should return a non-nil *CmdError")
	}
	if cmdErr.Code != CodeInvalidArgs {
		t.Errorf("CmdError.Code = %q, want %q", cmdErr.Code, CodeInvalidArgs)
	}
	if cmdErr.Message != "bad input" {
		t.Errorf("CmdError.Message = %q, want %q", cmdErr.Message, "bad input")
	}
	// Verify it satisfies the error interface.
	var _ error = cmdErr
}

func TestPrintError_AllErrorCodes(t *testing.T) {
	codes := []ErrorCode{
		CodeInvalidArgs,
		CodeDBOpenFailed,
		CodeFeedAlreadyExists,
		CodeFeedNotFound,
		CodeEntryNotFound,
		CodeFetchFailed,
		CodeParseFailed,
		CodeCategoryNotFound,
		CodeInternalError,
		CodeAlreadyExists,
	}

	for _, code := range codes {
		output := captureStdout(func() {
			PrintError(code, "test message")
		})

		var resp Response
		if err := json.Unmarshal([]byte(output), &resp); err != nil {
			t.Errorf("code %q: invalid JSON: %v", code, err)
			continue
		}
		if resp.OK {
			t.Errorf("code %q: expected ok=false", code)
		}
		if resp.Error == nil || resp.Error.Code != code {
			t.Errorf("code %q: error.code = %q, want %q", code, resp.Error.Code, code)
		}
	}
}

func TestNewCmdError(t *testing.T) {
	cmdErr := NewCmdError(CodeInternalError, "something broke")

	if cmdErr == nil {
		t.Fatal("NewCmdError should return non-nil")
	}
	if cmdErr.Code != CodeInternalError {
		t.Errorf("code = %q, want %q", cmdErr.Code, CodeInternalError)
	}
	if cmdErr.Message != "something broke" {
		t.Errorf("message = %q, want %q", cmdErr.Message, "something broke")
	}

	expectedErrStr := "internal_error: something broke"
	if cmdErr.Error() != expectedErrStr {
		t.Errorf("Error() = %q, want %q", cmdErr.Error(), expectedErrStr)
	}
	// Verify it satisfies the error interface.
	var _ error = cmdErr
}

func TestPrintJSONL_Basic(t *testing.T) {
	items := []any{
		map[string]any{"id": 1, "title": "first"},
		map[string]any{"id": 2, "title": "second"},
	}

	output := captureStdout(func() {
		PrintJSONL(items)
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d\noutput: %s", len(lines), output)
	}

	for i, line := range lines {
		if !json.Valid([]byte(line)) {
			t.Errorf("line %d is not valid JSON: %s", i, line)
			continue
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			t.Errorf("line %d unmarshal failed: %v", i, err)
		}
	}

	var obj1 map[string]any
	json.Unmarshal([]byte(lines[0]), &obj1)
	if obj1["id"] != float64(1) || obj1["title"] != "first" {
		t.Errorf("line 0 content mismatch: %v", obj1)
	}
}

func TestPrintJSONL_Empty(t *testing.T) {
	output := captureStdout(func() {
		PrintJSONL([]any{})
	})
	if output != "" {
		t.Errorf("expected empty output for empty slice, got: %q", output)
	}

	output = captureStdout(func() {
		PrintJSONL(nil)
	})
	if output != "" {
		t.Errorf("expected empty output for nil slice, got: %q", output)
	}
}

func TestPrintJSONL_MarshalError(t *testing.T) {
	ch := make(chan int)
	items := []any{
		map[string]any{"id": 1, "name": "good"},
		ch,
		map[string]any{"id": 3, "name": "also good"},
	}

	output := captureStdout(func() {
		PrintJSONL(items)
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d\noutput: %s", len(lines), output)
	}

	var obj1 map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &obj1); err != nil {
		t.Errorf("line 0 is not valid JSON: %v", err)
	}

	var errObj map[string]string
	if err := json.Unmarshal([]byte(lines[1]), &errObj); err != nil {
		t.Errorf("line 1 is not valid JSON: %v", err)
	}
	if _, hasError := errObj["error"]; !hasError {
		t.Errorf("line 1 should contain 'error' key, got: %v", errObj)
	}

	var obj3 map[string]any
	if err := json.Unmarshal([]byte(lines[2]), &obj3); err != nil {
		t.Errorf("line 2 is not valid JSON: %v", err)
	}
}

type jsonlTestItem struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

func TestPrintJSONLItems_WithStruct(t *testing.T) {
	items := []jsonlTestItem{
		{ID: 1, Title: "alpha"},
		{ID: 2, Title: "beta"},
	}

	output := captureStdout(func() {
		PrintJSONLItems(items)
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d\noutput: %s", len(lines), output)
	}

	for i, line := range lines {
		var item jsonlTestItem
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			t.Errorf("line %d parse failed: %v", i, err)
		}
	}

	var item1 jsonlTestItem
	json.Unmarshal([]byte(lines[0]), &item1)
	if item1.ID != 1 || item1.Title != "alpha" {
		t.Errorf("line 0 content mismatch: %+v", item1)
	}

	output = captureStdout(func() {
		PrintJSONLItems([]jsonlTestItem{})
	})
	if output != "" {
		t.Errorf("expected empty output for empty slice, got: %q", output)
	}
}

func TestPrintCSV_Basic(t *testing.T) {
	output := captureStdout(func() {
		PrintCSV([]string{"id", "name"}, [][]string{
			{"1", "Alice"},
			{"2", "Bob"},
		})
	})

	r := csv.NewReader(strings.NewReader(output))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to read CSV: %v", err)
	}

	if len(records) != 3 {
		t.Fatalf("expected 3 records (1 header + 2 rows), got %d", len(records))
	}
	if records[0][0] != "id" || records[0][1] != "name" {
		t.Errorf("header mismatch: %v", records[0])
	}
	if records[1][0] != "1" || records[1][1] != "Alice" {
		t.Errorf("row 1 mismatch: %v", records[1])
	}
	if records[2][0] != "2" || records[2][1] != "Bob" {
		t.Errorf("row 2 mismatch: %v", records[2])
	}
}

func TestPrintCSV_EmptyHeaders(t *testing.T) {
	output := captureStdout(func() {
		PrintCSV(nil, [][]string{{"1", "data"}})
	})
	if output != "" {
		t.Errorf("expected no output for nil headers, got: %q", output)
	}

	output = captureStdout(func() {
		PrintCSV([]string{}, [][]string{{"1", "data"}})
	})
	if output != "" {
		t.Errorf("expected no output for empty headers, got: %q", output)
	}
}

func TestPrintCSV_NoRows(t *testing.T) {
	output := captureStdout(func() {
		PrintCSV([]string{"id", "name"}, nil)
	})

	r := csv.NewReader(strings.NewReader(output))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to read CSV: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record (header only), got %d", len(records))
	}
	if records[0][0] != "id" || records[0][1] != "name" {
		t.Errorf("header mismatch: %v", records[0])
	}

	output = captureStdout(func() {
		PrintCSV([]string{"a", "b"}, [][]string{})
	})
	r = csv.NewReader(strings.NewReader(output))
	records, err = r.ReadAll()
	if err != nil {
		t.Fatalf("failed to read CSV with empty rows: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record (header only), got %d", len(records))
	}
}

func TestPrintError_NonZeroExitCodeHint(t *testing.T) {
	tests := []struct {
		code    ErrorCode
		message string
	}{
		{CodeInvalidArgs, "short"},
		{CodeInternalError, ""},
		{CodeFetchFailed, "special chars: <>&\"'"},
		{CodeDBOpenFailed, "multi\nline\nmessage"},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			_ = captureStdout(func() {
				_ = PrintError(tt.code, tt.message)
			})

			cmdErr := PrintError(tt.code, tt.message)
			if cmdErr == nil {
				t.Fatal("PrintError should return a non-nil *CmdError for non-zero exit code semantics")
			}
			if cmdErr.Code != tt.code {
				t.Errorf("CmdError.Code = %q, want %q", cmdErr.Code, tt.code)
			}
			errStr := cmdErr.Error()
			if errStr == "" {
				t.Error("Error() should not return an empty string")
			}
		})
	}
}

func TestResponse_SchemaConsistency(t *testing.T) {
	requiredFields := []string{"ok", "data", "error", "meta"}

	t.Run("PrintSuccess envelope", func(t *testing.T) {
		testCases := []struct {
			name string
			data any
			meta any
		}{
			{"string_data", "value", nil},
			{"int_data", 42, map[string]int{"count": 1}},
			{"nil_both", nil, nil},
			{"struct_data", jsonlTestItem{ID: 1, Title: "test"}, nil},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				output := captureStdout(func() {
					PrintSuccess(tc.data, tc.meta)
				})

				var resp Response
				if err := json.Unmarshal([]byte(output), &resp); err != nil {
					t.Fatalf("not valid JSON: %v", err)
				}
				if !resp.OK {
					t.Error("ok should be true")
				}

				raw := make(map[string]any)
				json.Unmarshal([]byte(output), &raw)
				for _, f := range requiredFields {
					if _, exists := raw[f]; !exists {
						t.Errorf("missing field %q", f)
					}
				}
			})
		}
	})

	t.Run("PrintError envelope", func(t *testing.T) {
		testCases := []struct {
			name    string
			code    ErrorCode
			message string
		}{
			{"normal", CodeInvalidArgs, "bad input"},
			{"empty_message", CodeInternalError, ""},
			{"special_chars", CodeFeedNotFound, "feed \"42\" not <found>"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				output := captureStdout(func() {
					PrintError(tc.code, tc.message)
				})

				var resp Response
				if err := json.Unmarshal([]byte(output), &resp); err != nil {
					t.Fatalf("not valid JSON: %v", err)
				}
				if resp.OK {
					t.Error("ok should be false")
				}
				if resp.Error == nil {
					t.Fatal("error should be non-nil")
				}

				raw := make(map[string]any)
				json.Unmarshal([]byte(output), &raw)
				for _, f := range requiredFields {
					if _, exists := raw[f]; !exists {
						t.Errorf("missing field %q", f)
					}
				}
			})
		}
	})
}

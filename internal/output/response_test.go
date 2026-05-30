package output

import (
	"encoding/json"
	"io"
	"os"
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

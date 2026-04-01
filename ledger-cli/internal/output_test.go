package ledger

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestAppErrorHelpers(t *testing.T) {
	t.Parallel()

	invalid := NewInvalidArgumentError("bad input")
	if invalid.Code != CodeInvalidArgument || invalid.Error() != "bad input" {
		t.Fatalf("NewInvalidArgumentError() = %+v", invalid)
	}

	notFound := NewNotFoundError("missing")
	if notFound.Code != CodeNotFound || notFound.Error() != "missing" {
		t.Fatalf("NewNotFoundError() = %+v", notFound)
	}

	cause := errors.New("disk full")
	internal := NewInternalError("write failed", cause)
	if internal.Code != CodeInternal || internal.Error() != "write failed" {
		t.Fatalf("NewInternalError() = %+v", internal)
	}
	if !errors.Is(internal, cause) {
		t.Fatalf("Unwrap() did not expose the wrapped error")
	}

	var nilErr *AppError
	if nilErr.Error() != "" {
		t.Fatalf("nil Error() = %q", nilErr.Error())
	}
	if nilErr.Unwrap() != nil {
		t.Fatalf("nil Unwrap() = %v", nilErr.Unwrap())
	}
}

func TestResponses(t *testing.T) {
	t.Parallel()

	success := SuccessResponse(map[string]string{"id": "entry-1"})
	if !success.Success || success.Error != nil {
		t.Fatalf("SuccessResponse() = %+v", success)
	}

	notFound := ErrorResponse(NewNotFoundError("missing"))
	if notFound.Success || notFound.Error == nil {
		t.Fatalf("ErrorResponse(not found) = %+v", notFound)
	}
	if notFound.Error.Code != CodeNotFound || notFound.Error.Message != "missing" {
		t.Fatalf("ErrorResponse(not found) = %+v", notFound)
	}

	internal := ErrorResponse(errors.New("boom"))
	if internal.Error == nil || internal.Error.Code != CodeInternal || internal.Error.Message != "internal error" {
		t.Fatalf("ErrorResponse(internal) = %+v", internal)
	}
}

func TestWriteResponse(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer
	if err := WriteResponse(&buffer, SuccessResponse(map[string]string{"id": "entry-1"})); err != nil {
		t.Fatalf("WriteResponse() error = %v", err)
	}

	output := buffer.String()
	if !strings.Contains(output, "\n  \"success\"") {
		t.Fatalf("WriteResponse() output = %q", output)
	}

	var response Response
	if err := json.Unmarshal(buffer.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if !response.Success {
		t.Fatalf("response = %+v", response)
	}
}

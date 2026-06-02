package models

import "strings"

// ClassifyError maps a raw error string and HTTP status (when known) into
// a structured FailureType plus the corresponding HTTP status code.
// errStr is the wrapped error chain from fetcher.Parser.Fetch (already
// stringified); statusCode is 0 when the request never completed.
func ClassifyError(errStr string, statusCode int) (FailureType, int) {
	if statusCode >= 400 && statusCode < 500 {
		return FailureHTTP4xx, statusCode
	}
	if statusCode >= 500 && statusCode < 600 {
		return FailureHTTP5xx, statusCode
	}

	low := strings.ToLower(errStr)
	switch {
	case strings.Contains(low, "timeout exceeded"),
		strings.Contains(low, "deadline exceeded"),
		strings.Contains(low, "i/o timeout"):
		return FailureTimeout, 0
	case strings.Contains(low, "no such host"),
		strings.Contains(low, "dns lookup"),
		strings.Contains(low, "no address associated"):
		return FailureDNS, 0
	case strings.Contains(low, "tls"),
		strings.Contains(low, "certificate"),
		strings.Contains(low, "handshake"),
		strings.Contains(low, "x509"):
		return FailureTLS, 0
	case strings.Contains(low, "parse feed"):
		return FailureParse, 0
	}

	return FailureUnknown, 0
}

// IsTransient reports whether the failure type usually resolves on its own.
func IsTransient(t FailureType) bool {
	switch t {
	case FailureHTTP5xx, FailureTimeout, FailureDNS:
		return true
	default:
		return false
	}
}

// IsPermanent reports whether the failure type almost never resolves
// without human intervention (dead URLs, removed feeds, TLS misconfig).
func IsPermanent(t FailureType) bool {
	switch t {
	case FailureHTTP4xx, FailureParse, FailureTLS:
		return true
	default:
		return false
	}
}

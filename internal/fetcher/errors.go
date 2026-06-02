package fetcher

import (
	"github.com/0xfig521/tide/internal/models"
)

// ClassifyError re-exports models.ClassifyError for fetcher callers.
func ClassifyError(errStr string, statusCode int) (models.FailureType, int) {
	return models.ClassifyError(errStr, statusCode)
}

// IsTransient reports whether the failure type usually resolves on its own.
func IsTransient(t models.FailureType) bool { return models.IsTransient(t) }

// IsPermanent reports whether the failure type almost never resolves
// without human intervention (dead URLs, removed feeds, TLS misconfig).
func IsPermanent(t models.FailureType) bool { return models.IsPermanent(t) }

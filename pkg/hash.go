package pkg

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// EntryHash generates a deterministic hash for deduplication.
// Uses feedID + guid to ensure uniqueness within a feed.
func EntryHash(feedID int64, guid string) string {
	h := sha256.New()
	h.Write(fmt.Appendf(nil, "%d:%s", feedID, guid))
	return hex.EncodeToString(h.Sum(nil))
}

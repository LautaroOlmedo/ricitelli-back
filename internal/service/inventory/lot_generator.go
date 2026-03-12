package inventory

import (
	"fmt"
	"time"
)

// GenerateLotNumber generates a lot number in the format L-DDMMYY-NNN-XX
// where NNN is the daily counter (zero-padded to 3 digits) and XX is the
// last 2 characters of the productID for traceability.
// Example: L-120326-001-f3
func GenerateLotNumber(productID string, dailyCounter int) string {
	dateStr := time.Now().UTC().Format("020106") // DDMMYY
	suffix := "00"
	if len(productID) >= 2 {
		suffix = productID[len(productID)-2:]
	}
	return fmt.Sprintf("L-%s-%03d-%s", dateStr, dailyCounter, suffix)
}

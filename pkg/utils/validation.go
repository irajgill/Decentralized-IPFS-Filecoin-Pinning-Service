package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidateCID validates an IPFS CID format
func ValidateCID(cid string) error {
	if len(cid) == 0 {
		return fmt.Errorf("CID cannot be empty")
	}

	// Basic CID validation (starts with Q for v0 or b for v1 base32)
	if !strings.HasPrefix(cid, "Qm") && !strings.HasPrefix(cid, "bafy") && !strings.HasPrefix(cid, "bafk") {
		return fmt.Errorf("invalid CID format")
	}

	// More comprehensive validation could be added here
	return nil
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// ValidateDuration validates storage duration
func ValidateDuration(days int) error {
	if days < 1 {
		return fmt.Errorf("duration must be at least 1 day")
	}
	if days > 1095 { // 3 years max
		return fmt.Errorf("duration cannot exceed 1095 days")
	}
	return nil
}

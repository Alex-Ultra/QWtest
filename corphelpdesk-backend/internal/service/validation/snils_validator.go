package validation

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ValidateSNILS validates a Russian SNILS number according to the official algorithm
// SNILS format: XXX-XXX-XXX XX (where X are digits)
func ValidateSNILS(snils string) error {
	// Remove formatting characters
	snils = regexp.MustCompile(`[-\s]`).ReplaceAllString(snils, "")

	// Check length (should be 11 digits)
	if len(snils) != 11 {
		return errors.New("SNILS must be 11 digits")
	}

	// Check if all characters are digits
	for _, char := range snils {
		if char < '0' || char > '9' {
			return errors.New("SNILS must contain only digits")
		}
	}

	// Extract check digit (last 2 digits) and main number (first 9 digits)
	mainNumber := snils[:9]
	checkDigitStr := snils[9:]

	// Calculate checksum according to official algorithm
	calculatedCheckDigit := calculateChecksum(mainNumber)

	// Parse provided check digit
	providedCheckDigit, err := strconv.Atoi(checkDigitStr)
	if err != nil {
		return fmt.Errorf("invalid check digit: %w", err)
	}

	// Compare calculated and provided check digits
	if calculatedCheckDigit != providedCheckDigit {
		return fmt.Errorf("invalid SNILS check digit: expected %d, got %d", calculatedCheckDigit, providedCheckDigit)
	}

	return nil
}

// calculateChecksum implements the official SNILS checksum calculation algorithm
// Algorithm: multiply each of the first 9 digits by its position (from right to left, starting at 1),
// sum the products, take the remainder of division by 101, and format as two digits
func calculateChecksum(mainNumber string) int {
	sum := 0

	// Calculate weighted sum: each digit multiplied by its position (right to left, 1-indexed)
	for i, char := range mainNumber {
		digit := int(char - '0')
		weight := 9 - i // Position from left to right
		sum += digit * weight
	}

	remainder := sum % 101

	// According to SNILS rules, if remainder is 100 or 101, the check digit becomes 00
	if remainder == 100 || remainder == 101 {
		remainder = 0
	}

	return remainder
}

// FormatSNILS formats a SNILS number to the standard XXX-XXX-XXX XX format
func FormatSNILS(snils string) string {
	// Remove existing formatting
	snils = regexp.MustCompile(`[-\s]`).ReplaceAllString(snils, "")
	
	if len(snils) != 11 {
		return snils // Return as is if invalid length
	}

	// Format as XXX-XXX-XXX XX
	return fmt.Sprintf("%s-%s-%s %s", 
		snils[0:3], 
		snils[3:6], 
		snils[6:9], 
		snils[9:11])
}

// SanitizeSNILS removes formatting characters and returns only digits
func SanitizeSNILS(snils string) string {
	return regexp.MustCompile(`[-\s]`).ReplaceAllString(snils, "")
}
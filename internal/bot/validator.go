package bot

import (
	"errors"
	"strings"
)

// ValidateContact checks if the organizer contact is valid.
func ValidateContact(contact string) error {
	if contact == "" {
		return errors.New("ORGANIZER_CONTACT is required")
	}
	if strings.Contains(contact, "example.com") {
		return errors.New("ORGANIZER_CONTACT must be a real contact (cannot contain 'example.com')")
	}
	return nil
}

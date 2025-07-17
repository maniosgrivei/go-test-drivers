package customer

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ValidateRegisterRequest checks that all required fields in the request are
// valid.
func ValidateRegisterRequest(request *RegisterRequest) error {
	var errs []error

	if err := ValidateName(request.Name); err != nil {
		errs = append(errs, fmt.Errorf("invalid name: '%s': %w", request.Name, err))
	}

	if err := ValidateEmail(request.Email); err != nil {
		errs = append(errs, fmt.Errorf("invalid email: '%s': %w", request.Email, err))
	}

	if err := ValidatePhone(request.Phone); err != nil {
		errs = append(errs, fmt.Errorf("invalid phone: '%s': %w", request.Phone, err))
	}

	return errors.Join(errs...)
}

//
// Name Validation

const (
	minimumNameParts              = 2
	minimumFirstAndLastNameLength = 3
	minimumNameLength             = (minimumNameParts * minimumFirstAndLastNameLength) + 1
	maximumNameLength             = 60

	nameAllowedCharsRegex        = `^[a-zA-Z0-9][a-zA-Z0-9'\-\. ]{5,58}[a-zA-Z0-9\.]$`
	nameInvalidCharSequenceRegex = `(['\-\.]{2,})|([ ]{2,})`
)

// ValidateName validates the provided name against a set of rules.
//
// The name must:
//   - Be between `minimumNameLength` and `maximumNameLength` characters long.
//   - Start and end with an alphanumeric character or a dot.
//   - Not contain consecutive special characters (', -, .') or spaces.
//   - Have at least two parts (first name and last name).
//   - Have a first and last name with at least `minimumFirstAndLastNameLength`
//     characters each.
func ValidateName(name string) error {
	//
	// Name length validation
	if len(name) < minimumNameLength {
		return fmt.Errorf("too short (length < %d)", minimumNameLength)
	}

	if len(name) > maximumNameLength {
		return fmt.Errorf("too long (length > %d)", maximumNameLength)
	}

	//
	// Invalid characters and sequences detection
	allowerChars, _ := regexp.Compile(nameAllowedCharsRegex)
	if !allowerChars.MatchString(name) {
		return fmt.Errorf("invalid characters")
	}

	invalidCharSequence, _ := regexp.Compile(nameInvalidCharSequenceRegex)
	if invalidCharSequence.MatchString(name) {
		return fmt.Errorf("invalid character sequence")
	}

	//
	// Name composition validation
	if err := validateNameComposition(name); err != nil {
		return err
	}

	return nil
}

// validateNameComposition checks that the name has at least `minimumNameParts`
// parts and that the first and last parts have at least
// `minimumFirstAndLastNameLength` characters.
func validateNameComposition(name string) error {
	parts := strings.Split(name, " ")

	if len(parts) < minimumNameParts {
		return fmt.Errorf("not a full name (parts < %d)", minimumNameParts)
	}

	if len(parts[0]) < minimumFirstAndLastNameLength || len(parts[len(parts)-1]) < minimumFirstAndLastNameLength {
		return fmt.Errorf("first or last name too short (length < %d)", minimumFirstAndLastNameLength)
	}

	return nil
}

//
// Email Validation

const (
	minimumEmailUsernameLength  = 3
	minimumEmailServiceLength   = 3
	minimumEmailExtensionLength = 2
	minimumEmailLength          = minimumEmailUsernameLength + minimumEmailServiceLength + minimumEmailExtensionLength + 2
	maximumEmailLength          = 60

	emailInvalidCharSequenceRegex   = `([@_\-\.]{2,})`
	emailUsernameAllowedCharsRegex  = `^[a-zA-Z0-9][a-zA-Z0-9\-\._]{1,51}[a-zA-Z0-9]$`
	emailServiceAllowedCharsRegex   = `^[a-zA-Z0-9][a-zA-Z0-9\-\.]{1,51}[a-zA-Z0-9]$`
	emailExtensionAllowedCharsRegex = `^[a-zA-Z]{2,52}$`
)

// ValidateEmail validates the provided email against a set of rules.
//
// The email must:
//   - Be between `minimumEmailLength` and `maximumEmailLength` characters long.
//   - Not contain consecutive special characters ('@', '_', '-', '.').
//   - Have a valid username, service, and extension.
//   - The username must:
//     -- Be at least `minimumEmailUsernameLength` characters long.
//     -- Start and end with an alphanumeric character.
//     -- Only contain alphanumeric characters, hyphens, underscores, and dots.
//   - The service must:
//     -- Be at least `minimumEmailServiceLength` characters long.
//     -- Start and end with an alphanumeric character.
//     -- Only contain alphanumeric characters and hyphens.
//   - The extension must:
//     -- Be at least `minimumEmailExtensionLength` characters long.
//     -- Only contain alphabetic characters.
func ValidateEmail(email string) error {
	//
	// Email length valication
	if len(email) < minimumEmailLength {
		return fmt.Errorf("too short (length < %d)", minimumEmailLength)
	}

	if len(email) > maximumEmailLength {
		return fmt.Errorf("too long (length > %d)", maximumEmailLength)
	}

	//
	// Email format validation
	invalidCharSequence, _ := regexp.Compile(emailInvalidCharSequenceRegex)
	if invalidCharSequence.MatchString(email) {
		return fmt.Errorf("invalid character sequence")
	}

	username, service, extension, err := decomposeEmail(email)
	if err != nil {
		return err
	}

	if err = validateEmailUsername(username); err != nil {
		return fmt.Errorf("invalid username: %w", err)
	}

	if err = validateEmailService(service); err != nil {
		return fmt.Errorf("invalid service: %w", err)
	}

	if err = validateEmailExtension(extension); err != nil {
		return fmt.Errorf("invalid extension: %w", err)
	}

	return nil
}

// decomposeEmail decomposes an email into its username, service, and extension
// parts.
func decomposeEmail(email string) (username, service, extension string, err error) {
	userAndDomain := strings.Split(email, "@")
	if len(userAndDomain) != 2 {
		return "", "", "", fmt.Errorf("invalid email format")
	}

	username = userAndDomain[0]

	domainParts := strings.Split(userAndDomain[1], ".")
	if len(domainParts) < 2 {
		return "", "", "", fmt.Errorf("invalid email format")
	}

	service = strings.Join(domainParts[:len(domainParts)-1], ".")

	extension = domainParts[len(domainParts)-1]

	return username, service, extension, nil
}

// validateEmailUsername validates the username part of an email.
func validateEmailUsername(username string) error {
	if len(username) < minimumEmailUsernameLength {
		return fmt.Errorf("too short (length < %d)", minimumEmailUsernameLength)
	}

	usernameAllowedChars, _ := regexp.Compile(emailUsernameAllowedCharsRegex)
	if !usernameAllowedChars.MatchString(username) {
		return fmt.Errorf("invalid characters")
	}

	return nil
}

// validateEmailService validates the service part of an email.
func validateEmailService(service string) error {
	if len(service) < minimumEmailServiceLength {
		return fmt.Errorf("too short (length < %d)", minimumEmailServiceLength)
	}

	serviceAllowedChars, _ := regexp.Compile(emailServiceAllowedCharsRegex)
	if !serviceAllowedChars.MatchString(service) {
		return fmt.Errorf("invalid characters")
	}

	return nil
}

// validateEmailExtension validates the extension part of an email.
func validateEmailExtension(extension string) error {
	if len(extension) < minimumEmailExtensionLength {
		return fmt.Errorf("too short (length < %d)", minimumEmailExtensionLength)
	}

	extensionAllowedChars, _ := regexp.Compile(emailExtensionAllowedCharsRegex)
	if !extensionAllowedChars.MatchString(extension) {
		return fmt.Errorf("invalid characters")
	}

	return nil
}

//
// Phone Validation

const (
	minimumPhoneCountryLength = 1
	maximumPhoneCountryLength = 3
	minimumPhoneNumberLength  = 3
	maximumPhoneNumberLength  = 17
	minimumPhoneLength        = minimumPhoneCountryLength + minimumPhoneNumberLength + 2
	maximumPhoneLength        = 20

	phoneCountryAllowedCharsRegex = `^\+[0-9]{1,3}$`
	phoneNumberAllowedCharsRegex  = `^[0-9][0-9 ]{1,15}[0-9]$`
	phoneInvalidCharSequenceRegex = `([ ]{2,})`
)

// ValidatePhone validates the provided phone number against a set of rules.
//
// The phone number must:
//   - Be between `minimumPhoneLength` and `maximumPhoneLength` characters long.
//   - Not contain consecutive spaces.
//   - Have a valid country code and number.
//   - The country code must:
//     -- Start with a '+'.
//     -- Be between `minimumPhoneCountryLength` and `maximumPhoneCountryLength`
//     digits long (excluding the '+').
//     -- Only contain digits after the '+'.
//   - The number must:
//     -- Be between `minimumPhoneNumberLength` and `maximumPhoneNumberLength`
//     digits long (excluding spaces).
//     -- Start and end with a digit.
//     -- Only contain digits and spaces.
func ValidatePhone(phone string) error {
	//
	// Phone length validation
	if len(phone) < minimumPhoneLength {
		return fmt.Errorf("too short (length < %d)", minimumPhoneLength)
	}

	if len(phone) > maximumPhoneLength {
		return fmt.Errorf("too long (length > %d)", maximumPhoneLength)
	}

	//
	// Phone format validation
	invalidCharSequence, _ := regexp.Compile(phoneInvalidCharSequenceRegex)
	if invalidCharSequence.MatchString(phone) {
		return fmt.Errorf("invalid character sequence")
	}

	country, number, err := decomposePhone(phone)
	if err != nil {
		return err
	}

	if err = validatePhoneCountry(country); err != nil {
		return fmt.Errorf("invalid country code: %w", err)
	}

	if err = validatePhoneNumber(number); err != nil {
		return fmt.Errorf("invalid phone number: %w", err)
	}

	return nil
}

// decomposePhone decomposes a phone number into its country code and number
// parts.
func decomposePhone(phone string) (country, number string, err error) {
	parts := strings.SplitN(phone, " ", 2)
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid phone format")
	}

	country = parts[0]
	number = strings.Join(parts[1:], " ")

	return country, number, nil
}

// validatePhoneCountry validates the country code part of a phone number.
func validatePhoneCountry(country string) error {
	if len(country) > 0 && country[0] != '+' {
		return fmt.Errorf("missing the leading plus symbol")
	}

	if len(country) < minimumPhoneCountryLength+1 {
		return fmt.Errorf("too short (length < %d)", minimumPhoneCountryLength)
	}

	if len(country) > maximumPhoneCountryLength+1 {
		return fmt.Errorf("too long (length > %d)", maximumPhoneCountryLength)
	}

	countryAllowedChars, _ := regexp.Compile(phoneCountryAllowedCharsRegex)
	if !countryAllowedChars.MatchString(country) {
		return fmt.Errorf("invalid characters")
	}

	return nil
}

// validatePhoneNumber validates the number part of a phone number.
func validatePhoneNumber(number string) error {
	if len(number) < minimumPhoneNumberLength {
		return fmt.Errorf("too short (length < %d)", minimumPhoneNumberLength)
	}

	if len(number) > maximumPhoneNumberLength {
		return fmt.Errorf("too long (length > %d)", maximumPhoneNumberLength)
	}

	numberAllowedChars, _ := regexp.Compile(phoneNumberAllowedCharsRegex)
	if !numberAllowedChars.MatchString(number) {
		return fmt.Errorf("invalid characters")
	}

	return nil
}

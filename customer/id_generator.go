package customer

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// GenerateID generates customer IDs based on its names and registration time.
//
// The generated IDs have the format `CCCC-YYMD-DNNN`, which means:
//   - CCCC: four characters extracted from the customer name, preferencially
//     consonants but can be completed with vowels.
//   - YY  : the registration year with two digits.
//   - M   : a letter from A to L representing the registration month.
//   - DD  : the registration day with two digits.
//   - NNN : the registration unix milliseconds in Base 36 encoding.
//
// The `name` is expected to have at least six valid characteres.
//
// The `registrationTimestamp` parameter is expected to be after January 1st
// 1970.
func GenerateID(name string, registrationTimestamp time.Time) (string, error) {
	purgedName, err := purgeAndCapsNames(name)
	if err != nil {
		return "", err
	}

	consonants, err := extractConsonants(purgedName)
	if err != nil {
		return "", err
	}

	prefix := completeConsonants(consonants, purgedName)

	date := julianDate(registrationTimestamp)

	millis := unixsMilliBase36(registrationTimestamp)
	if len(millis) > 3 {
		millis = millis[len(millis)-3:]
	}

	return fmt.Sprintf("%s-%s-%s%03s", prefix, date[:4], date[4:], millis), nil
}

// purgeAndCapsNames purges the name from accents, spaces, numbers, and symbols,
// and converts it to uppercase.
func purgeAndCapsNames(name string) (string, error) {
	transformer := transform.Chain(
		norm.NFD,
		runes.Remove(runes.In(unicode.Mn)),
		runes.Map(unicode.ToUpper),
		runes.Remove(runes.NotIn(unicode.Letter)),
		norm.NFC,
	)

	transformedName, _, err := transform.String(transformer, name)
	if err != nil {
		return "", fmt.Errorf("%w: invalid name: %w", ErrValidation, err)
	}

	return transformedName, nil
}

// extractConsonants extracts the consonants from the name. It expects to
// receive a string which contains only upper case ascii characters.
func extractConsonants(src string) (string, error) {
	transformer := runes.Remove(runes.Predicate(
		func(r rune) bool {
			return !strings.ContainsRune("BCDFGHJKLMNPQRSTVWXYZ", r)
		},
	))

	consonants, _, err := transform.String(transformer, src)
	if err != nil {
		return "", fmt.Errorf("%w: invalid name: %w", ErrValidation, err)
	}

	return consonants, nil
}

// completeConsonants returns a four letter string which is the combination of
// the `consonants` and the `purgedName`.
//
// If `consonants` has four or more letters, then it will return the first four
// letters of `consonants`.
//
// If `consonants` has less then four letters, it will complete with the end of
// `purgedName`.
func completeConsonants(consonants, purgedName string) string {
	consLen := len(consonants)

	if consLen >= 4 {
		return consonants[:4]
	}

	nameLen := len(purgedName)

	for consLen > 0 {
		nameIdx := nameLen - (4 - consLen)
		if consonants[consLen-1] != purgedName[nameIdx] {
			break
		}

		consLen--
	}

	fromName := nameLen - (4 - consLen)

	return consonants[:consLen] + purgedName[fromName:]
}

// julianDate generates a Julian date string in the format YYMDD where the month
// is represented by a letter from A to L.
func julianDate(t time.Time) string {
	return fmt.Sprintf("%02d%c%02d", t.Year()%100, rune('A'+t.Month()-1), t.Day())
}

// unixsMilliBase36 generates a base36 string from the Unix milliseconds of the
// given time.
func unixsMilliBase36(t time.Time) string {
	return strings.ToUpper(strconv.FormatInt(t.UnixMilli(), 36))
}

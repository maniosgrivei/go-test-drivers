package customer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGenerateID(t *testing.T) {
	testCases := []struct {
		title    string
		name     string
		regTS    time.Time
		expected string
	}{
		{"john doe jan 2 2006", "John Doe", time.Date(2006, 1, 2, 0, 0, 0, 10000000, time.UTC), "JHND-06A0-2UOA"},
		{"jane doe feb 14 2010", "Jane Doe", time.Date(2010, 2, 14, 0, 0, 0, 1000000, time.UTC), "JNDE-10B1-41C1"},
		{"peter pan dec 25 1999", "Peter Pan", time.Date(1999, 12, 25, 0, 0, 0, 35000000, time.UTC), "PTRP-99L2-5S0Z"},
		{"alice wonderland jul 4 1976", "Alice Wonderland", time.Date(1976, 7, 4, 0, 0, 0, 2000000, time.UTC), "LCWN-76G0-4002"},
		{"bob the builder sep 1 2023", "Bob The Builder", time.Date(2023, 9, 1, 0, 0, 0, 25000000, time.UTC), "BBTH-23I0-15CP"},
		{"david copperfield may 1 1970", "David Copperfield", time.Date(1970, 5, 1, 0, 0, 0, 27000000, time.UTC), "DVDC-70E0-180R"},
		{"eve harrington aug 12 1980", "Eve Harrington", time.Date(1980, 8, 12, 0, 0, 0, 12000000, time.UTC), "VHRR-80H1-2S0C"},
		{"xasptrto norman jan 01 1970", "Xasptrto Norman", time.Date(1970, 1, 1, 0, 0, 0, 1000000, time.UTC), "XSPT-70A0-1001"},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			actual, err := GenerateID(tc.name, tc.regTS)
			require.NoError(t, err)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestPurgeAndCapsNames(t *testing.T) {
	testCases := []struct {
		title    string
		name     string
		expected string
	}{
		{"empty string", "", ""},
		{"simple name", "  john doe  ", "JOHNDOE"},
		{"shuffled alphabet with numbers and symbols", "`!1@2#3$4%5^6&.,7*8(9)0_-+=~`aB`cD`eF`gH`iJ`kL`mN`oP`qR`sT`uV`wX`yZ`", "ABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{"letters with accents", "áéíóúÁÉÍÓÚàèìòùÀÈÌÒÙâêîôûÂÊÎÔÛãõÃÕñÑçÇ", "AEIOUAEIOUAEIOUAEIOUAEIOUAEIOUAOAONNCC"},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			actual, err := purgeAndCapsNames(tc.name)
			require.NoError(t, err)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestExtractAndCapsConsonants(t *testing.T) {
	testCases := []struct {
		title    string
		name     string
		expected string
	}{
		{"empty string", "", ""},
		{"only vowels", "AEIOU", ""},
		{"only consonants", "BCDFG", "BCDFG"},
		{"john due", "JOHNDOE", "JHND"},
		{"peerpoint peperontino", "JHANSENASPEERPOINTPEPPERONTINO", "JHNSNSPRPNTPPPRNTN"},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			actual, err := extractConsonants(tc.name)
			require.NoError(t, err)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestCompleteConsonants(t *testing.T) {
	testCases := []struct {
		title      string
		consonants string
		purgedName string
		expected   string
	}{
		{"john due", "JHND", "JOHNDOE", "JHND"},
		{"jee dee", "JD", "JEEDEE", "JDEE"},
		{"oei uuae", "", "OEIUUAE", "UUAE"},
		{"joe ede", "JD", "JOEEDE", "JEDE"},
		{"jo edde", "JDD", "JOEDDE", "JDDE"},
		{"joe edd", "JDD", "JOEEDD", "JEDD"},
		{"oee jdd", "JDD", "OEEJDD", "EJDD"},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			actual := completeConsonants(tc.consonants, tc.purgedName)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestJulianDate(t *testing.T) {
	testCases := []struct {
		title    string
		date     time.Time
		expected string
	}{
		{"january 2, 2006", time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC), "06A02"},
		{"february 29, 2008 (leap year)", time.Date(2008, 2, 29, 0, 0, 0, 0, time.UTC), "08B29"},
		{"december 31, 2010", time.Date(2010, 12, 31, 0, 0, 0, 0, time.UTC), "10L31"},
		{"september 15, 2023", time.Date(2023, 9, 15, 0, 0, 0, 0, time.UTC), "23I15"},
		{"august 1, 1999", time.Date(1999, 8, 1, 0, 0, 0, 0, time.UTC), "99H01"},
		{"july 4, 1776", time.Date(1776, 7, 4, 0, 0, 0, 0, time.UTC), "76G04"},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			actual := julianDate(tc.date)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestUnixsMilliBase36(t *testing.T) {
	testCases := []struct {
		title    string
		date     time.Time
		expected string
	}{
		{"1 milli", time.Date(1970, 1, 1, 0, 0, 0, 1000000, time.UTC), "1"},
		{"2 milli", time.Date(1970, 1, 1, 0, 0, 0, 2000000, time.UTC), "2"},
		{"3 milli", time.Date(1970, 1, 1, 0, 0, 0, 3000000, time.UTC), "3"},
		{"10 milli", time.Date(1970, 1, 1, 0, 0, 0, 10000000, time.UTC), "A"},
		{"35 milli", time.Date(1970, 1, 1, 0, 0, 0, 35000000, time.UTC), "Z"},
		{"36 milli", time.Date(1970, 1, 1, 0, 0, 0, 36000000, time.UTC), "10"},
		{"march 15 2023 00:00:00", time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC), "LF8X1C00"},
		{"march 15 2023 00:00:00.001", time.Date(2023, 3, 15, 0, 0, 0, 1000000, time.UTC), "LF8X1C01"},
		{"march 15 2023 00:00:00.035", time.Date(2023, 3, 15, 0, 0, 0, 35000000, time.UTC), "LF8X1C0Z"},
		{"march 15 2023 00:00:00.036", time.Date(2023, 3, 15, 0, 0, 0, 36000000, time.UTC), "LF8X1C10"},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			actual := unixsMilliBase36(tc.date)
			require.Equal(t, tc.expected, actual)
		})
	}
}

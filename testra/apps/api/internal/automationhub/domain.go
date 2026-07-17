package automationhub

import (
	"time"

	"github.com/google/uuid"
)

type IngestionFormat string

const (
	FormatJUnit      IngestionFormat = "junit"
	FormatPlaywright IngestionFormat = "playwright"
	FormatCypress    IngestionFormat = "cypress"
)

type IngestResult struct {
	RunID      uuid.UUID
	Total      int
	Passed     int
	Failed     int
	Skipped    int
	DurationMs int64
}

type JUnitTestCase struct {
	Name      string        `xml:"name,attr"`
	Classname string        `xml:"classname,attr"`
	Time      float64       `xml:"time,attr"`
	Status    string        `xml:"status,attr"`
	Failure   *JUnitFailure `xml:"failure"`
}

type JUnitFailure struct {
	Message  string `xml:"message,attr"`
	Type     string `xml:"type,attr"`
	Contents string `xml:",chardata"`
}

type JUnitTestSuite struct {
	Name     string          `xml:"name,attr"`
	Tests    int             `xml:"tests,attr"`
	Failures int             `xml:"failures,attr"`
	Errors   int             `xml:"errors,attr"`
	Skipped  int             `xml:"skipped,attr"`
	Time     float64         `xml:"time,attr"`
	Cases    []JUnitTestCase `xml:"testcase"`
}

type JUnitTestSuites struct {
	Suites []JUnitTestSuite `xml:"testsuite"`
}

type PlaywrightSuite struct {
	Title  string
	Status string
	Tests  []PlaywrightTest
}

type PlaywrightTest struct {
	Title    string
	Status   string
	Duration int64
	Error    string
}

type PlaywrightReport struct {
	Suites []PlaywrightSuite
}

func IsValidFormat(s string) bool {
	switch IngestionFormat(s) {
	case FormatJUnit, FormatPlaywright, FormatCypress:
		return true
	}
	return false
}

func durationFromFloat(seconds float64) int64 {
	return int64(seconds * 1000)
}

func nowUTC() time.Time {
	return time.Now().UTC()
}

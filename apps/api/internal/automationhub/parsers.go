package automationhub

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// ParseReport parses a CI/CD test report into a common representation.
func ParseReport(format IngestionFormat, body []byte) (*ParsedReport, error) {
	switch format {
	case FormatJUnit, FormatPytestJUnit:
		return parseJUnit(body)
	case FormatPlaywright, FormatCypress:
		return parsePlaywrightCypress(body)
	case FormatNewman:
		return parseNewman(body)
	case FormatRobot:
		return parseRobot(body)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// ---------------------------------------------------------------------------
// JUnit / Pytest JUnit
// ---------------------------------------------------------------------------

type jUnitTestSuites struct {
	Suites []jUnitTestSuite `xml:"testsuite"`
}

type jUnitTestSuite struct {
	Name     string          `xml:"name,attr"`
	Tests    int             `xml:"tests,attr"`
	Failures int             `xml:"failures,attr"`
	Errors   int             `xml:"errors,attr"`
	Skipped  int             `xml:"skipped,attr"`
	Time     float64         `xml:"time,attr"`
	Cases    []jUnitTestCase `xml:"testcase"`
}

type jUnitTestCase struct {
	Name      string         `xml:"name,attr"`
	Classname string         `xml:"classname,attr"`
	Time      float64        `xml:"time,attr"`
	Status    string         `xml:"status,attr"`
	Failure   *jUnitFailure  `xml:"failure"`
	Error     *jUnitErrorElt `xml:"error"`
	Skipped   *jUnitSkipped  `xml:"skipped"`
	SystemOut string         `xml:"system-out"`
	SystemErr string         `xml:"system-err"`
}

type jUnitFailure struct {
	Message  string `xml:"message,attr"`
	Type     string `xml:"type,attr"`
	Contents string `xml:",chardata"`
}

type jUnitErrorElt struct {
	Message  string `xml:"message,attr"`
	Type     string `xml:"type,attr"`
	Contents string `xml:",chardata"`
}

type jUnitSkipped struct {
	Message string `xml:"message,attr"`
}

func parseJUnit(body []byte) (*ParsedReport, error) {
	r := bytesReader(body)
	decoder := xml.NewDecoder(r)

	var start xml.StartElement
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return nil, fmt.Errorf("invalid xml root")
		}
		if err != nil {
			return nil, fmt.Errorf("read xml: %w", err)
		}
		var ok bool
		if start, ok = token.(xml.StartElement); ok {
			break
		}
	}

	var suites []jUnitTestSuite
	switch start.Name.Local {
	case "testsuites":
		var root jUnitTestSuites
		if err := decoder.DecodeElement(&root, &start); err != nil {
			return nil, fmt.Errorf("decode testsuites: %w", err)
		}
		suites = root.Suites
	case "testsuite":
		var suite jUnitTestSuite
		if err := decoder.DecodeElement(&suite, &start); err != nil {
			return nil, fmt.Errorf("decode testsuite: %w", err)
		}
		suites = []jUnitTestSuite{suite}
	default:
		return nil, fmt.Errorf("unexpected xml root: %s", start.Name.Local)
	}

	report := &ParsedReport{}
	for _, suite := range suites {
		s := ParsedSuite{Name: nonEmpty(suite.Name, "suite")}
		for _, tc := range suite.Cases {
			c := jUnitCaseToParsed(tc)
			s.Cases = append(s.Cases, c)
			report.DurationMs += c.DurationMs
			report.Total++
			switch c.Status {
			case "passed":
				report.Passed++
			case "failed":
				report.Failed++
			case "skipped":
				report.Skipped++
			case "blocked":
				report.Blocked++
			}
		}
		if len(s.Cases) > 0 {
			report.Suites = append(report.Suites, s)
		}
	}

	// Fall back to suite-level aggregate counts if no cases were parsed.
	if report.Total == 0 {
		for _, suite := range suites {
			report.Total += suite.Tests
			report.Failed += suite.Failures + suite.Errors
			report.Skipped += suite.Skipped
			report.DurationMs += durationFromFloat(suite.Time)
		}
		report.Passed = report.Total - report.Failed - report.Skipped
	}

	return report, nil
}

func jUnitCaseToParsed(tc jUnitTestCase) ParsedCase {
	status := strings.ToLower(tc.Status)
	errMsg := ""
	stack := ""

	switch {
	case tc.Failure != nil:
		status = "failed"
		errMsg = tc.Failure.Message
		if errMsg == "" {
			errMsg = tc.Failure.Type
		}
		stack = strings.TrimSpace(tc.Failure.Contents)
	case tc.Error != nil:
		status = "failed"
		errMsg = tc.Error.Message
		if errMsg == "" {
			errMsg = tc.Error.Type
		}
		stack = strings.TrimSpace(tc.Error.Contents)
	case tc.Skipped != nil:
		status = "skipped"
		errMsg = tc.Skipped.Message
	}

	if status == "" || status == "completed" {
		if tc.Skipped != nil {
			status = "skipped"
		} else {
			status = "passed"
		}
	}

	name := tc.Name
	if name == "" {
		name = tc.Classname
	}

	logs := []string{}
	if out := strings.TrimSpace(tc.SystemOut); out != "" {
		logs = append(logs, out)
	}
	if err := strings.TrimSpace(tc.SystemErr); err != "" {
		logs = append(logs, err)
	}

	return ParsedCase{
		Name:         name,
		Status:       status,
		DurationMs:   durationFromFloat(tc.Time),
		ErrorMessage: errMsg,
		StackTrace:   stack,
		Logs:         logs,
	}
}

// ---------------------------------------------------------------------------
// Playwright / Cypress (simplified and Mochawesome JSON)
// ---------------------------------------------------------------------------

type playwrightSuite struct {
	Title  string            `json:"title"`
	Status string            `json:"status"`
	Tests  []playwrightTest  `json:"tests"`
	Specs  []playwrightSpec  `json:"specs"`
	Suites []playwrightSuite `json:"suites"`
	File   string            `json:"file"`
}

type playwrightSpec struct {
	Title string           `json:"title"`
	Tests []playwrightTest `json:"tests"`
	Ok    bool             `json:"ok"`
}

type playwrightTest struct {
	Title    string `json:"title"`
	Status   string `json:"status"`
	Duration int64  `json:"duration"`
	Error    string `json:"error"`
	State    string `json:"state"`
	Err      *struct {
		Message string `json:"message"`
		Stack   string `json:"stack"`
	} `json:"err"`
	FullTitle string `json:"fullTitle"`
}

type mochawesomeReport struct {
	Stats   mochawesomeStats    `json:"stats"`
	Results []mochawesomeResult `json:"results"`
}

type mochawesomeStats struct {
	Suites  int `json:"suites"`
	Tests   int `json:"tests"`
	Passes  int `json:"passes"`
	Pending int `json:"pending"`
	Failed  int `json:"failed"`
}

type mochawesomeResult struct {
	Title  string             `json:"title"`
	Suites []mochawesomeSuite `json:"suites"`
}

type mochawesomeSuite struct {
	Title  string             `json:"title"`
	Suites []mochawesomeSuite `json:"suites"`
	Tests  []mochawesomeTest  `json:"tests"`
}

type mochawesomeTest struct {
	Title     string `json:"title"`
	FullTitle string `json:"fullTitle"`
	State     string `json:"state"`
	Duration  int64  `json:"duration"`
	Err       *struct {
		Message string `json:"message"`
		Stack   string `json:"stack"`
	} `json:"err"`
}

func parsePlaywrightCypress(body []byte) (*ParsedReport, error) {
	// Try simplified Playwright/Cypress shape first.
	var simple struct {
		Report *playwrightReport `json:"report"`
		Suites []playwrightSuite `json:"suites"`
	}
	if err := json.Unmarshal(body, &simple); err == nil && (len(simple.Suites) > 0 || (simple.Report != nil && len(simple.Report.Suites) > 0)) {
		suites := simple.Suites
		if simple.Report != nil {
			suites = simple.Report.Suites
		}
		return flattenPlaywrightSuites(suites), nil
	}

	// Try Mochawesome (Cypress default JSON reporter).
	var mocha mochawesomeReport
	if err := json.Unmarshal(body, &mocha); err == nil && len(mocha.Results) > 0 {
		return flattenMochawesome(mocha.Results), nil
	}

	return nil, fmt.Errorf("unrecognized playwright/cypress report structure")
}

type playwrightReport struct {
	Suites []playwrightSuite `json:"suites"`
}

func flattenPlaywrightSuites(suites []playwrightSuite) *ParsedReport {
	report := &ParsedReport{}
	for _, suite := range suites {
		flattenPlaywrightSuite(suite, report, "")
	}
	report.Passed = report.Total - report.Failed - report.Skipped
	return report
}

func flattenPlaywrightSuite(suite playwrightSuite, report *ParsedReport, parent string) {
	name := suite.Title
	if parent != "" {
		name = parent + " / " + suite.Title
	}

	s := ParsedSuite{Name: nonEmpty(name, "suite")}
	for _, t := range suite.Tests {
		c := playwrightTestToParsed(t)
		s.Cases = append(s.Cases, c)
		aggregateCase(report, c)
	}
	for _, spec := range suite.Specs {
		for _, t := range spec.Tests {
			c := playwrightTestToParsed(t)
			s.Cases = append(s.Cases, c)
			aggregateCase(report, c)
		}
	}
	if len(s.Cases) > 0 {
		report.Suites = append(report.Suites, s)
	}
	for _, child := range suite.Suites {
		flattenPlaywrightSuite(child, report, name)
	}
}

func playwrightTestToParsed(t playwrightTest) ParsedCase {
	status := strings.ToLower(nonEmpty(t.Status, t.State))
	errMsg := t.Error
	stack := ""
	if t.Err != nil {
		if errMsg == "" {
			errMsg = t.Err.Message
		}
		stack = t.Err.Stack
	}
	if status == "" {
		switch status {
		case "passed", "pass":
			status = "passed"
		case "failed", "fail", "failure":
			status = "failed"
		case "pending", "skipped", "skip":
			status = "skipped"
		default:
			if errMsg != "" || stack != "" {
				status = "failed"
			} else {
				status = "passed"
			}
		}
	}
	if status == "pass" {
		status = "passed"
	}
	if status == "fail" || status == "failure" {
		status = "failed"
	}
	if status == "pending" || status == "skip" {
		status = "skipped"
	}
	return ParsedCase{
		Name:         nonEmpty(t.Title, t.FullTitle, "test"),
		Status:       status,
		DurationMs:   t.Duration,
		ErrorMessage: errMsg,
		StackTrace:   stack,
	}
}

func flattenMochawesome(results []mochawesomeResult) *ParsedReport {
	report := &ParsedReport{}
	for _, result := range results {
		for _, suite := range result.Suites {
			flattenMochaSuite(suite, report, "")
		}
	}
	report.Passed = report.Total - report.Failed - report.Skipped
	return report
}

func flattenMochaSuite(suite mochawesomeSuite, report *ParsedReport, parent string) {
	name := suite.Title
	if parent != "" {
		name = parent + " / " + suite.Title
	}
	s := ParsedSuite{Name: nonEmpty(name, "suite")}
	for _, t := range suite.Tests {
		c := mochaTestToParsed(t)
		s.Cases = append(s.Cases, c)
		aggregateCase(report, c)
	}
	if len(s.Cases) > 0 {
		report.Suites = append(report.Suites, s)
	}
	for _, child := range suite.Suites {
		flattenMochaSuite(child, report, name)
	}
}

func mochaTestToParsed(t mochawesomeTest) ParsedCase {
	status := strings.ToLower(t.State)
	errMsg := ""
	stack := ""
	if t.Err != nil {
		errMsg = t.Err.Message
		stack = t.Err.Stack
	}
	switch status {
	case "passed", "pass":
		status = "passed"
	case "failed", "fail":
		status = "failed"
	case "pending", "skipped", "skip":
		status = "skipped"
	default:
		if errMsg != "" || stack != "" {
			status = "failed"
		} else {
			status = "passed"
		}
	}
	return ParsedCase{
		Name:         nonEmpty(t.FullTitle, t.Title, "test"),
		Status:       status,
		DurationMs:   t.Duration,
		ErrorMessage: errMsg,
		StackTrace:   stack,
	}
}

func aggregateCase(report *ParsedReport, c ParsedCase) {
	report.Total++
	report.DurationMs += c.DurationMs
	switch c.Status {
	case "passed":
		report.Passed++
	case "failed":
		report.Failed++
	case "skipped":
		report.Skipped++
	case "blocked":
		report.Blocked++
	}
}

// ---------------------------------------------------------------------------
// Newman (Postman)
// ---------------------------------------------------------------------------

type newmanReport struct {
	Run newmanRun `json:"run"`
}

type newmanRun struct {
	Stats      newmanStats       `json:"stats"`
	Executions []newmanExecution `json:"executions"`
}

type newmanStats struct {
	Requests struct {
		Total int `json:"total"`
	} `json:"requests"`
	Assertions struct {
		Total int `json:"total"`
	} `json:"assertions"`
}

type newmanExecution struct {
	Item       newmanItem        `json:"item"`
	Response   *newmanResponse   `json:"response"`
	Assertions []newmanAssertion `json:"assertions"`
}

type newmanItem struct {
	Name string `json:"name"`
}

type newmanResponse struct {
	ResponseTime int `json:"responseTime"`
	Code         int `json:"code"`
}

type newmanAssertion struct {
	Assertion string      `json:"assertion"`
	Error     interface{} `json:"error"`
}

func parseNewman(body []byte) (*ParsedReport, error) {
	var rep newmanReport
	if err := json.Unmarshal(body, &rep); err != nil {
		return nil, fmt.Errorf("decode newman report: %w", err)
	}

	report := &ParsedReport{}
	s := ParsedSuite{Name: "Newman collection"}
	for _, exec := range rep.Run.Executions {
		passed := true
		errMsg := ""
		duration := int64(0)
		if exec.Response != nil {
			duration = int64(exec.Response.ResponseTime)
		}
		for _, a := range exec.Assertions {
			if a.Error != nil {
				passed = false
				if errMsg == "" {
					switch v := a.Error.(type) {
					case string:
						errMsg = v
					case map[string]interface{}:
						if msg, ok := v["message"].(string); ok {
							errMsg = msg
						} else {
							errMsg = a.Assertion
						}
					default:
						errMsg = a.Assertion
					}
				}
			}
		}

		status := "passed"
		if !passed {
			status = "failed"
		}
		c := ParsedCase{
			Name:         nonEmpty(exec.Item.Name, "request"),
			Status:       status,
			DurationMs:   duration,
			ErrorMessage: errMsg,
		}
		s.Cases = append(s.Cases, c)
		aggregateCase(report, c)
	}
	if len(s.Cases) > 0 {
		report.Suites = append(report.Suites, s)
	}
	return report, nil
}

// ---------------------------------------------------------------------------
// Robot Framework
// ---------------------------------------------------------------------------

type robotRoot struct {
	XMLName xml.Name     `xml:"robot"`
	Suites  []robotSuite `xml:"suite"`
}

type robotSuite struct {
	Name   string       `xml:"name,attr"`
	Suites []robotSuite `xml:"suite"`
	Tests  []robotTest  `xml:"test"`
}

type robotTest struct {
	Name     string         `xml:"name,attr"`
	Status   robotStatus    `xml:"status"`
	Keywords []robotKeyword `xml:"kw"`
}

type robotKeyword struct {
	Name     string         `xml:"name,attr"`
	Msgs     []robotMsg     `xml:"msg"`
	Keywords []robotKeyword `xml:"kw"`
}

type robotMsg struct {
	Level     string `xml:"level,attr"`
	Timestamp string `xml:"timestamp,attr"`
	Content   string `xml:",chardata"`
}

type robotStatus struct {
	Status    string `xml:"status,attr"`
	StartTime string `xml:"starttime,attr"`
	EndTime   string `xml:"endtime,attr"`
}

func parseRobot(body []byte) (*ParsedReport, error) {
	var root robotRoot
	if err := xml.Unmarshal(body, &root); err != nil {
		return nil, fmt.Errorf("decode robot output: %w", err)
	}

	report := &ParsedReport{}
	for _, suite := range root.Suites {
		flattenRobotSuite(suite, report, "")
	}
	report.Passed = report.Total - report.Failed - report.Skipped
	return report, nil
}

func flattenRobotSuite(suite robotSuite, report *ParsedReport, parent string) {
	name := suite.Name
	if parent != "" {
		name = parent + " / " + suite.Name
	}
	s := ParsedSuite{Name: nonEmpty(name, "suite")}
	for _, t := range suite.Tests {
		c := robotTestToParsed(t)
		s.Cases = append(s.Cases, c)
		aggregateCase(report, c)
	}
	if len(s.Cases) > 0 {
		report.Suites = append(report.Suites, s)
	}
	for _, child := range suite.Suites {
		flattenRobotSuite(child, report, name)
	}
}

func robotTestToParsed(t robotTest) ParsedCase {
	status := strings.ToLower(t.Status.Status)
	duration := robotDuration(t.Status.StartTime, t.Status.EndTime)
	errMsg := ""
	logs := []string{}
	collectRobotLogs(t.Keywords, &logs, &errMsg)

	switch status {
	case "pass", "passed":
		status = "passed"
	case "fail", "failed":
		status = "failed"
	case "skip", "skipped":
		status = "skipped"
	default:
		if errMsg != "" {
			status = "failed"
		} else {
			status = "passed"
		}
	}

	return ParsedCase{
		Name:         t.Name,
		Status:       status,
		DurationMs:   duration,
		ErrorMessage: errMsg,
		Logs:         logs,
	}
}

func collectRobotLogs(kws []robotKeyword, logs *[]string, errMsg *string) {
	for _, kw := range kws {
		for _, m := range kw.Msgs {
			level := strings.ToUpper(m.Level)
			if level == "FAIL" || level == "ERROR" {
				if *errMsg == "" {
					*errMsg = m.Content
				}
			}
			if m.Content != "" {
				*logs = append(*logs, fmt.Sprintf("[%s] %s", m.Level, m.Content))
			}
		}
		collectRobotLogs(kw.Keywords, logs, errMsg)
	}
}

func robotDuration(start, end string) int64 {
	const layout = "20060102 15:04:05.000"
	if start == "" || end == "" {
		return 0
	}
	s, err := time.Parse(layout, start)
	if err != nil {
		return 0
	}
	e, err := time.Parse(layout, end)
	if err != nil {
		return 0
	}
	return e.Sub(s).Milliseconds()
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func nonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func bytesReader(body []byte) io.Reader {
	return strings.NewReader(string(body))
}

func parseInt(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

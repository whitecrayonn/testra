package testgen

import (
	"strings"
	"testing"
)

// testSpec exercises every rule GenerateDraftCases implements: required
// fields, enum, type, string/numeric boundaries, and auth. Mirrors the
// Python prototype's test spec so both implementations are checked against
// the same scenarios.
func testSpec() map[string]interface{} {
	return map[string]interface{}{
		"openapi": "3.0.3",
		"info":    map[string]interface{}{"title": "Test", "version": "1.0"},
		"paths": map[string]interface{}{
			"/widgets": map[string]interface{}{
				"post": map[string]interface{}{
					"security": []interface{}{map[string]interface{}{"bearerAuth": []interface{}{}}},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":     "object",
									"required": []interface{}{"name", "priority"},
									"properties": map[string]interface{}{
										"name":     map[string]interface{}{"type": "string", "minLength": 2.0, "maxLength": 10.0},
										"priority": map[string]interface{}{"type": "string", "enum": []interface{}{"low", "high"}},
										"count":    map[string]interface{}{"type": "integer", "minimum": 1.0, "maximum": 5.0},
										"notes":    map[string]interface{}{"type": "string"},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{"201": map[string]interface{}{"description": "Created"}},
				},
			},
			"/widgets/{id}": map[string]interface{}{
				"get": map[string]interface{}{
					"parameters": []interface{}{
						map[string]interface{}{"name": "id", "in": "path", "required": true, "schema": map[string]interface{}{"type": "string"}},
					},
					"responses": map[string]interface{}{"200": map[string]interface{}{"description": "OK"}},
				},
			},
		},
	}
}

func titles(cases []draftCase) []string {
	out := make([]string, len(cases))
	for i, c := range cases {
		out[i] = c.Title
	}
	return out
}

func containsSubstr(list []string, sub string) bool {
	for _, s := range list {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func TestGeneratesOneHappyPathCasePerEndpoint(t *testing.T) {
	cases, endpointCount := GenerateDraftCases(testSpec())
	if endpointCount != 2 {
		t.Fatalf("expected 2 endpoints, got %d", endpointCount)
	}
	count := 0
	for _, title := range titles(cases) {
		if strings.Contains(title, "valid request succeeds") {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("expected 2 happy path cases, got %d", count)
	}
}

func TestFlagsMissingRequiredFields(t *testing.T) {
	cases, _ := GenerateDraftCases(testSpec())
	ts := titles(cases)
	if !containsSubstr(ts, "missing required 'name'") {
		t.Error("expected a missing required 'name' case")
	}
	if !containsSubstr(ts, "missing required 'priority'") {
		t.Error("expected a missing required 'priority' case")
	}
	if containsSubstr(ts, "missing required 'notes'") {
		t.Error("'notes' is not required and should not get a missing-required case")
	}
}

func TestFlagsPathParamRequired(t *testing.T) {
	cases, _ := GenerateDraftCases(testSpec())
	if !containsSubstr(titles(cases), "missing required 'id'") {
		t.Error("expected a missing required 'id' case for the path parameter")
	}
}

func TestEnumViolationCaseGenerated(t *testing.T) {
	cases, _ := GenerateDraftCases(testSpec())
	if !containsSubstr(titles(cases), "'priority' with value outside its enum") {
		t.Error("expected an enum violation case for 'priority'")
	}
}

func TestStringLengthBoundaries(t *testing.T) {
	cases, _ := GenerateDraftCases(testSpec())
	ts := titles(cases)
	if !containsSubstr(ts, "'name' below min length boundary") {
		t.Error("expected a below-min-length case for 'name'")
	}
	if !containsSubstr(ts, "'name' above max length boundary") {
		t.Error("expected an above-max-length case for 'name'")
	}
}

func TestNumericBoundaries(t *testing.T) {
	cases, _ := GenerateDraftCases(testSpec())
	ts := titles(cases)
	if !containsSubstr(ts, "'count' below minimum boundary") {
		t.Error("expected a below-minimum case for 'count'")
	}
	if !containsSubstr(ts, "'count' above maximum boundary") {
		t.Error("expected an above-maximum case for 'count'")
	}
}

func TestNoBoundaryCaseForUnconstrainedField(t *testing.T) {
	cases, _ := GenerateDraftCases(testSpec())
	ts := titles(cases)
	if containsSubstr(ts, "'notes' below") || containsSubstr(ts, "'notes' above") {
		t.Error("'notes' has no min/max/enum and should not get a boundary case")
	}
}

func TestAuthCasesOnlyOnSecuredEndpoints(t *testing.T) {
	cases, _ := GenerateDraftCases(testSpec())
	ts := titles(cases)
	if !containsSubstr(ts, "POST /widgets — missing auth token") {
		t.Error("expected a missing-auth case for the secured POST /widgets endpoint")
	}
	if !containsSubstr(ts, "POST /widgets — invalid auth token") {
		t.Error("expected an invalid-auth case for the secured POST /widgets endpoint")
	}
	if containsSubstr(ts, "GET /widgets/{id} — missing auth token") {
		t.Error("GET /widgets/{id} has no security block and should not get an auth case")
	}
}

func TestEveryCaseHasNonEmptyExpectedResultAndValidPriority(t *testing.T) {
	cases, _ := GenerateDraftCases(testSpec())
	validPriorities := map[string]bool{"low": true, "medium": true, "high": true, "critical": true}
	for _, c := range cases {
		if strings.TrimSpace(c.ExpectedResult) == "" {
			t.Errorf("case %q has an empty expected result", c.Title)
		}
		if strings.TrimSpace(c.Action) == "" {
			t.Errorf("case %q has an empty action", c.Title)
		}
		if !validPriorities[c.Priority] {
			t.Errorf("case %q has invalid priority %q", c.Title, c.Priority)
		}
	}
}

func TestDeterministicSameSpecSameCaseCount(t *testing.T) {
	spec := testSpec()
	a, _ := GenerateDraftCases(spec)
	b, _ := GenerateDraftCases(spec)
	if len(a) != len(b) {
		t.Fatalf("expected the same case count on repeated runs, got %d and %d", len(a), len(b))
	}
	at, bt := titles(a), titles(b)
	for i := range at {
		if at[i] != bt[i] {
			t.Fatalf("expected identical, stably-ordered output; diverged at index %d: %q vs %q", i, at[i], bt[i])
		}
	}
}

func TestNoPathsReturnsNoCases(t *testing.T) {
	cases, endpointCount := GenerateDraftCases(map[string]interface{}{"openapi": "3.0.3"})
	if len(cases) != 0 || endpointCount != 0 {
		t.Fatalf("expected no cases for a spec with no paths, got %d cases / %d endpoints", len(cases), endpointCount)
	}
}

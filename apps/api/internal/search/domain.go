package search

import "github.com/google/uuid"

type Result struct {
	WorkspaceID    uuid.UUID    `json:"workspace_id"`
	Query          string       `json:"query"`
	Results        []SearchItem `json:"results"`
	Projects       []SearchItem `json:"projects"`
	TestCases      []SearchItem `json:"test_cases"`
	Defects        []SearchItem `json:"defects"`
	Automation     []SearchItem `json:"automation"`
	APICollections []SearchItem `json:"api_collections"`
	TestPlans      []SearchItem `json:"test_plans"`
	Users          []SearchItem `json:"users"`
}

type SearchItem struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	URL      string `json:"url"`
}

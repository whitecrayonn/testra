package integrationhub

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	sharederrors "github.com/testra/testra/apps/api/internal/shared/errors"
	"github.com/testra/testra/apps/api/internal/shared/security"
)

var providerHTTPClient = &http.Client{Timeout: 10 * time.Second}

// --- Jira ---

type jiraProvider struct{}

func (p *jiraProvider) Type() IntegrationType { return TypeJira }

func (p *jiraProvider) Validate(cfg map[string]string) error {
	for _, k := range []string{"url", "token", "project_key"} {
		if cfg[k] == "" {
			return fmt.Errorf("%w: jira config requires '%s'", sharederrors.ErrInvalidInput, k)
		}
	}
	return nil
}

func (p *jiraProvider) Test(ctx context.Context, i *Integration) (string, error) {
	return p.Send(ctx, i, map[string]interface{}{"summary": "Testra connection test", "description": "This is a test issue from Testra."})
}

func (p *jiraProvider) Health(ctx context.Context, i *Integration) (string, error) {
	_, err := p.Test(ctx, i)
	if err != nil {
		return "error", err
	}
	return "healthy", nil
}

func (p *jiraProvider) Send(ctx context.Context, i *Integration, payload map[string]interface{}) (string, error) {
	if err := p.Validate(i.Config); err != nil {
		return "", err
	}
	baseURL := strings.TrimRight(i.Config["url"], "/")
	if err := security.ValidateURL(ctx, baseURL); err != nil {
		return "", sharederrors.ErrInvalidInput
	}

	summary, _ := payload["summary"].(string)
	if summary == "" {
		summary = "Testra event"
	}
	description, _ := payload["description"].(string)

	body, _ := json.Marshal(map[string]interface{}{
		"fields": map[string]interface{}{
			"project":     map[string]string{"key": i.Config["project_key"]},
			"summary":     summary,
			"description": description,
			"issuetype":   map[string]string{"name": "Bug"},
		},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/rest/api/3/issue", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+basicAuth(i.Config["username"], i.Config["token"]))

	resp, err := providerHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("jira returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Key string `json:"key"`
	}
	_ = json.Unmarshal(respBody, &result)
	return result.Key, nil
}

// --- GitHub ---

type githubProvider struct{}

func (p *githubProvider) Type() IntegrationType { return TypeGitHub }

func (p *githubProvider) VerifyWebhook(i *Integration, body []byte, signature string) error {
	if i == nil {
		return fmt.Errorf("missing integration")
	}
	secret := i.Config["webhook_secret"]
	if secret == "" {
		secret = i.Config["secret"]
	}
	return verifyHMACSignature(body, signature, secret)
}

func (p *githubProvider) Validate(cfg map[string]string) error {
	for _, k := range []string{"url", "token", "owner", "repo"} {
		if cfg[k] == "" {
			return fmt.Errorf("%w: github config requires '%s'", sharederrors.ErrInvalidInput, k)
		}
	}
	return nil
}

func (p *githubProvider) Test(ctx context.Context, i *Integration) (string, error) {
	if err := p.Validate(i.Config); err != nil {
		return "", err
	}
	baseURL := strings.TrimRight(i.Config["url"], "/")
	if err := security.ValidateURL(ctx, baseURL); err != nil {
		return "", sharederrors.ErrInvalidInput
	}
	u := fmt.Sprintf("%s/repos/%s/%s", baseURL, i.Config["owner"], i.Config["repo"])
	return p.get(ctx, u, i.Config["token"])
}

func (p *githubProvider) Health(ctx context.Context, i *Integration) (string, error) {
	_, err := p.Test(ctx, i)
	if err != nil {
		return "error", err
	}
	return "healthy", nil
}

func (p *githubProvider) Send(ctx context.Context, i *Integration, payload map[string]interface{}) (string, error) {
	if err := p.Validate(i.Config); err != nil {
		return "", err
	}
	baseURL := strings.TrimRight(i.Config["url"], "/")
	if err := security.ValidateURL(ctx, baseURL); err != nil {
		return "", sharederrors.ErrInvalidInput
	}

	title, _ := payload["title"].(string)
	if title == "" {
		title = "Testra event"
	}
	bodyText, _ := payload["body"].(string)

	body, _ := json.Marshal(map[string]interface{}{
		"title": title,
		"body":  bodyText,
	})
	u := fmt.Sprintf("%s/repos/%s/%s/issues", baseURL, i.Config["owner"], i.Config["repo"])
	return p.post(ctx, u, i.Config["token"], body)
}

func (p *githubProvider) get(ctx context.Context, url, token string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := providerHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("github returned %d: %s", resp.StatusCode, string(respBody))
	}
	return string(respBody), nil
}

func (p *githubProvider) post(ctx context.Context, url, token string, body []byte) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := providerHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("github returned %d: %s", resp.StatusCode, string(respBody))
	}
	var result struct {
		Number int `json:"number"`
	}
	_ = json.Unmarshal(respBody, &result)
	return fmt.Sprintf("#%d", result.Number), nil
}

// --- GitLab ---

type gitlabProvider struct{}

func (p *gitlabProvider) Type() IntegrationType { return TypeGitLab }

func (p *gitlabProvider) Validate(cfg map[string]string) error {
	for _, k := range []string{"url", "token", "project_id"} {
		if cfg[k] == "" {
			return fmt.Errorf("%w: gitlab config requires '%s'", sharederrors.ErrInvalidInput, k)
		}
	}
	return nil
}

func (p *gitlabProvider) Test(ctx context.Context, i *Integration) (string, error) {
	if err := p.Validate(i.Config); err != nil {
		return "", err
	}
	baseURL := strings.TrimRight(i.Config["url"], "/")
	if err := security.ValidateURL(ctx, baseURL); err != nil {
		return "", sharederrors.ErrInvalidInput
	}
	u := fmt.Sprintf("%s/api/v4/projects/%s", baseURL, i.Config["project_id"])
	return p.get(ctx, u, i.Config["token"])
}

func (p *gitlabProvider) Health(ctx context.Context, i *Integration) (string, error) {
	_, err := p.Test(ctx, i)
	if err != nil {
		return "error", err
	}
	return "healthy", nil
}

func (p *gitlabProvider) Send(ctx context.Context, i *Integration, payload map[string]interface{}) (string, error) {
	if err := p.Validate(i.Config); err != nil {
		return "", err
	}
	baseURL := strings.TrimRight(i.Config["url"], "/")
	if err := security.ValidateURL(ctx, baseURL); err != nil {
		return "", sharederrors.ErrInvalidInput
	}

	title, _ := payload["title"].(string)
	if title == "" {
		title = "Testra event"
	}
	description, _ := payload["description"].(string)

	body, _ := json.Marshal(map[string]interface{}{
		"title":       title,
		"description": description,
	})
	u := fmt.Sprintf("%s/api/v4/projects/%s/issues", baseURL, i.Config["project_id"])
	return p.post(ctx, u, i.Config["token"], body)
}

func (p *gitlabProvider) get(ctx context.Context, url, token string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("PRIVATE-TOKEN", token)
	resp, err := providerHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("gitlab returned %d: %s", resp.StatusCode, string(respBody))
	}
	return string(respBody), nil
}

func (p *gitlabProvider) post(ctx context.Context, url, token string, body []byte) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("PRIVATE-TOKEN", token)
	resp, err := providerHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("gitlab returned %d: %s", resp.StatusCode, string(respBody))
	}
	var result struct {
		IID int `json:"iid"`
	}
	_ = json.Unmarshal(respBody, &result)
	return fmt.Sprintf("!%d", result.IID), nil
}

// --- Bitbucket ---

type bitbucketProvider struct{}

func (p *bitbucketProvider) Type() IntegrationType { return TypeBitbucket }

func (p *bitbucketProvider) Validate(cfg map[string]string) error {
	for _, k := range []string{"url", "workspace", "repo_slug", "username", "app_password"} {
		if cfg[k] == "" {
			return fmt.Errorf("%w: bitbucket config requires '%s'", sharederrors.ErrInvalidInput, k)
		}
	}
	return nil
}

func (p *bitbucketProvider) Test(ctx context.Context, i *Integration) (string, error) {
	if err := p.Validate(i.Config); err != nil {
		return "", err
	}
	baseURL := strings.TrimRight(i.Config["url"], "/")
	if err := security.ValidateURL(ctx, baseURL); err != nil {
		return "", sharederrors.ErrInvalidInput
	}
	u := fmt.Sprintf("%s/2.0/repositories/%s/%s", baseURL, i.Config["workspace"], i.Config["repo_slug"])
	return p.get(ctx, u, i.Config["username"], i.Config["app_password"])
}

func (p *bitbucketProvider) Health(ctx context.Context, i *Integration) (string, error) {
	_, err := p.Test(ctx, i)
	if err != nil {
		return "error", err
	}
	return "healthy", nil
}

func (p *bitbucketProvider) Send(ctx context.Context, i *Integration, payload map[string]interface{}) (string, error) {
	if err := p.Validate(i.Config); err != nil {
		return "", err
	}
	baseURL := strings.TrimRight(i.Config["url"], "/")
	if err := security.ValidateURL(ctx, baseURL); err != nil {
		return "", sharederrors.ErrInvalidInput
	}

	title, _ := payload["title"].(string)
	if title == "" {
		title = "Testra event"
	}
	bodyText, _ := payload["body"].(string)

	body, _ := json.Marshal(map[string]interface{}{
		"title": title,
		"content": map[string]interface{}{
			"raw": bodyText,
		},
	})
	u := fmt.Sprintf("%s/2.0/repositories/%s/%s/issues", baseURL, i.Config["workspace"], i.Config["repo_slug"])
	return p.post(ctx, u, i.Config["username"], i.Config["app_password"], body)
}

func (p *bitbucketProvider) get(ctx context.Context, url, username, password string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Basic "+basicAuth(username, password))
	resp, err := providerHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("bitbucket returned %d: %s", resp.StatusCode, string(respBody))
	}
	return string(respBody), nil
}

func (p *bitbucketProvider) post(ctx context.Context, url, username, password string, body []byte) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+basicAuth(username, password))
	resp, err := providerHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("bitbucket returned %d: %s", resp.StatusCode, string(respBody))
	}
	var result struct {
		ID int `json:"id"`
	}
	_ = json.Unmarshal(respBody, &result)
	return fmt.Sprintf("#%d", result.ID), nil
}

// --- Azure DevOps ---

type azureDevOpsProvider struct{}

func (p *azureDevOpsProvider) Type() IntegrationType { return TypeAzureDevOps }

func (p *azureDevOpsProvider) Validate(cfg map[string]string) error {
	for _, k := range []string{"url", "organization", "project", "token"} {
		if cfg[k] == "" {
			return fmt.Errorf("%w: azure_devops config requires '%s'", sharederrors.ErrInvalidInput, k)
		}
	}
	return nil
}

func (p *azureDevOpsProvider) Test(ctx context.Context, i *Integration) (string, error) {
	if err := p.Validate(i.Config); err != nil {
		return "", err
	}
	baseURL := strings.TrimRight(i.Config["url"], "/")
	if err := security.ValidateURL(ctx, baseURL); err != nil {
		return "", sharederrors.ErrInvalidInput
	}
	u := fmt.Sprintf("%s/%s/_apis/projects/%s?api-version=7.0", baseURL, i.Config["organization"], i.Config["project"])
	return p.get(ctx, u, i.Config["token"])
}

func (p *azureDevOpsProvider) Health(ctx context.Context, i *Integration) (string, error) {
	_, err := p.Test(ctx, i)
	if err != nil {
		return "error", err
	}
	return "healthy", nil
}

func (p *azureDevOpsProvider) Send(ctx context.Context, i *Integration, payload map[string]interface{}) (string, error) {
	if err := p.Validate(i.Config); err != nil {
		return "", err
	}
	baseURL := strings.TrimRight(i.Config["url"], "/")
	if err := security.ValidateURL(ctx, baseURL); err != nil {
		return "", sharederrors.ErrInvalidInput
	}

	title, _ := payload["title"].(string)
	if title == "" {
		title = "Testra event"
	}
	description, _ := payload["description"].(string)

	body, _ := json.Marshal([]map[string]interface{}{
		{"op": "add", "path": "/fields/System.Title", "value": title},
		{"op": "add", "path": "/fields/System.Description", "value": description},
	})
	u := fmt.Sprintf("%s/%s/%s/_apis/wit/workitems/$Bug?api-version=7.0", baseURL, i.Config["organization"], i.Config["project"])
	return p.patch(ctx, u, i.Config["token"], body)
}

func (p *azureDevOpsProvider) get(ctx context.Context, url, token string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Basic "+basicAuth("", token))
	resp, err := providerHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("azure_devops returned %d: %s", resp.StatusCode, string(respBody))
	}
	return string(respBody), nil
}

func (p *azureDevOpsProvider) patch(ctx context.Context, url, token string, body []byte) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json-patch+json")
	req.Header.Set("Authorization", "Basic "+basicAuth("", token))
	resp, err := providerHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("azure_devops returned %d: %s", resp.StatusCode, string(respBody))
	}
	var result struct {
		ID int `json:"id"`
	}
	_ = json.Unmarshal(respBody, &result)
	return fmt.Sprintf("#%d", result.ID), nil
}

// --- Linear ---

type linearProvider struct{}

func (p *linearProvider) Type() IntegrationType { return TypeLinear }

func (p *linearProvider) Validate(cfg map[string]string) error {
	if cfg["token"] == "" {
		return fmt.Errorf("%w: linear config requires 'token'", sharederrors.ErrInvalidInput)
	}
	return nil
}

func (p *linearProvider) Test(ctx context.Context, i *Integration) (string, error) {
	if err := p.Validate(i.Config); err != nil {
		return "", err
	}
	query := `{"query": "query { viewer { id name } }"}`
	return p.postGraphQL(ctx, i.Config["token"], []byte(query))
}

func (p *linearProvider) Health(ctx context.Context, i *Integration) (string, error) {
	_, err := p.Test(ctx, i)
	if err != nil {
		return "error", err
	}
	return "healthy", nil
}

func (p *linearProvider) Send(ctx context.Context, i *Integration, payload map[string]interface{}) (string, error) {
	if err := p.Validate(i.Config); err != nil {
		return "", err
	}
	title, _ := payload["title"].(string)
	if title == "" {
		title = "Testra event"
	}
	description, _ := payload["description"].(string)

	teamID := i.Config["team_id"]
	var query string
	if teamID != "" {
		query = fmt.Sprintf(`{"query": "mutation { issueCreate(input: { title: \"%s\" description: \"%s\" teamId: \"%s\" }) { issue { id identifier } } }"}`, title, description, teamID)
	} else {
		query = fmt.Sprintf(`{"query": "mutation { issueCreate(input: { title: \"%s\" description: \"%s\" }) { issue { id identifier } } }"}`, title, description)
	}
	resp, err := p.postGraphQL(ctx, i.Config["token"], []byte(query))
	if err != nil {
		return "", err
	}
	var result struct {
		Data struct {
			IssueCreate struct {
				Issue struct {
					Identifier string `json:"identifier"`
				} `json:"issue"`
			} `json:"issueCreate"`
		} `json:"data"`
	}
	_ = json.Unmarshal([]byte(resp), &result)
	return result.Data.IssueCreate.Issue.Identifier, nil
}

func (p *linearProvider) postGraphQL(ctx context.Context, token string, body []byte) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.linear.app/graphql", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)
	resp, err := providerHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("linear returned %d: %s", resp.StatusCode, string(respBody))
	}
	return string(respBody), nil
}

// --- Slack ---

type slackProvider struct{}

func (p *slackProvider) Type() IntegrationType { return TypeSlack }

func (p *slackProvider) Validate(cfg map[string]string) error {
	if cfg["url"] == "" {
		return fmt.Errorf("%w: slack config requires 'url'", sharederrors.ErrInvalidInput)
	}
	return nil
}

func (p *slackProvider) Test(ctx context.Context, i *Integration) (string, error) {
	return p.Send(ctx, i, map[string]interface{}{"text": "Testra connection test"})
}

func (p *slackProvider) Health(ctx context.Context, i *Integration) (string, error) {
	_, err := p.Test(ctx, i)
	if err != nil {
		return "error", err
	}
	return "healthy", nil
}

func (p *slackProvider) Send(ctx context.Context, i *Integration, payload map[string]interface{}) (string, error) {
	if err := p.Validate(i.Config); err != nil {
		return "", err
	}
	if err := security.ValidateURL(ctx, i.Config["url"]); err != nil {
		return "", sharederrors.ErrInvalidInput
	}
	text, _ := payload["text"].(string)
	if text == "" {
		text = "Testra event"
	}
	body, _ := json.Marshal(map[string]string{"text": text})
	return postJSON(ctx, i.Config["url"], body, nil)
}

// --- Discord ---

type discordProvider struct{}

func (p *discordProvider) Type() IntegrationType { return TypeDiscord }

func (p *discordProvider) Validate(cfg map[string]string) error {
	if cfg["url"] == "" {
		return fmt.Errorf("%w: discord config requires 'url'", sharederrors.ErrInvalidInput)
	}
	return nil
}

func (p *discordProvider) Test(ctx context.Context, i *Integration) (string, error) {
	return p.Send(ctx, i, map[string]interface{}{"content": "Testra connection test"})
}

func (p *discordProvider) Health(ctx context.Context, i *Integration) (string, error) {
	_, err := p.Test(ctx, i)
	if err != nil {
		return "error", err
	}
	return "healthy", nil
}

func (p *discordProvider) Send(ctx context.Context, i *Integration, payload map[string]interface{}) (string, error) {
	if err := p.Validate(i.Config); err != nil {
		return "", err
	}
	if err := security.ValidateURL(ctx, i.Config["url"]); err != nil {
		return "", sharederrors.ErrInvalidInput
	}
	content, _ := payload["content"].(string)
	if content == "" {
		content = "Testra event"
	}
	body, _ := json.Marshal(map[string]string{"content": content})
	return postJSON(ctx, i.Config["url"], body, nil)
}

// --- SMTP ---

type smtpProvider struct{}

func (p *smtpProvider) Type() IntegrationType { return TypeSMTP }

func (p *smtpProvider) Validate(cfg map[string]string) error {
	for _, k := range []string{"host", "port", "from", "to"} {
		if cfg[k] == "" {
			return fmt.Errorf("%w: smtp config requires '%s'", sharederrors.ErrInvalidInput, k)
		}
	}
	return nil
}

func (p *smtpProvider) Test(ctx context.Context, i *Integration) (string, error) {
	if err := p.Validate(i.Config); err != nil {
		return "", err
	}
	c, err := smtp.Dial(i.Config["host"] + ":" + i.Config["port"])
	if err != nil {
		return "", err
	}
	_ = c.Close()
	return "ok", nil
}

func (p *smtpProvider) Health(ctx context.Context, i *Integration) (string, error) {
	_, err := p.Test(ctx, i)
	if err != nil {
		return "error", err
	}
	return "healthy", nil
}

func (p *smtpProvider) Send(ctx context.Context, i *Integration, payload map[string]interface{}) (string, error) {
	if err := p.Validate(i.Config); err != nil {
		return "", err
	}
	subject, _ := payload["subject"].(string)
	if subject == "" {
		subject = "Testra notification"
	}
	bodyText, _ := payload["body"].(string)
	if bodyText == "" {
		bodyText = "Testra event"
	}
	to := i.Config["to"]
	from := i.Config["from"]
	msg := []byte("From: " + from + "\r\nTo: " + to + "\r\nSubject: " + subject + "\r\n\r\n" + bodyText + "\r\n")
	recipients := strings.Split(to, ",")

	var auth smtp.Auth
	if i.Config["username"] != "" && i.Config["password"] != "" {
		auth = smtp.PlainAuth("", i.Config["username"], i.Config["password"], i.Config["host"])
	}

	addr := i.Config["host"] + ":" + i.Config["port"]
	err := smtp.SendMail(addr, auth, from, recipients, msg)
	if err != nil {
		return "", err
	}
	return "sent", nil
}

// --- Generic Webhook ---

type webhookProvider struct{}

func (p *webhookProvider) Type() IntegrationType { return TypeWebhook }

func (p *webhookProvider) VerifyWebhook(i *Integration, body []byte, signature string) error {
	if i == nil {
		return fmt.Errorf("missing integration")
	}
	return verifyHMACSignature(body, signature, i.Config["secret"])
}

func (p *webhookProvider) Validate(cfg map[string]string) error {
	if cfg["url"] == "" {
		return fmt.Errorf("%w: webhook config requires 'url'", sharederrors.ErrInvalidInput)
	}
	return nil
}

func (p *webhookProvider) Test(ctx context.Context, i *Integration) (string, error) {
	return p.Send(ctx, i, map[string]interface{}{"text": "Testra connection test"})
}

func (p *webhookProvider) Health(ctx context.Context, i *Integration) (string, error) {
	_, err := p.Test(ctx, i)
	if err != nil {
		return "error", err
	}
	return "healthy", nil
}

func (p *webhookProvider) Send(ctx context.Context, i *Integration, payload map[string]interface{}) (string, error) {
	if err := p.Validate(i.Config); err != nil {
		return "", err
	}
	if err := security.ValidateURL(ctx, i.Config["url"]); err != nil {
		return "", sharederrors.ErrInvalidInput
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	headers := map[string]string{"Content-Type": "application/json"}
	if secret := i.Config["secret"]; secret != "" {
		sig := hmacSHA256(body, secret)
		headers["X-Webhook-Signature"] = "sha256=" + sig
	}
	return postJSON(ctx, i.Config["url"], body, headers)
}

func postJSON(ctx context.Context, url string, body []byte, headers map[string]string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := providerHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("webhook returned %d: %s", resp.StatusCode, string(respBody))
	}
	return string(respBody), nil
}

func verifyHMACSignature(body []byte, signature, secret string) error {
	if secret == "" {
		if signature == "" {
			return nil
		}
		return fmt.Errorf("webhook secret not configured")
	}
	expected := "sha256=" + hmacSHA256(body, secret)
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return fmt.Errorf("invalid webhook signature")
	}
	return nil
}

func hmacSHA256(body []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}

func basicAuth(username, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}

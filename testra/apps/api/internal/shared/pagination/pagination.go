package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

type Params struct {
	Cursor string
	Limit  int
}

type Meta struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

func ParseParams(r *http.Request) Params {
	q := r.URL.Query()
	limit := DefaultLimit
	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			if n > MaxLimit {
				n = MaxLimit
			}
			limit = n
		}
	}
	return Params{
		Cursor: q.Get("cursor"),
		Limit:  limit,
	}
}

func EncodeCursor(id string) (string, error) {
	b, err := json.Marshal(map[string]string{"id": id})
	if err != nil {
		return "", fmt.Errorf("failed to encode cursor: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func DecodeCursor(cursor string) (string, error) {
	b, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return "", fmt.Errorf("invalid cursor: %w", err)
	}
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return "", fmt.Errorf("invalid cursor: %w", err)
	}
	return m["id"], nil
}

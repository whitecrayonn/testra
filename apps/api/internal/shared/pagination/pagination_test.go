package pagination

import (
	"net/http"
	"net/url"
	"testing"
)

func TestParseParamsDefaults(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: ""}}
	p := ParseParams(req)
	if p.Cursor != "" {
		t.Fatalf("cursor: want empty, got %q", p.Cursor)
	}
	if p.Limit != DefaultLimit {
		t.Fatalf("limit: want %d, got %d", DefaultLimit, p.Limit)
	}
}

func TestParseParamsCustomLimit(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: "limit=42&cursor=abc"}}
	p := ParseParams(req)
	if p.Cursor != "abc" {
		t.Fatalf("cursor: want abc, got %q", p.Cursor)
	}
	if p.Limit != 42 {
		t.Fatalf("limit: want 42, got %d", p.Limit)
	}
}

func TestParseParamsClampsOverMax(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: "limit=999"}}
	p := ParseParams(req)
	if p.Limit != MaxLimit {
		t.Fatalf("limit: want %d (max), got %d", MaxLimit, p.Limit)
	}
}

func TestParseParamsIgnoresInvalidLimit(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: "limit=notanumber"}}
	p := ParseParams(req)
	if p.Limit != DefaultLimit {
		t.Fatalf("limit: want default %d, got %d", DefaultLimit, p.Limit)
	}
}

func TestEncodeDecodeCursor(t *testing.T) {
	id := "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"
	enc, err := EncodeCursor(id)
	if err != nil {
		t.Fatalf("EncodeCursor failed: %v", err)
	}
	dec, err := DecodeCursor(enc)
	if err != nil {
		t.Fatalf("DecodeCursor failed: %v", err)
	}
	if dec != id {
		t.Fatalf("DecodeCursor: want %q, got %q", id, dec)
	}
}

func TestDecodeCursorInvalid(t *testing.T) {
	_, err := DecodeCursor("not-base64!!!")
	if err == nil {
		t.Fatalf("DecodeCursor expected error for invalid base64")
	}
}

func TestDecodeCursorMissingID(t *testing.T) {
	enc, _ := EncodeCursor("")
	dec, err := DecodeCursor(enc)
	if err != nil {
		t.Fatalf("DecodeCursor failed: %v", err)
	}
	if dec != "" {
		t.Fatalf("DecodeCursor: want empty, got %q", dec)
	}
}

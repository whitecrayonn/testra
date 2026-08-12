package password

import "testing"

func TestPolicyValidation(t *testing.T) {
	p := DefaultPolicy()

	tests := []struct {
		password string
		wantErr  bool
	}{
		{"ValidPass123!", false},
		{"Another$trong9", false},
		{"short1!", true},          // too short
		{"onlylowercase!", true},   // no uppercase or digit
		{"ONLYUPPERCASE1!", true},  // no lowercase
		{"NoSpecialChar123", true}, // no special character
		{"Password123!", true},     // blocked common password
		{"password123!", true},     // blocked common password (lowercase)
	}

	for _, tc := range tests {
		err := p.Validate(tc.password)
		if tc.wantErr && err == nil {
			t.Fatalf("expected error for %q, got nil", tc.password)
		}
		if !tc.wantErr && err != nil {
			t.Fatalf("unexpected error for %q: %v", tc.password, err)
		}
	}
}

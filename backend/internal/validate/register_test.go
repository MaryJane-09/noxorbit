package validate

import (
	"strings"
	"testing"

	"github.com/MaryJane-09/noxorbit/backend/internal/user"
)

func TestValidateEmailAcceptsOnlyOnePlainAddress(t *testing.T) {
	cases := []struct{
		address string
		wantValid bool
	}{
		{"person@example.com", true},
		{"person+tag@example.com", true},
		{"a@x.com, b@y.com", false},
		{"Name <a@x.com>", false},
		{"", false},
		{"person@example.com ", false},
		{strings.Repeat("a", 245)+"@example.com", false},
	}
	for _, tc := range cases {
		err := ValidateEmail(tc.address)
		if (err == nil) != tc.wantValid {
			t.Errorf("ValidateEmail(%q): got %v, valid=%v", tc.address, err, tc.wantValid)
		}
	}
}

func TestValidateRegisterAllowsRealNamesAndCapsLength(t *testing.T) {
	makeUser := func(name string) user.User {
		return user.User{Name: name, Email: "person@example.com", Password: "VeryStrongPass123"}
	}
	for _, name := range []string{"Mary-Jane", "O'Brien", "O’Brien", "Élodie Martin", "Ayo Adewale"} {
		if err := ValidateRegister(makeUser(name)); err != nil {
			t.Errorf("name %q incorrectly rejected: %v", name, err)
		}
	}
	for _, name := range []string{"", "  ", strings.Repeat("a", 101), "bad<>name", "Name123"} {
		if err := ValidateRegister(makeUser(name)); err == nil {
			t.Errorf("name %q should be rejected", name)
		}
	}
}

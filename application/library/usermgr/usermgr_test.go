package usermgr

import (
	"errors"
	"strings"
	"testing"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{name: "valid", password: "StrongP@ssw0rd!123", wantErr: false},
		{name: "valid with spaces", password: "with space", wantErr: false},
		{name: "valid unicode", password: "密码123", wantErr: false},
		{name: "empty", password: "", wantErr: true},
		{name: "newline injection", password: "mynewpass\nroot:HAXXED_ROOT_PW\n", wantErr: true},
		{name: "embedded newline", password: "a\nb", wantErr: true},
		{name: "carriage return", password: "pass\rfoo", wantErr: true},
		{name: "colon", password: "pass:word", wantErr: true},
		{name: "too long", password: strings.Repeat("a", 4097), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidatePassword(%q) error = %v, wantErr %v", tt.password, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePasswordWrapsErrInvalidInput(t *testing.T) {
	for _, pwd := range []string{"", "bad\npassword", "bad\rpassword", "bad:password"} {
		if err := ValidatePassword(pwd); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("ValidatePassword(%q) = %v, want ErrInvalidInput", pwd, err)
		}
	}
}

package domain

import (
	"errors"
	"strings"
	"testing"
)

const validClientID = "3f2b8c1e-9d4a-4b7e-8f21-0a1b2c3d4e5f"

func TestValidateClientID(t *testing.T) {
	tests := []struct {
		name     string
		clientID string
		wantErr  bool
	}{
		{name: "valid lowercase uuid", clientID: validClientID},
		{name: "uppercase rejected", clientID: strings.ToUpper(validClientID), wantErr: true},
		{name: "not a uuid", clientID: "abc", wantErr: true},
		{name: "empty", clientID: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateClientID(tt.clientID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateClientID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateProfile(t *testing.T) {
	valid := Profile{
		ClientID:    validClientID,
		Name:        "  Alex  ",
		Category:    CategorySport,
		Subcategory: " football ",
	}

	tests := []struct {
		name          string
		mutate        func(*Profile)
		wantErr       error
		wantName      string
		wantIntensity Intensity
	}{
		{name: "trims and defaults intensity", wantName: "Alex", wantIntensity: IntensityMedium},
		{name: "bad client id", mutate: func(p *Profile) { p.ClientID = "x" }, wantErr: ErrInvalidClientID},
		{name: "missing name", mutate: func(p *Profile) { p.Name = " " }, wantErr: ErrInvalidProfile},
		{name: "unknown category", mutate: func(p *Profile) { p.Category = "cooking-show" }, wantErr: ErrInvalidProfile},
		{name: "missing subcategory", mutate: func(p *Profile) { p.Subcategory = "" }, wantErr: ErrInvalidProfile},
		{name: "details too long", mutate: func(p *Profile) { p.Details = strings.Repeat("a", maxDetailsLength+1) }, wantErr: ErrInvalidProfile},
		{name: "unknown intensity", mutate: func(p *Profile) { p.Intensity = "nuclear" }, wantErr: ErrInvalidProfile},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := valid
			if tt.mutate != nil {
				tt.mutate(&input)
			}

			got, err := ValidateProfile(input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateProfile() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if got.Name != tt.wantName || got.Subcategory != "football" || got.Intensity != tt.wantIntensity {
				t.Errorf("ValidateProfile() = %+v", got)
			}
		})
	}
}

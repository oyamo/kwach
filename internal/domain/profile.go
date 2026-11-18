package domain

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

const (
	maxNameLength        = 50
	maxSubcategoryLength = 60
	maxDetailsLength     = 500
)

var (
	ErrInvalidClientID = errors.New("client ID is invalid")
	ErrInvalidProfile  = errors.New("profile is invalid")
	ErrProfileNotFound = errors.New("profile not found")
)

type Category string

const (
	CategorySport      Category = "sport"
	CategoryPolitics   Category = "politics"
	CategoryCareer     Category = "career"
	CategoryMusic      Category = "music"
	CategoryTechnology Category = "technology"
	CategoryFood       Category = "food"
	CategorySchool     Category = "school"
	CategoryLifestyle  Category = "lifestyle"
	CategoryOther      Category = "other"
)

var validCategories = []Category{
	CategorySport, CategoryPolitics, CategoryCareer, CategoryMusic,
	CategoryTechnology, CategoryFood, CategorySchool, CategoryLifestyle, CategoryOther,
}

type Intensity string

const (
	IntensityMild   Intensity = "mild"
	IntensityMedium Intensity = "medium"
	IntensitySavage Intensity = "savage"
)

var validIntensities = []Intensity{IntensityMild, IntensityMedium, IntensitySavage}

type Profile struct {
	ClientID    string
	Name        string
	Category    Category
	Subcategory string
	Details     string
	Intensity   Intensity
}

var clientIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func ValidateClientID(clientID string) error {
	if clientIDPattern.MatchString(clientID) {
		return nil
	}
	return ErrInvalidClientID
}

func ValidateProfile(profile Profile) (Profile, error) {
	if err := ValidateClientID(profile.ClientID); err != nil {
		return Profile{}, err
	}

	profile.Name = strings.TrimSpace(profile.Name)
	profile.Subcategory = strings.TrimSpace(profile.Subcategory)
	profile.Details = strings.TrimSpace(profile.Details)
	if profile.Intensity == "" {
		profile.Intensity = IntensityMedium
	}

	if err := checkProfileFields(profile); err != nil {
		return Profile{}, fmt.Errorf("%w: %v", ErrInvalidProfile, err)
	}
	return profile, nil
}

func checkProfileFields(profile Profile) error {
	switch {
	case profile.Name == "":
		return errors.New("name is required")
	case utf8.RuneCountInString(profile.Name) > maxNameLength:
		return fmt.Errorf("name must be at most %d characters", maxNameLength)
	case !slices.Contains(validCategories, profile.Category):
		return fmt.Errorf("unknown category %q", profile.Category)
	case profile.Subcategory == "":
		return errors.New("subcategory is required")
	case utf8.RuneCountInString(profile.Subcategory) > maxSubcategoryLength:
		return fmt.Errorf("subcategory must be at most %d characters", maxSubcategoryLength)
	case utf8.RuneCountInString(profile.Details) > maxDetailsLength:
		return fmt.Errorf("details must be at most %d characters", maxDetailsLength)
	case !slices.Contains(validIntensities, profile.Intensity):
		return fmt.Errorf("unknown intensity %q", profile.Intensity)
	}
	return nil
}

package models

import "time"

// JobListing represents a job listing
type JobListing struct {
	// Basic information
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Employer       string    `json:"employer"`
	Location       string    `json:"location"`
	EmploymentType string    `json:"employment_type"`

	// Dates
	PublishedDate       time.Time  `json:"published_date"`
	ApplicationDeadline *time.Time `json:"application_deadline,omitempty"`

	// Description
	Summary     string `json:"summary"`
	Description string `json:"description,omitempty"` // Full description (only in details)

	// Additional details
	Requirements []string `json:"requirements,omitempty"`
	Benefits     []string `json:"benefits,omitempty"`
	SalaryInfo   string   `json:"salary_info,omitempty"`

	// Link
	URL string `json:"url"`
}

// SearchParams contains parameters for job search
type SearchParams struct {
	// Search & Basic Filters
	Query          string `json:"query,omitempty"`
	Location       string `json:"location,omitempty"`
	OccupationGroup string `json:"occupation_group,omitempty"`

	// Employment Filters
	EmployerType string `json:"employer_type,omitempty"` // company, public, nonprofit
	WorkingHours string `json:"working_hours,omitempty"` // full-time, part-time
	Duration     string `json:"duration,omitempty"`      // permanent, temporary

	// Date & Language
	PublishedAfter string `json:"published_after,omitempty"` // ISO 8601
	Language       string `json:"language,omitempty"`         // fi, sv, en

	// Pagination
	Page     int `json:"page,omitempty"`      // 0-based
	PageSize int `json:"page_size,omitempty"` // 100-500
}

// FilterOptions contains available filter options
type FilterOptions struct {
	OccupationGroups []OccupationGroup `json:"occupation_groups"`
	EmployerTypes    []EmployerType    `json:"employer_types"`
	WorkingHours     []string          `json:"working_hours"`
	Durations        []string          `json:"durations"`
	Regions          []string          `json:"regions"`
	Languages        []string          `json:"languages"`
}

// OccupationGroup represents an ESCO occupation group
type OccupationGroup struct {
	Code   string `json:"code"`
	NameFI string `json:"name_fi"`
	NameEN string `json:"name_en,omitempty"`
	NameSV string `json:"name_sv,omitempty"`
}

// EmployerType represents an employer type option
type EmployerType struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

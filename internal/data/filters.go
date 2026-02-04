package data

import (
	"greenlight/internal/validator"
	"slices"
	"strings"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

type Metadata struct {
	CurrentPage  int `json:"current_page,omitzero"`
	PageSize     int `json:"page_size,omitzero"`
	FirstPage    int `json:"first_page,omitzero"`
	LastPage     int `json:"last_page,omitzero"`
	TotalRecords int `json:"total_records,omitzero"`
}

func ValidateFilters(v *validator.Validator, filters *Filters) {
	v.Check(filters.Page > 0, "page", "must be greater than zero")
	v.Check(filters.Page <= 10_000_000, "page", "must be less then a million")

	v.Check(filters.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(filters.PageSize <= 100, "page_size", "must be less than 100")

	v.Check(validator.PermittedValue(filters.Sort, filters.SortSafeList...), "sort", "invalid sort value")
}

// extract the collums name form the sort parameter
func (f Filters) sortCollumn() string {
	if slices.Contains(f.SortSafeList, f.Sort) {
		return strings.TrimPrefix(f.Sort, "-")
	}
	panic("unsafe sort parameter" + f.Sort)
}

func (f Filters) sortDirection() string {

	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"

}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

func calculateMetadata(totalRecords, page, pageSize int) Metadata {

	if totalRecords == 0 {
		return Metadata{}
	}
	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     (totalRecords + pageSize - 1) / pageSize,
		TotalRecords: totalRecords,
	}
}

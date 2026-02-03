package data

import (
	"greenlight/internal/validator"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

func ValidateFilters(v *validator.Validator, filters *Filters) {
	v.Check(filters.Page > 0, "page", "must be greater than zero")
	v.Check(filters.Page <= 10_000_000, "page", "must be less then a million")

	v.Check(filters.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(filters.PageSize <= 100, "page_size", "must be less than 100")

	v.Check(validator.PermittedValue(filters.Sort, filters.SortSafeList...), "sort", "invalid sort value")
}

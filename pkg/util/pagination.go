// Package util holds small, dependency-free helpers shared across layers.
package util

const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
)

// PaginationMeta is the metadata block returned alongside any paginated list.
type PaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

// NormalizePage clamps page to a minimum of DefaultPage.
func NormalizePage(page int) int {
	if page < 1 {
		return DefaultPage
	}
	return page
}

// NormalizePageSize clamps pageSize to [1, MaxPageSize], defaulting to
// DefaultPageSize when zero or negative. This is the single place that
// enforces the "clients cannot force page_size=100000" rule.
func NormalizePageSize(pageSize int) int {
	if pageSize <= 0 {
		return DefaultPageSize
	}
	if pageSize > MaxPageSize {
		return MaxPageSize
	}
	return pageSize
}

// NewPaginationMeta builds a PaginationMeta from normalized page/pageSize and
// the total row count returned by the database.
func NewPaginationMeta(page, pageSize int, totalItems int64) *PaginationMeta {
	page = NormalizePage(page)
	pageSize = NormalizePageSize(pageSize)

	totalPages := int(totalItems / int64(pageSize))
	if totalItems%int64(pageSize) != 0 {
		totalPages++
	}

	return &PaginationMeta{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}

package domain

import (
	apperrors "github.com/fauzie/golang-sekeleton/pkg/errors"
	"github.com/fauzie/golang-sekeleton/pkg/util"
)

// PaginationMeta mirrors pkg/util.PaginationMeta so domain payloads don't
// need to import pkg/util directly.
type PaginationMeta = util.PaginationMeta

// PaginationRequest is the input to any GetAllXPaginated use case.
type PaginationRequest struct {
	Page     int
	PageSize int
}

// Validate normalizes Page/PageSize in place (page >= 1, 1 <= page_size <=
// 100). It never rejects a request — out-of-range values are clamped, not
// treated as validation errors — but returns error to satisfy call sites
// that treat all request validation uniformly.
func (r *PaginationRequest) Validate() error {
	if r == nil {
		return apperrors.NewValidationError("pagination request is nil", nil)
	}
	r.Page = util.NormalizePage(r.Page)
	r.PageSize = util.NormalizePageSize(r.PageSize)
	return nil
}

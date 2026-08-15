package usecase

// normalizePagination clamps list-endpoint paging parameters to the platform's
// shared window: default page size 20, max 100, non-negative offset.
func normalizePagination(limit, offset int) (int, int) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

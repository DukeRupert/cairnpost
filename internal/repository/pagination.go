package repository

const (
	DefaultLimit = 25
	MaxLimit     = 100
)

type Pagination struct {
	Limit  int
	Offset int
}

func (p *Pagination) normalize() (limit, offset int) {
	limit = p.Limit
	if limit <= 0 {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
	offset = p.Offset
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

package pagination

type Params struct{ Page, Limit, MaxLimit int }
type Metadata struct{ Page, Limit, Offset, Total, TotalPages int }

func (p Params) Normalize() Params {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 20
	}
	if p.MaxLimit < 1 {
		p.MaxLimit = 100
	}
	if p.Limit > p.MaxLimit {
		p.Limit = p.MaxLimit
	}
	return p
}
func (p Params) Offset() int { p = p.Normalize(); return (p.Page - 1) * p.Limit }
func (p Params) Metadata(total int) Metadata {
	p = p.Normalize()
	pages := 0
	if total > 0 {
		pages = (total + p.Limit - 1) / p.Limit
	}
	return Metadata{p.Page, p.Limit, p.Offset(), total, pages}
}

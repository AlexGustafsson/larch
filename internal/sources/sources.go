package sources

import "context"

type Source interface {
	URLs(context.Context) ([]string, error)
}

type URLSource struct {
	URL string
}

func (u *URLSource) URLs(ctx context.Context) ([]string, error) {
	return []string{u.URL}, nil
}

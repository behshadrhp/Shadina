package storage

import (
	"net/http"
	"strconv"
	"time"
)

type PaginatedQuery struct {
	Limit  int    `json:"limit" validate:"gte=1,lte=100"`
	Offset int    `json:"offset" validate:"gte=0"`
	Sort   string `json:"sort" validate:"oneof=ASC DESC asc desc"`
	Search string `json:"search" validate:"max=100"`
	Since  string `json:"since"`
	Until  string `json:"until"`
}

func CreatePaginatedQuery(limit, offset int, sort string) PaginatedQuery {
	return PaginatedQuery{
		Limit:  limit,
		Offset: offset,
		Sort:   sort,
	}
}

func (pq *PaginatedQuery) Parse(r *http.Request) (PaginatedQuery, error) {
	qs := r.URL.Query()

	limit := qs.Get("limit")
	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return *pq, nil
		}
		pq.Limit = l
	}

	offset := qs.Get("offset")
	if offset != "" {
		f, err := strconv.Atoi(offset)
		if err != nil {
			return *pq, nil
		}
		pq.Offset = f
	}

	sort := qs.Get("sort")
	if sort != "" {
		pq.Sort = sort
	}

	search := qs.Get("search")
	if search != "" {
		pq.Search = search
	}

	since := qs.Get("since")
	if since != "" {
		pq.Since = formatTime(since)
	}

	until := qs.Get("until")
	if until != "" {
		pq.Until = until
	}

	return *pq, nil
}

func formatTime(s string) string {
	t, err := time.Parse(time.DateTime, s)
	if err != nil {
		return ""
	}

	return t.Format(time.DateTime)
}

package entity

import "time"

type (
	DaoQuery struct {
		DataHashes     []string  `query:"data_hash"`
		Deleted        []bool    `query:"deleted"`
		CreatedAtStart time.Time `query:"created_at_start"`
		CreatedAtEnd   time.Time `query:"created_at_end"`
		UpdatedAtStart time.Time `query:"updated_at_start"`
		UpdatedAtEnd   time.Time `query:"updated_at_end"`
	}

	DaoEntity struct {
		DataHash  Hash       `json:"dataHash"`
		Deleted   bool       `json:"deleted"`
		CreatedAt time.Time  `json:"createdAt"`
		UpdatedAt *time.Time `json:"updatedAt"`
	}
)

func (de DaoEntity) MakeDataHash(fields ...string) Hash {
	return MakeHash(fields...)
}

// Paging and Limit
type (
	Page  int
	Limit int
)

func (page Page) Int() int {
	return int(page)
}

func (limit Limit) Int() int {
	return int(limit)
}

type (
	PagingQuery struct {
		Page    Page  `json:"page"`
		Limit   Limit `json:"limit"`
		NoLimit bool  `json:"no_limit"`
	}
)

func (e PagingQuery) GetOffset() int {
	return int(e.Page-1) * int(e.Limit)
}

package entity

import (
	pubEntity "rakit-tiket-be/pkg/entity"
)

type EventStatus string

const (
	EventStatusDraft     EventStatus = "DRAFT"
	EventStatusPublished EventStatus = "PUBLISHED"
	EventStatusCompleted EventStatus = "COMPLETED"
	EventStatusCanceled  EventStatus = "CANCELED"
)

type (
	EventQuery struct {
		IDs      []string      `query:"id"`
		Slugs    []string      `query:"slug"`
		Statuses []EventStatus `query:"status"`
	}

	Event struct {
		ID   pubEntity.UUID `json:"id"`
		Slug string         `json:"slug"`
		Name string         `json:"name"`

		// Status & Configuration
		Status           EventStatus `json:"status"`
		TicketPrefixCode string      `json:"ticket_prefix_code"`
		MaxTicketPerTx   int         `json:"max_ticket_per_tx"`

		pubEntity.DaoEntity
	}

	Events []Event
)

package app_registrant

import (
	"time"

	pubEntity "rakit-tiket-be/pkg/entity"
)

type (
	RegistrantQuery struct {
		IDs         []string `query:"id"`
		EventIDs    []string `query:"event_id"`
		TicketIDs   []string `query:"ticket_id"`
		UniqueCodes []string `query:"unique_code"`
		Emails      []string `query:"email"`
		Statuses    []string `query:"status"`

		// List filters
		Search        string   `query:"search"`
		TicketTypes   []string `query:"ticket_type"`
		PaymentStatus []string `query:"payment_status"`
		DateStart     string   `query:"date_start"`
		DateEnd       string   `query:"date_end"`
		SortBy        string   `query:"sort_by"`
		SortOrder     string   `query:"sort_order"`

		pubEntity.DaoQuery
		pubEntity.PagingQuery
	}

	Registrant struct {
		ID      pubEntity.UUID `json:"id"`
		EventID pubEntity.UUID `json:"event_id"`

		// Registrant Info
		UniqueCode string          `json:"unique_code"`
		TicketID   *pubEntity.UUID `json:"ticket_id"`
		Name       string          `json:"name"`
		Email      string          `json:"email"`
		Phone      string          `json:"phone"`
		Gender     *string         `json:"gender"`
		Birthdate  *time.Time      `json:"birthdate"`

		// Transaction Info
		TotalCost    float64 `json:"total_cost"`
		TotalTickets int     `json:"total_tickets"`
		Status       string  `json:"status"`

		Attendees Attendee `json:"attendee"`

		pubEntity.DaoEntity
	}

	Registrants []Registrant

	DashboardSummary struct {
		TotalTicketsSold  int     `json:"total_tickets_sold"`
		TicketsSoldChange float64 `json:"tickets_sold_change"`
		TotalRegistrants  int     `json:"total_registrants"`
		RegistrantsChange float64 `json:"registrants_change"`
		TotalRevenue      float64 `json:"total_revenue"`
		RevenueChange     float64 `json:"revenue_change"`
		ActiveEvents      int     `json:"active_events"`
	}

	DailySales struct {
		Date        string  `json:"date"`
		TicketsSold int     `json:"tickets_sold"`
		Revenue     float64 `json:"revenue"`
	}

	DailySalesTrend []DailySales

	TicketDistribution struct {
		TicketType    string  `json:"ticket_type"`
		TicketsSold   int     `json:"tickets_sold"`
		TotalCapacity int     `json:"total_capacity"`
		Percentage    float64 `json:"percentage"`
	}

	TicketDistributions []TicketDistribution

	RecentTransaction struct {
		ID         string  `json:"id"`
		BuyerName  string  `json:"buyer_name"`
		TicketType string  `json:"ticket_type"`
		Quantity   int     `json:"quantity"`
		Amount     float64 `json:"amount"`
		Status     string  `json:"status"`
		TimeAgo    string  `json:"time_ago"`
	}

	RecentTransactions []RecentTransaction

	DashboardData struct {
		Summary             DashboardSummary    `json:"summary"`
		DailySalesTrend     DailySalesTrend     `json:"daily_sales_trend"`
		TicketDistributions TicketDistributions `json:"ticket_distribution"`
		RecentTransactions  RecentTransactions  `json:"recent_transactions"`
	}
)

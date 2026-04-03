package dao

import (
	"context"
	"fmt"
	"time"

	pubEntity "rakit-tiket-be/pkg/entity"
	entity "rakit-tiket-be/pkg/entity/app_registrant"
	"rakit-tiket-be/pkg/util"

	"gitlab.com/threetopia/sqlgo/v2"
	"go.uber.org/zap"
)

type RegistrantDAO interface {
	Search(ctx context.Context, query entity.RegistrantQuery) (entity.Registrants, int, error)
	GetDashboardSummary(ctx context.Context) (entity.DashboardSummary, error)
	GetDailySalesTrend(ctx context.Context, days int) (entity.DailySalesTrend, error)
	GetTicketDistribution(ctx context.Context) (entity.TicketDistributions, error)
	GetRecentTransactions(ctx context.Context, limit int) (entity.RecentTransactions, error)
	Insert(ctx context.Context, registrants entity.Registrants) error
	Update(ctx context.Context, registrants entity.Registrants) error
	Delete(ctx context.Context, id pubEntity.UUID) error
	SoftDelete(ctx context.Context, id pubEntity.UUID) error
}

type registrantDAO struct {
	log   util.LogUtil
	dbTrx DBTransaction
}

func MakeRegistrantDAO(log util.LogUtil, dbTrx DBTransaction) RegistrantDAO {
	return registrantDAO{
		log:   log,
		dbTrx: dbTrx,
	}
}

func (d registrantDAO) Search(ctx context.Context, query entity.RegistrantQuery) (entity.Registrants, int, error) {
	sqlSelect := sqlgo.NewSQLGoSelect().
		SetSQLSelect("r.id", "id").
		SetSQLSelect("r.event_id", "event_id").
		SetSQLSelect("r.unique_code", "unique_code").
		SetSQLSelect("r.ticket_id", "ticket_id").
		SetSQLSelect("r.name", "name").
		SetSQLSelect("r.email", "email").
		SetSQLSelect("r.phone", "phone").
		SetSQLSelect("r.gender", "gender").
		SetSQLSelect("r.birthdate", "birthdate").
		SetSQLSelect("r.total_cost", "total_cost").
		SetSQLSelect("r.total_tickets", "total_tickets").
		SetSQLSelect("r.status", "status").
		SetSQLSelect("r.data_hash", "data_hash").
		SetSQLSelect("r.deleted", "deleted").
		SetSQLSelect("r.created_at", "created_at").
		SetSQLSelect("r.updated_at", "updated_at")

	sqlFrom := sqlgo.NewSQLGoFrom().
		SetSQLFrom("registrants", "r")

	sqlJoin := sqlgo.NewSQLGoJoin()
	if query.Search != "" || len(query.PaymentStatus) > 0 {
		sqlJoin.SetSQLJoin("LEFT", "orders", "o", sqlgo.SetSQLJoinWhere("AND", "o.registrant_id", "=", "r.id"))
	}
	if len(query.TicketTypes) > 0 {
		sqlJoin.SetSQLJoin("LEFT", "tickets", "rt", sqlgo.SetSQLJoinWhere("AND", "rt.id", "=", "r.ticket_id"))
		sqlJoin.SetSQLJoin("LEFT", "attendees", "a", sqlgo.SetSQLJoinWhere("AND", "a.registrant_id", "=", "r.id"))
		sqlJoin.SetSQLJoin("LEFT", "tickets", "at", sqlgo.SetSQLJoinWhere("AND", "at.id", "=", "a.ticket_id"))
	}

	sqlWhere := sqlgo.NewSQLGoWhere()
	sqlWhere.SetSQLWhere("AND", "r.deleted", "=", false)

	if query.Search != "" {
		searchPattern := "%" + query.Search + "%"
		sqlWhere.SetSQLWhere("AND", "r.name", "ILIKE", searchPattern)
		sqlWhere.SetSQLWhere("OR", "r.email", "ILIKE", searchPattern)
		sqlWhere.SetSQLWhere("OR", "o.order_number", "ILIKE", searchPattern)
	}

	if len(query.IDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "r.id", "IN", query.IDs)
	}

	if len(query.EventIDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "r.event_id", "IN", query.EventIDs)
	}

	if len(query.UniqueCodes) > 0 {
		sqlWhere.SetSQLWhere("AND", "r.unique_code", "IN", query.UniqueCodes)
	}

	if len(query.Emails) > 0 {
		sqlWhere.SetSQLWhere("AND", "r.email", "IN", query.Emails)
	}

	if len(query.Statuses) > 0 {
		sqlWhere.SetSQLWhere("AND", "r.status", "IN", query.Statuses)
	}

	if len(query.PaymentStatus) > 0 {
		sqlWhere.SetSQLWhere("AND", "o.payment_status", "IN", query.PaymentStatus)
	}

	if len(query.TicketTypes) > 0 {
		sqlWhere.SetSQLWhere("AND", "rt.type", "IN", query.TicketTypes)
		sqlWhere.SetSQLWhere("OR", "at.type", "IN", query.TicketTypes)
	}

	if query.DateStart != "" {
		sqlWhere.SetSQLWhere("AND", "DATE(r.created_at)", ">=", query.DateStart)
	}

	if query.DateEnd != "" {
		sqlWhere.SetSQLWhere("AND", "DATE(r.created_at)", "<=", query.DateEnd)
	}

	sortBy := "r.created_at"
	switch query.SortBy {
	case "total_cost":
		sortBy = "r.total_cost"
	case "name":
		sortBy = "r.name"
	case "email":
		sortBy = "r.email"
	}

	sortOrder := "DESC"
	if query.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	sqlOrder := sqlgo.NewSQLGoOrder()
	sqlOrder.SetSQLOrder(sortBy, sortOrder)

	sqlOffsetLimit := sqlgo.NewSQLGoOffsetLimit()
	if !query.PagingQuery.NoLimit {
		if query.PagingQuery.Page > 0 {
			sqlOffsetLimit.SQLPageLimit(query.PagingQuery.Page.Int(), query.PagingQuery.Limit.Int())
		} else {
			sqlOffsetLimit.SetSQLLimit(query.PagingQuery.Limit.Int())
		}
	}

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoSelect(sqlSelect).
		SetSQLGoFrom(sqlFrom).
		SetSQLGoJoin(sqlJoin).
		SetSQLGoWhere(sqlWhere).
		SetSQLGoOrder(sqlOrder).
		SetSQLGoOffsetLimit(sqlOffsetLimit)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "registrantDAO.Search",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	rows, err := d.dbTrx.GetSqlDB().QueryContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "registrantDAO.Search",
			zap.String("SQL", sqlStr),
			zap.Any("Params", sqlParams),
			zap.Error(err),
		)
		return nil, 0, err
	}
	defer rows.Close()

	var registrants entity.Registrants
	for rows.Next() {
		var reg entity.Registrant

		if err := rows.Scan(
			&reg.ID,
			&reg.EventID,
			&reg.UniqueCode,
			&reg.TicketID,
			&reg.Name,
			&reg.Email,
			&reg.Phone,
			&reg.Gender,
			&reg.Birthdate,
			&reg.TotalCost,
			&reg.TotalTickets,
			&reg.Status,
			&reg.DaoEntity.DataHash,
			&reg.DaoEntity.Deleted,
			&reg.DaoEntity.CreatedAt,
			&reg.DaoEntity.UpdatedAt,
		); err != nil {
			d.log.Error(ctx, "registrantDAO.Search.Scan", zap.Error(err))
			return nil, 0, err
		}

		registrants = append(registrants, reg)
	}

	totalCount, err := d.Count(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	return registrants, totalCount, nil
}

func (d registrantDAO) Count(ctx context.Context, query entity.RegistrantQuery) (int, error) {
	sqlSelect := sqlgo.NewSQLGoSelect()
	sqlSelect.SetSQLSelect("COUNT(DISTINCT r.id)", "count")

	sqlFrom := sqlgo.NewSQLGoFrom()
	sqlFrom.SetSQLFrom("registrants", "r")

	sqlJoin := sqlgo.NewSQLGoJoin()
	if query.Search != "" || len(query.PaymentStatus) > 0 {
		sqlJoin.SetSQLJoin("LEFT", "orders", "o", sqlgo.SetSQLJoinWhere("AND", "o.registrant_id", "=", "r.id"))
	}
	if len(query.TicketTypes) > 0 {
		sqlJoin.SetSQLJoin("LEFT", "tickets", "rt", sqlgo.SetSQLJoinWhere("AND", "rt.id", "=", "r.ticket_id"))
		sqlJoin.SetSQLJoin("LEFT", "attendees", "a", sqlgo.SetSQLJoinWhere("AND", "a.registrant_id", "=", "r.id"))
		sqlJoin.SetSQLJoin("LEFT", "tickets", "at", sqlgo.SetSQLJoinWhere("AND", "at.id", "=", "a.ticket_id"))
	}

	sqlWhere := sqlgo.NewSQLGoWhere()
	sqlWhere.SetSQLWhere("AND", "r.deleted", "=", false)

	if query.Search != "" {
		searchPattern := "%" + query.Search + "%"
		sqlWhere.SetSQLWhere("AND", "r.name", "ILIKE", searchPattern)
		sqlWhere.SetSQLWhere("OR", "r.email", "ILIKE", searchPattern)
		sqlWhere.SetSQLWhere("OR", "o.order_number", "ILIKE", searchPattern)
	}

	if len(query.IDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "r.id", "IN", query.IDs)
	}

	if len(query.EventIDs) > 0 {
		sqlWhere.SetSQLWhere("AND", "r.event_id", "IN", query.EventIDs)
	}

	if len(query.UniqueCodes) > 0 {
		sqlWhere.SetSQLWhere("AND", "r.unique_code", "IN", query.UniqueCodes)
	}

	if len(query.Emails) > 0 {
		sqlWhere.SetSQLWhere("AND", "r.email", "IN", query.Emails)
	}

	if len(query.Statuses) > 0 {
		sqlWhere.SetSQLWhere("AND", "r.status", "IN", query.Statuses)
	}

	if len(query.PaymentStatus) > 0 {
		sqlWhere.SetSQLWhere("AND", "o.payment_status", "IN", query.PaymentStatus)
	}

	if len(query.TicketTypes) > 0 {
		sqlWhere.SetSQLWhere("AND", "rt.type", "IN", query.TicketTypes)
		sqlWhere.SetSQLWhere("OR", "at.type", "IN", query.TicketTypes)
	}

	if query.DateStart != "" {
		sqlWhere.SetSQLWhere("AND", "DATE(r.created_at)", ">=", query.DateStart)
	}

	if query.DateEnd != "" {
		sqlWhere.SetSQLWhere("AND", "DATE(r.created_at)", "<=", query.DateEnd)
	}

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoSelect(sqlSelect).
		SetSQLGoFrom(sqlFrom).
		SetSQLGoJoin(sqlJoin).
		SetSQLGoWhere(sqlWhere)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "registrantDAO.Count",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	var totalCount int
	err := d.dbTrx.GetSqlDB().QueryRowContext(ctx, sqlStr, sqlParams...).Scan(&totalCount)
	if err != nil {
		d.log.Error(ctx, "registrantDAO.Count",
			zap.String("SQL", sqlStr),
			zap.Any("Params", sqlParams),
			zap.Error(err),
		)
		return 0, err
	}

	return totalCount, nil
}

func (d registrantDAO) Insert(ctx context.Context, registrants entity.Registrants) error {

	if len(registrants) < 1 {
		return fmt.Errorf("empty registrant data")
	}

	sqlInsert := sqlgo.NewSQLGoInsert().
		SetSQLInsert("registrants").
		SetSQLInsertColumn(
			"id",
			"event_id",
			"unique_code",
			"ticket_id",
			"name",
			"email",
			"phone",
			"gender",
			"birthdate",
			"total_cost",
			"total_tickets",
			"status",
			"data_hash",
			"created_at",
		)

	for i, reg := range registrants {
		reg.CreatedAt = time.Now()

		if reg.ID == "" {
			reg.ID = pubEntity.MakeUUID(
				reg.UniqueCode,
				reg.Email,
				reg.CreatedAt.String(),
			)
		}

		sqlInsert.SetSQLInsertValue(
			reg.ID,
			reg.EventID,
			reg.UniqueCode,
			reg.TicketID,
			reg.Name,
			reg.Email,
			reg.Phone,
			reg.Gender,
			reg.Birthdate,
			reg.TotalCost,
			reg.TotalTickets,
			reg.Status,
			reg.DaoEntity.DataHash,
			reg.CreatedAt,
		)

		registrants[i] = reg
	}

	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLGoInsert(sqlInsert)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "registrantDAO.Insert",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
		zap.Int("Len", len(registrants)),
	)

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "registrantDAO.Insert",
			zap.String("SQL", sqlStr),
			zap.Any("Params", sqlParams),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (d registrantDAO) Update(ctx context.Context, registrants entity.Registrants) error {

	if len(registrants) < 1 {
		return fmt.Errorf("empty registrant data")
	}

	for i, reg := range registrants {
		now := time.Now()
		reg.UpdatedAt = &now

		sql := sqlgo.NewSQLGo().
			SetSQLSchema("public").
			SetSQLUpdate("registrants").
			SetSQLUpdateValue("event_id", reg.EventID).
			SetSQLUpdateValue("ticket_id", reg.TicketID).
			SetSQLUpdateValue("name", reg.Name).
			SetSQLUpdateValue("email", reg.Email).
			SetSQLUpdateValue("phone", reg.Phone).
			SetSQLUpdateValue("gender", reg.Gender).
			SetSQLUpdateValue("birthdate", reg.Birthdate).
			SetSQLUpdateValue("total_cost", reg.TotalCost).
			SetSQLUpdateValue("total_tickets", reg.TotalTickets).
			SetSQLUpdateValue("status", reg.Status).
			SetSQLUpdateValue("updated_at", reg.UpdatedAt).
			SetSQLWhere("AND", "id", "=", reg.ID)

		sqlStr := sql.BuildSQL()
		sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

		d.log.Debug(ctx, "registrantDAO.Update",
			zap.String("SQL", sqlStr),
			zap.Any("Params", sqlParams),
		)

		_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
		if err != nil {
			d.log.Error(ctx, "registrantDAO.Update",
				zap.String("SQL", sqlStr),
				zap.Any("Params", sqlParams),
				zap.Error(err),
			)
			return err
		}

		registrants[i] = reg
	}

	return nil
}

func (d registrantDAO) Delete(ctx context.Context, id pubEntity.UUID) error {
	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLDelete("registrants").
		SetSQLWhere("AND", "id", "=", id)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "registrantDAO.Delete",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "registrantDAO.Delete",
			zap.String("SQL", sqlStr),
			zap.Any("Params", sqlParams),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (d registrantDAO) GetDashboardSummary(ctx context.Context) (entity.DashboardSummary, error) {
	summary := entity.DashboardSummary{}

	thisMonthStart := time.Now().AddDate(0, 0, -30).Truncate(24 * time.Hour)
	lastMonthStart := time.Now().AddDate(0, -1, -30).Truncate(24 * time.Hour)

	query := `
		WITH this_month AS (
			SELECT 
				COALESCE(SUM(r.total_tickets), 0) as tickets_sold,
				COALESCE(COUNT(DISTINCT r.id), 0) as registrants,
				COALESCE(SUM(CASE WHEN o.payment_status = 'paid' THEN o.amount ELSE 0 END), 0) as revenue
			FROM registrants r
			LEFT JOIN orders o ON o.registrant_id = r.id AND o.deleted = false
			WHERE r.deleted = false AND r.created_at >= $1
		),
		last_month AS (
			SELECT 
				COALESCE(SUM(r.total_tickets), 0) as tickets_sold,
				COALESCE(COUNT(DISTINCT r.id), 0) as registrants,
				COALESCE(SUM(CASE WHEN o.payment_status = 'paid' THEN o.amount ELSE 0 END), 0) as revenue
			FROM registrants r
			LEFT JOIN orders o ON o.registrant_id = r.id AND o.deleted = false
			WHERE r.deleted = false AND r.created_at >= $2 AND r.created_at < $1
		),
		active_events AS (
			SELECT COUNT(DISTINCT event_id) as count
			FROM registrants r
			WHERE r.deleted = false
		)
		SELECT 
			tm.tickets_sold,
			lm.tickets_sold,
			tm.registrants,
			lm.registrants,
			tm.revenue,
			lm.revenue,
			ae.count
		FROM this_month tm, last_month lm, active_events ae
	`

	d.log.Debug(ctx, "registrantDAO.GetDashboardSummary", zap.String("SQL", query))

	var thisTickets, lastTickets, thisRegistrants, lastRegistrants, activeEvents int
	var thisRevenue, lastRevenue float64

	err := d.dbTrx.GetSqlDB().QueryRowContext(ctx, query, thisMonthStart, lastMonthStart).Scan(
		&thisTickets, &lastTickets,
		&thisRegistrants, &lastRegistrants,
		&thisRevenue, &lastRevenue,
		&activeEvents,
	)
	if err != nil {
		d.log.Error(ctx, "registrantDAO.GetDashboardSummary", zap.Error(err))
		return summary, err
	}

	calcChange := func(current, previous int) float64 {
		if previous == 0 {
			return 0
		}
		return float64(current-previous) / float64(previous) * 100
	}

	summary = entity.DashboardSummary{
		TotalTicketsSold:  thisTickets,
		TicketsSoldChange: calcChange(thisTickets, lastTickets),
		TotalRegistrants:  thisRegistrants,
		RegistrantsChange: calcChange(thisRegistrants, lastRegistrants),
		TotalRevenue:      thisRevenue,
		RevenueChange:     calcChange(int(thisRevenue), int(lastRevenue)),
		ActiveEvents:      activeEvents,
	}

	return summary, nil
}

func (d registrantDAO) GetDailySalesTrend(ctx context.Context, days int) (entity.DailySalesTrend, error) {
	if days <= 0 {
		days = 14
	}

	query := `
		SELECT 
			DATE(o.created_at) as sale_date,
			COALESCE(SUM(r.total_tickets), 0) as tickets_sold,
			COALESCE(SUM(CASE WHEN o.payment_status = 'paid' THEN o.amount ELSE 0 END), 0) as revenue
		FROM registrants r
		LEFT JOIN orders o ON o.registrant_id = r.id AND o.deleted = false
		WHERE r.deleted = false AND o.created_at >= NOW() - INTERVAL '` + fmt.Sprintf("%d days", days) + `'
		GROUP BY DATE(o.created_at)
		ORDER BY sale_date ASC
	`

	d.log.Debug(ctx, "registrantDAO.GetDailySalesTrend", zap.String("SQL", query))

	rows, err := d.dbTrx.GetSqlDB().QueryContext(ctx, query)
	if err != nil {
		d.log.Error(ctx, "registrantDAO.GetDailySalesTrend", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var trend entity.DailySalesTrend
	for rows.Next() {
		var ds entity.DailySales
		var saleDate time.Time
		if err := rows.Scan(&saleDate, &ds.TicketsSold, &ds.Revenue); err != nil {
			d.log.Error(ctx, "registrantDAO.GetDailySalesTrend.Scan", zap.Error(err))
			return nil, err
		}
		ds.Date = saleDate.Format("2006-01-02")
		trend = append(trend, ds)
	}

	return trend, nil
}

func (d registrantDAO) GetTicketDistribution(ctx context.Context) (entity.TicketDistributions, error) {
	query := `
		SELECT 
			t.type as ticket_type,
			COALESCE(SUM(r.total_tickets), 0) as tickets_sold,
			COALESCE(t.capacity, 0) as total_capacity
		FROM registrants r
		LEFT JOIN orders o ON o.registrant_id = r.id AND o.deleted = false
		LEFT JOIN tickets t ON t.id = r.ticket_id AND t.deleted = false
		WHERE r.deleted = false
		GROUP BY t.type, t.capacity
		ORDER BY tickets_sold DESC
	`

	d.log.Debug(ctx, "registrantDAO.GetTicketDistribution", zap.String("SQL", query))

	rows, err := d.dbTrx.GetSqlDB().QueryContext(ctx, query)
	if err != nil {
		d.log.Error(ctx, "registrantDAO.GetTicketDistribution", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var distributions entity.TicketDistributions
	var totalSold int
	for rows.Next() {
		var td entity.TicketDistribution
		if err := rows.Scan(&td.TicketType, &td.TicketsSold, &td.TotalCapacity); err != nil {
			d.log.Error(ctx, "registrantDAO.GetTicketDistribution.Scan", zap.Error(err))
			return nil, err
		}
		totalSold += td.TicketsSold
		distributions = append(distributions, td)
	}

	for i := range distributions {
		if totalSold > 0 {
			distributions[i].Percentage = float64(distributions[i].TicketsSold) / float64(totalSold) * 100
		}
	}

	return distributions, nil
}

func (d registrantDAO) GetRecentTransactions(ctx context.Context, limit int) (entity.RecentTransactions, error) {
	if limit <= 0 {
		limit = 5
	}

	query := `
		SELECT 
			o.id as order_id,
			r.name as buyer_name,
			t.type as ticket_type,
			r.total_tickets as quantity,
			o.amount as amount,
			o.payment_status as status,
			r.created_at as created_at
		FROM registrants r
		LEFT JOIN orders o ON o.registrant_id = r.id AND o.deleted = false
		LEFT JOIN tickets t ON t.id = r.ticket_id AND t.deleted = false
		WHERE r.deleted = false
		ORDER BY r.created_at DESC
		LIMIT ` + fmt.Sprintf("%d", limit)

	d.log.Debug(ctx, "registrantDAO.GetRecentTransactions", zap.String("SQL", query))

	rows, err := d.dbTrx.GetSqlDB().QueryContext(ctx, query)
	if err != nil {
		d.log.Error(ctx, "registrantDAO.GetRecentTransactions", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var transactions entity.RecentTransactions
	now := time.Now()
	for rows.Next() {
		var t entity.RecentTransaction
		var createdAt time.Time
		if err := rows.Scan(&t.ID, &t.BuyerName, &t.TicketType, &t.Quantity, &t.Amount, &t.Status, &createdAt); err != nil {
			d.log.Error(ctx, "registrantDAO.GetRecentTransactions.Scan", zap.Error(err))
			return nil, err
		}
		t.TimeAgo = timeSince(createdAt, now)
		transactions = append(transactions, t)
	}

	return transactions, nil
}

func timeSince(t time.Time, now time.Time) string {
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "baru saja"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%d menit lalu", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		return fmt.Sprintf("%d jam lalu", hours)
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d hari lalu", days)
	default:
		return t.Format("02 Jan 2006")
	}
}

func (d registrantDAO) SoftDelete(ctx context.Context, id pubEntity.UUID) error {
	sql := sqlgo.NewSQLGo().
		SetSQLSchema("public").
		SetSQLUpdate("registrants").
		SetSQLUpdateValue("deleted", true).
		SetSQLWhere("AND", "id", "=", id)

	sqlStr := sql.BuildSQL()
	sqlParams := sql.GetSQLGoParameter().GetSQLParameter()

	d.log.Debug(ctx, "registrantDAO.SoftDelete",
		zap.String("SQL", sqlStr),
		zap.Any("Params", sqlParams),
	)

	_, err := d.dbTrx.GetSqlTx().ExecContext(ctx, sqlStr, sqlParams...)
	if err != nil {
		d.log.Error(ctx, "registrantDAO.SoftDelete",
			zap.String("SQL", sqlStr),
			zap.Any("Params", sqlParams),
			zap.Error(err),
		)
		return err
	}

	return nil
}

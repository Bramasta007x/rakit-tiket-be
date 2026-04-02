package app_registrant

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"rakit-tiket-be/pkg/entity"
	orderEntity "rakit-tiket-be/pkg/entity/app_order"
	appRegistrantEntity "rakit-tiket-be/pkg/entity/app_registrant"
	ticketEntity "rakit-tiket-be/pkg/entity/app_ticket"
	httpModel "rakit-tiket-be/pkg/model/http"
)

type RegistrantData struct {
	TicketID  entity.UUID `json:"ticket_id" validate:"required"`
	Name      string      `json:"name" validate:"required"`
	Email     string      `json:"email" validate:"required,email"`
	Phone     string      `json:"phone" validate:"required"`
	Gender    *string     `json:"gender"`
	Birthdate *string     `json:"birthdate"`
	Document  *string     `json:"document"`
}

type AttendeeData struct {
	TicketID  entity.UUID `json:"ticket_id" validate:"required"`
	Name      string      `json:"name" validate:"required"`
	Gender    *string     `json:"gender"`
	Birthdate *string     `json:"birthdate"`
	Document  *string     `json:"document"`
}

type RegisterRequest struct {
	Registrant RegistrantData `json:"registrant" validate:"required"`
	Attendees  []AttendeeData `json:"attendees"`
}

type RegisterResponse struct {
	Order      OrderInfo      `json:"order"`
	Registrant RegistrantInfo `json:"registrant"`
}

type OrderInfo struct {
	OrderID       string  `json:"order_id"`
	OrderNumber   string  `json:"order_number"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	PaymentStatus string  `json:"payment_status"`
	PaymentToken  string  `json:"payment_token"`
	RedirectURL   string  `json:"redirect_url"`
}

type RegistrantInfo struct {
	ID         string `json:"id"`
	UniqueCode string `json:"unique_code"`
}

type SummaryResponseModel struct {
	httpModel.HTTPResponseModel
	Data SummaryData `json:"data"`
}

type SummaryData struct {
	TotalRegistrants   int     `json:"total_registrants"`
	TotalAttendees     int     `json:"total_attendees"`
	TotalTickets       int     `json:"total_tickets"`
	TotalRevenue       float64 `json:"total_revenue"`
	PaidRegistrants    int     `json:"paid_registrants"`
	PendingRegistrants int     `json:"pending_registrants"`
	FailedRegistrants  int     `json:"failed_registrants"`
}

func MakeSummaryResponseModel(httpCode int, summary SummaryData) (int, SummaryResponseModel) {
	return httpCode, SummaryResponseModel{
		HTTPResponseModel: httpModel.HTTPResponseModel{
			Code: httpCode,
		},
		Data: summary,
	}
}

type (
	SearchRegistrantsRequestModel struct {
		httpModel.HTTPRequestModel
		appRegistrantEntity.RegistrantQuery
	}

	SearchRegistrantsResponseModel struct {
		httpModel.HTTPResponseModel
		Data RegistrantsModel `json:"data"`
	}
)

func MakeSearchRegistrantsResponseModel(httpCode, count int, registrants appRegistrantEntity.Registrants, orderMap map[string]orderEntity.Order, ticketMap map[string]ticketEntity.Ticket, attendeeMap map[string][]appRegistrantEntity.Attendee, baseURL string) (int, SearchRegistrantsResponseModel) {
	return httpCode, SearchRegistrantsResponseModel{
		HTTPResponseModel: httpModel.HTTPResponseModel{
			Code:  httpCode,
			Count: count,
		},
		Data: MakeRegistrantsModelFromEntity(registrants, orderMap, ticketMap, attendeeMap, baseURL),
	}
}

func (m SearchRegistrantsRequestModel) BuildUrlValues() url.Values {
	urlValues := url.Values{}
	return httpModel.BuildDaoFieldUrlValues(m.HTTPRequestModel.BuildUrlValues(urlValues), m.DaoQuery)
}

type PaymentInfo struct {
	ID            string     `json:"id"`
	Status        string     `json:"status"`
	Method        string     `json:"method"`
	PaymentMethod string     `json:"payment_method"`
	Time          *time.Time `json:"time"`
	Total         float64    `json:"total"`
}

type RegistrantInfoDetail struct {
	Name        string  `json:"name"`
	TicketTitle string  `json:"ticket_title"`
	TicketType  string  `json:"ticket_type"`
	Email       string  `json:"email"`
	Phone       string  `json:"phone"`
	Gender      string  `json:"gender"`
	Birthdate   *string `json:"birthdate"`
	ETicket     *string `json:"e_ticket"`
}

type AttendeeInfoModel struct {
	TicketTitle string  `json:"ticket_title"`
	TicketType  string  `json:"ticket_type"`
	Name        string  `json:"name"`
	Gender      string  `json:"gender"`
	Birthdate   *string `json:"birthdate"`
}

type RegistrantModel struct {
	UniqueID     string               `json:"unique_id"`
	Payment      PaymentInfo          `json:"payment"`
	OrderNumber  string               `json:"order_number"`
	Registrant   RegistrantInfoDetail `json:"registrant"`
	Attendees    []AttendeeInfoModel  `json:"attendees"`
	TotalTickets int                  `json:"total_tickets"`
	TicketTypes  []string             `json:"ticket_types"`
	TicketTitles []string             `json:"ticket_titles"`
}

type RegistrantsModel []RegistrantModel

func MakeRegistrantsModelFromEntity(registrants appRegistrantEntity.Registrants, orderMap map[string]orderEntity.Order, ticketMap map[string]ticketEntity.Ticket, attendeeMap map[string][]appRegistrantEntity.Attendee, baseURL string) RegistrantsModel {
	var result RegistrantsModel
	for _, reg := range registrants {
		order, hasOrder := orderMap[string(reg.ID)]
		attendees := attendeeMap[string(reg.ID)]
		result = append(result, MakeRegistrantModelFromEntity(reg, order, hasOrder, ticketMap, attendees, baseURL))
	}
	return result
}

func MakeRegistrantModelFromEntity(reg appRegistrantEntity.Registrant, order orderEntity.Order, hasOrder bool, ticketMap map[string]ticketEntity.Ticket, attendees []appRegistrantEntity.Attendee, baseURL string) RegistrantModel {
	var regGender string
	if reg.Gender != nil {
		regGender = *reg.Gender
	}

	var regBirthdate *string
	if reg.Birthdate != nil {
		t := reg.Birthdate.Format("2006-01-02")
		regBirthdate = &t
	}

	var ticketTitle, ticketType string
	if reg.TicketID != nil {
		if t, ok := ticketMap[string(*reg.TicketID)]; ok {
			ticketTitle = t.Title
			ticketType = t.Type
		}
	}
	if ticketTitle == "" {
		ticketTitle = "-"
	}
	if ticketType == "" {
		ticketType = "-"
	}

	var paymentInfo PaymentInfo
	if hasOrder {
		var formattedPaymentMethod string
		if order.PaymentChannel != nil {
			formattedPaymentMethod = strings.Title(strings.ReplaceAll(strings.ReplaceAll(*order.PaymentChannel, "_", " "), "-", " "))
		} else {
			formattedPaymentMethod = "-"
		}

		var paymentMethod string
		if order.PaymentMethod != nil {
			paymentMethod = *order.PaymentMethod
		}

		paymentInfo = PaymentInfo{
			ID:            string(order.ID),
			Status:        order.PaymentStatus,
			Method:        paymentMethod,
			PaymentMethod: formattedPaymentMethod,
			Time:          order.PaymentTime,
			Total:         order.Amount,
		}
	} else {
		paymentInfo = PaymentInfo{
			ID:            "",
			Status:        reg.Status,
			Method:        "",
			PaymentMethod: "-",
			Total:         reg.TotalCost,
		}
	}

	ticketTypes := []string{}
	ticketTitles := []string{}
	if ticketType != "-" {
		ticketTypes = append(ticketTypes, ticketType)
	}
	if ticketTitle != "-" {
		ticketTitles = append(ticketTitles, ticketTitle)
	}

	var attendeeInfos []AttendeeInfoModel
	for _, att := range attendees {
		var attGender string
		if att.Gender != nil {
			attGender = *att.Gender
		}

		var attBirthdate *string
		if att.Birthdate != nil {
			t := att.Birthdate.Format("2006-01-02")
			attBirthdate = &t
		}

		attTicketTitle := "-"
		attTicketType := "-"
		if t, ok := ticketMap[string(att.TicketID)]; ok {
			attTicketTitle = t.Title
			attTicketType = t.Type
			if attTicketType != "-" {
				ticketTypes = append(ticketTypes, attTicketType)
			}
			if attTicketTitle != "-" {
				ticketTitles = append(ticketTitles, attTicketTitle)
			}
		}

		attendeeInfos = append(attendeeInfos, AttendeeInfoModel{
			TicketTitle: attTicketTitle,
			TicketType:  attTicketType,
			Name:        att.Name,
			Gender:      attGender,
			Birthdate:   attBirthdate,
		})
	}

	orderNum := ""
	if hasOrder {
		orderNum = order.OrderNumber
	}

	makePdfName := func(name string) string {
		safeName := strings.ReplaceAll(name, " ", "_")
		return fmt.Sprintf("E-Voucher-%s-%s.pdf", orderNum, safeName)
	}

	var eTicket *string
	if orderNum != "" {
		url := baseURL + makePdfName(reg.Name)
		eTicket = &url
	}

	return RegistrantModel{
		UniqueID:    string(reg.ID),
		Payment:     paymentInfo,
		OrderNumber: orderNum,
		Registrant: RegistrantInfoDetail{
			Name:        reg.Name,
			TicketTitle: ticketTitle,
			TicketType:  ticketType,
			Email:       reg.Email,
			Phone:       reg.Phone,
			Gender:      regGender,
			Birthdate:   regBirthdate,
			ETicket:     eTicket,
		},
		Attendees:    attendeeInfos,
		TotalTickets: reg.TotalTickets,
		TicketTypes:  ticketTypes,
		TicketTitles: ticketTitles,
	}
}

func (m RegistrantsModel) GetRegistrantsEntity() appRegistrantEntity.Registrants {
	var result appRegistrantEntity.Registrants
	for _, regModel := range m {
		result = append(result, appRegistrantEntity.Registrant{
			ID: entity.UUID(regModel.UniqueID),
		})
	}
	return result
}

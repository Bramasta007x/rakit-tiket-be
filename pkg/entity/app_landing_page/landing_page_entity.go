package entity

import (
	pubEntity "rakit-tiket-be/pkg/entity"
)

type TicketInfo struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Description   *string `json:"description"`
	Price         float64 `json:"price"`
	Total         int     `json:"total"`
	Remaining     int     `json:"remaining"`
	IsPresale     bool    `json:"is_presale"`
	OrderPriority int     `json:"order_priority"`
}

type (
	LandingPageQuery struct {
		IDs       []string `query:"id"`
		EventIDs  []string `query:"event_id"`
		EventName []string `query:"event_name"`
	}

	LandingPage struct {
		ID      pubEntity.UUID `json:"id"`
		EventID pubEntity.UUID `json:"event_id"`

		// Event Info
		EventName      string  `json:"event_name"`
		EventSubtitle  *string `json:"event_subtitle"`
		EventCreator   *string `json:"event_creator"`
		EventDate      string  `json:"event_date"`
		EventTimeStart string  `json:"event_time_start"`
		EventTimeEnd   string  `json:"event_time_end"`
		EventLocation  *string `json:"event_location"`
		LogoImage      *string `json:"logo_image"`

		// Hero Section
		HeroID         *string `json:"hero_id"`
		BannerImage    *string `json:"banner_image"`
		BannerColor    *string `json:"banner_color"`
		HeroButtonID   *string `json:"hero_button_id"`
		HeroButtonText *string `json:"hero_button_text"`
		HeroButtonLink *string `json:"hero_button_link"`

		// Countdown
		HeroCountdownID        *string `json:"hero_countdown_id"`
		HeroCountdownDate      *string `json:"hero_countdown_date"`
		HeroCountdownTimeStart *string `json:"hero_countdown_time_start"`
		HeroCountdownTimeEnd   *string `json:"hero_countdown_time_end"`
		HeroCountdownAfterText *string `json:"hero_countdown_after_text"`

		// Venue Section
		VenueID      *string `json:"venue_id"`
		VenueImage   *string `json:"venue_image"`
		VenueLayout  *string `json:"venue_layout"`
		VenueAddress *string `json:"venue_address"`
		VenueMapLink *string `json:"venue_map_link"`
		VenueGoogle  *string `json:"venue_google"`

		// Ticket Section
		TicketID          *string      `json:"ticket_id"`
		TicketTitle       *string      `json:"ticket_title"`
		TicketDescription *string      `json:"ticket_description"`
		Tickets           []TicketInfo `json:"tickets"`

		// Artist Section
		ArtistID       *string      `json:"artist_id"`
		ArtistTitle    *string      `json:"artist_title"`
		ArtistSubtitle *string      `json:"artist_subtitle"`
		Artists        []ArtistInfo `json:"artist"`

		// FAQ & Terms
		FAQID                *string  `json:"faq_id"`
		Faqs                 []string `json:"faqs"`
		TermsAndConditionsID *string  `json:"terms_and_conditions_id"`
		TermsAndConditions   []string `json:"terms_and_conditions"`

		pubEntity.DaoEntity
	}

	LandingPages []LandingPage

	ArtistInfo struct {
		ID                pubEntity.UUID      `json:"id"`
		Image             *string             `json:"image"`
		ImageUrl          *string             `json:"imageUrl"`
		Name              string              `json:"name"`
		Genre             string              `json:"genre"`
		ArtistSocialMedia []ArtistSocialMedia `json:"artist_social_media"`
	}

	ArtistSocialMedia struct {
		Link string `json:"link"`
		Name string `json:"name"`
	}
)

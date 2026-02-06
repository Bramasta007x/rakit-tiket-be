package entity

import (
	pubEntity "rakit-tiket-be/pkg/entity"
)

type (
	LandingPageQuery struct {
		IDs       []string `query:"id"`
		EventName []string `query:"event_name"`
	}

	LandingPage struct {
		ID pubEntity.UUID `json:"id"`

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

		TermsAndConditions []string `json:"terms_and_conditions"`
		Faqs               []string `json:"faqs"`

		pubEntity.DaoEntity
	}

	LandingPages []LandingPage
)

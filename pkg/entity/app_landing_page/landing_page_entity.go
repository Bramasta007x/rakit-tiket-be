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

		BannerImage *string `json:"banner_image"`
		VenueImage  *string `json:"venue_image"`

		EventCreator *string `json:"event_creator"`
		EventName    string  `json:"event_name"`

		EventDate      string `json:"event_date"`
		EventTimeStart string `json:"event_time_start"`
		EventTimeEnd   string `json:"event_time_end"`

		EventLocation *string `json:"event_location"`

		TermsAndConditions []string `json:"terms_and_conditions"`

		pubEntity.DaoEntity
	}

	LandingPages []LandingPage
)

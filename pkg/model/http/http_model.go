package model

import (
	"net/http"
	"net/url"
	"strconv"

	"rakit-tiket-be/pkg/entity"
)

// Struct used for main rest api response
type (
	HTTPRequestModel struct {
		Page    int  `query:"page"`
		Limit   int  `query:"limit"`
		NoLimit bool `query:"no_limit"`
	}

	HTTPResponseModel struct {
		Code           int         `json:"code,omitempty"`
		Links          LinkMapMap  `json:"links,omitempty"`
		Message        string      `json:"message,omitempty"`
		DisplayMessage string      `json:"display_message,omitempty"`
		Count          int         `json:"totalCount,omitempty"`
		Data           interface{} `json:"data,omitempty"`
	}
)

func MakeHTTPResponseModel(httpCode, count int, data interface{}) HTTPResponseModel {
	return HTTPResponseModel{
		Code:    httpCode,
		Message: http.StatusText(httpCode),
		Data:    data,
		Count:   count,
	}
}

func MakeHTTPMessageResponseModel(httpCode int, displayMessage string, data interface{}) HTTPResponseModel {
	return HTTPResponseModel{
		Code:           httpCode,
		Message:        http.StatusText(httpCode),
		DisplayMessage: displayMessage,
		Data:           data,
	}
}

func (httpRequest HTTPRequestModel) BuildUrlValues(urlValues url.Values) url.Values {
	if httpRequest.Page > 0 {
		urlValues.Add("page", strconv.Itoa(httpRequest.Page))
	}
	if httpRequest.Limit > 0 {
		urlValues.Add("limit", strconv.Itoa(httpRequest.Limit))
	}
	if httpRequest.NoLimit {
		urlValues.Add("no_limit", strconv.FormatBool(httpRequest.NoLimit))
	}
	return urlValues
}

func (httpResponse HTTPResponseModel) BuildLinks(uri *url.URL, linkRel string) HTTPResponseModel {
	httpResponse.Links = httpResponse.Links.
		AddLink(uri.Path, "next", http.MethodGet, uri.Query()).
		AddLink(uri.Path, "prev", http.MethodGet, uri.Query())
	return httpResponse
}

func (httpResponse HTTPResponseModel) SetTotalCount(count int) HTTPResponseModel {
	httpResponse.Count = count
	return httpResponse
}

func BuildDaoFieldUrlValues(urlValues url.Values, daoField entity.DaoQuery) url.Values {
	if len(daoField.DataHashes) > 0 {
		for _, field := range daoField.DataHashes {
			urlValues.Add("data_hash", field)
		}
	}
	if !daoField.CreatedAtStart.IsZero() {
		urlValues.Add("created_at_start", daoField.CreatedAtStart.String())
	}
	if !daoField.CreatedAtEnd.IsZero() {
		urlValues.Add("created_at_end", daoField.CreatedAtEnd.String())
	}
	if !daoField.UpdatedAtStart.IsZero() {
		urlValues.Add("updated_at_start", daoField.CreatedAtStart.String())
	}
	if !daoField.UpdatedAtEnd.IsZero() {
		urlValues.Add("updated_at_end", daoField.UpdatedAtEnd.String())
	}
	return urlValues
}

func BuildDaoRelationUrlValues(urlValues url.Values, relationField entity.RelationQuery) url.Values {
	if len(relationField.RelationIDs) > 0 {
		for _, relationID := range relationField.RelationIDs {
			urlValues.Add("relation_id", relationID.String())
		}
	}
	if len(relationField.RelationSources) > 0 {
		for _, relationSource := range relationField.RelationSources {
			urlValues.Add("relation_source", relationSource)
		}
	}
	return urlValues
}

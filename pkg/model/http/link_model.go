package model

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"rakit-tiket-be/pkg/entity"
	"rakit-tiket-be/pkg/util"
)

type (
	Link struct {
		Href       string          `json:"href,omitempty"`
		Rel        string          `json:"rel,omitempty"`
		Type       string          `json:"type,omitempty"`
		Query      url.Values      `json:"query,omitempty"`
		Pagination PaginationModel `json:"pagination,omitempty"`
	}

	LinkSlice    []Link
	LinkMap      map[string]Link
	LinkMapMap   map[string]LinkMap
	LinkMapSlice map[string]LinkSlice
)

func (l LinkMapMap) AddLink(linkHref string, linkRel string, linkType string, linkQuery url.Values) LinkMapMap {
	if l == nil {
		l = make(LinkMapMap)
	}
	linkType = strings.ToLower(linkType)
	if _, ok := l[linkType]; !ok {
		l[linkType] = make(LinkMap)
	}
	var link Link
	link = link.BuildLink(linkHref, linkRel, linkType, linkQuery)
	if link.Href != "" {
		l[linkType][linkRel] = link
	}

	return l
}

func (Link) BuildLink(linkHref string, linkRel string, linkType string, linkQuery url.Values) Link {
	link := Link{
		Href:       linkHref,
		Rel:        linkRel,
		Type:       linkType,
		Query:      linkQuery,
		Pagination: MakePaginationModel(0, 0, false),
	}
	var page, limit, nextPage, prevPage int
	if query, ok := linkQuery["page"]; ok && len(query) > 0 {
		if val, err := strconv.Atoi(query[0]); err == nil {
			page = val
			nextPage = page + 1
			prevPage = page - 1
		}
		if linkRel == "prev" && prevPage <= 0 {
			return Link{}
		}
	}
	if query, ok := linkQuery["limit"]; ok && len(query) > 0 {
		if val, err := strconv.Atoi(query[0]); err == nil {
			limit = val
		}
	}

	if len(linkQuery) > 0 {
		count := 0
		for key, values := range linkQuery {
			for i, value := range values {
				if count > 0 {
					link.Href = fmt.Sprintf("%s&", link.Href)
				} else {
					link.Href = fmt.Sprintf("%s?", link.Href)
				}
				if key == "page" {
					if linkRel == "next" {
						value = strconv.Itoa(nextPage)
					} else if linkRel == "prev" {
						value = strconv.Itoa(prevPage)
					}
					link.Pagination.Page = entity.Page(util.AtoI(value))
					link.Pagination.Limit = entity.Limit(limit)
				}

				link.Href = fmt.Sprintf("%s%s=%s", link.Href, key, value)
				count++
				values[i] = value
			}
			link.Query[key] = values
		}
	}
	return link
}

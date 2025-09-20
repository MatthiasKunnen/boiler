package steamworkshop

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/go-json-experiment/json"
)

type CollectionDetailFileType int

const (
	CollectionDetailFileTypeWorkshopItem CollectionDetailFileType = iota
	// CollectionDetailFileTypeUnknown is unknown, let the author know if you find a collection
	// that uses it.
	CollectionDetailFileTypeUnknown
	CollectionDetailFileTypeCollection
)

type collectionDetailsResponse struct {
	Response collectionDetailsResponseResponse `json:"response"`
}

type collectionDetailsResponseResponse struct {
	Result            int     `json:"result"`
	ResultCount       int     `json:"resultcount"`
	CollectionDetails []cdrcd `json:"collectiondetails"`
}

type cdrcd struct {
	PublishedFileId uint64                  `json:"publishedfileid,string"`
	Result          int                     `json:"result"`
	Children        []collectionDetailChild `json:"children"`
}

type collectionDetailChild struct {
	PublishedFileId uint64                   `json:"publishedfileid,string"`
	SortOrder       int                      `json:"sortorder"`
	FileType        CollectionDetailFileType `json:"filetype"`
}

type CollectionDetailApi struct {
	CollectionId uint64
	Items        []CollectionDetailItem
}

type CollectionDetailItem struct {
	Id        uint64
	SortOrder int
	Type      CollectionDetailFileType
}

// CollectionDetailsApi returns the details of the collection according to the
// [GetCollectionDetails] API endpoint.
// The response contains the details in the same order as the input.
// The child items are sorted according to the sort order.
//
// [GetCollectionDetails]: https://api.steampowered.com/ISteamRemoteStorage/GetCollectionDetails/v1/
func CollectionDetailsApi(ctx context.Context, collectionIds ...uint64) ([]CollectionDetailApi, error) {
	data := url.Values{}
	data.Set("collectioncount", strconv.Itoa(len(collectionIds)))
	for i, id := range collectionIds {
		data.Set(fmt.Sprintf("publishedfileids[%d]", i), strconv.FormatUint(id, 10))
	}

	client := &http.Client{}
	r, _ := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.steampowered.com/ISteamRemoteStorage/GetCollectionDetails/v1/",
		strings.NewReader(data.Encode()),
	)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Accept", "application/json")

	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return CollectionDetailsApiFromReader(resp.Body, collectionIds...)
}

func CollectionDetailsApiFromReader(
	r io.Reader,
	collectionIds ...uint64,
) ([]CollectionDetailApi, error) {
	var response collectionDetailsResponse
	err := json.UnmarshalRead(r, &response)
	if err != nil {
		return nil, err
	}
	if len(response.Response.CollectionDetails) != len(collectionIds) {
		return nil, fmt.Errorf("expected %d results, got %d", len(collectionIds), response.Response.ResultCount)
	}

	result := make([]CollectionDetailApi, len(response.Response.CollectionDetails))
	for _, detail := range response.Response.CollectionDetails {
		var item CollectionDetailApi
		item.CollectionId = detail.PublishedFileId
		item.Items = make([]CollectionDetailItem, 0)
		for _, child := range detail.Children {
			item.Items = append(item.Items, CollectionDetailItem{
				Id:        child.PublishedFileId,
				Type:      child.FileType,
				SortOrder: child.SortOrder,
			})
		}
		slices.SortFunc(item.Items, func(a, b CollectionDetailItem) int {
			return a.SortOrder - b.SortOrder
		})
		index := slices.Index(collectionIds, detail.PublishedFileId)
		if index < 0 {
			return nil, fmt.Errorf("unexpected collection returned %d", detail.PublishedFileId)
		}
		result[index] = item
	}

	return result, err
}

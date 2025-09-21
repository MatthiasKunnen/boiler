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
	"time"

	"github.com/go-json-experiment/json"
)

type getPublishedFileDetailsResponse struct {
	Response getPublishedFileDetailsResponseInner `json:"response"`
}
type getPublishedFileDetailsResponseInner struct {
	Result               int             `json:"result"`
	ResultCount          int             `json:"resultcount"`
	PublishedFileDetails []fileDetailApi `json:"publishedfiledetails"`
}

type fileDetailApi struct {
	CreatorAppId int    `json:"creator_app_id"`
	Id           uint64 `json:"publishedfileid,string"`
	TimeCreated  int64  `json:"time_created"`
	TimeUpdated  int64  `json:"time_updated"`
	Title        string `json:"title"`
}
type FileDetailApi struct {
	// The ID of the game that the workshop item relates to.
	CreatorAppId int
	Id           uint64
	TimeCreated  time.Time
	TimeUpdated  time.Time
	Title        string
}

// FileDetailsApi returns the details of the workshop items according to
// the [GetPublishedFileDetails] API endpoint.
// The response contains the details in the same order as the input.
//
// [GetPublishedFileDetails]: https://partner.steamgames.com/doc/webapi/ISteamRemoteStorage#GetPublishedFileDetails
func FileDetailsApi(ctx context.Context, ids ...uint64) ([]FileDetailApi, error) {
	data := url.Values{}
	data.Set("itemcount", strconv.Itoa(len(ids)))
	for i, id := range ids {
		data.Set(fmt.Sprintf("publishedfileids[%d]", i), strconv.FormatUint(id, 10))
	}

	client := &http.Client{}
	r, _ := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.steampowered.com/ISteamRemoteStorage/GetPublishedFileDetails/v1/",
		strings.NewReader(data.Encode()),
	)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Accept", "application/json")

	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return FileDetailsApiFromReader(resp.Body, ids...)
}

// FileDetailsApiFromReader parses a response as received from the [GetPublishedFileDetails] API endpoint.
// The response contains the details in the same order as the input.
//
// [GetPublishedFileDetails]: https://partner.steamgames.com/doc/webapi/ISteamRemoteStorage#GetPublishedFileDetails
func FileDetailsApiFromReader(r io.Reader, ids ...uint64) ([]FileDetailApi, error) {
	var response getPublishedFileDetailsResponse
	err := json.UnmarshalRead(r, &response)
	if err != nil {
		return nil, err
	}
	if len(response.Response.PublishedFileDetails) != len(ids) {
		return nil, fmt.Errorf(
			"expected %d results, got %d",
			len(ids),
			len(response.Response.PublishedFileDetails),
		)
	}

	result := make([]FileDetailApi, len(response.Response.PublishedFileDetails))
	for _, detail := range response.Response.PublishedFileDetails {
		index := slices.Index(ids, detail.Id)
		if index < 0 {
			return nil, fmt.Errorf("unexpected file detail returned %d", detail.Id)
		}
		result[index] = FileDetailApi{
			CreatorAppId: detail.CreatorAppId,
			Id:           detail.Id,
			TimeCreated:  time.Unix(detail.TimeCreated, 0),
			TimeUpdated:  time.Unix(detail.TimeUpdated, 0),
			Title:        detail.Title,
		}
	}

	return result, err
}

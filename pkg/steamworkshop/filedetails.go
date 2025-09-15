package steamworkshop

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
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
	Id          string `json:"publishedfileid"`
	TimeCreated int64  `json:"time_created"`
	TimeUpdated int64  `json:"time_updated"`
	Title       string `json:"title"`
}
type FileDetailApi struct {
	Id          string
	TimeCreated time.Time
	TimeUpdated time.Time
	Title       string
}

// FileDetailsApi returns the details of the workshop items according to
// the [GetPublishedFileDetails] API endpoint.
// The response contains the details in the same order as the input.
//
// [GetPublishedFileDetails]: https://api.steampowered.com/ISteamRemoteStorage/GetPublishedFileDetails/v1/
func FileDetailsApi(ctx context.Context, ids ...string) ([]FileDetailApi, error) {
	data := url.Values{}
	data.Set("itemcount", strconv.Itoa(len(ids)))
	for i, id := range ids {
		data.Set(fmt.Sprintf("publishedfileids[%d]", i), id)
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
	var response getPublishedFileDetailsResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
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
			return nil, fmt.Errorf("unexpected file detail returned %s", detail.Id)
		}
		result[index] = FileDetailApi{
			Id:          detail.Id,
			TimeCreated: time.Unix(detail.TimeCreated, 0),
			TimeUpdated: time.Unix(detail.TimeUpdated, 0),
			Title:       detail.Title,
		}
	}

	return result, err
}

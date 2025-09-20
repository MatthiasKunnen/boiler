package steamworkshop

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"

	"golang.org/x/net/html"
)

type FileDetailsWeb struct {
	RequiredItems []FileDetailsRequiredItems
	Title         string
}

type FileDetailsRequiredItems struct {
	Id    uint64
	Title string
}

// GetFileDetailsWeb fetches the HTML of the mod and extracts data from it.
func GetFileDetailsWeb(ctx context.Context, id uint64) (FileDetailsWeb, error) {
	client := &http.Client{}
	r, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("https://steamcommunity.com/sharedfiles/filedetails/?id=%d", id),
		nil,
	)
	if err != nil {
		return FileDetailsWeb{}, err
	}
	resp, err := client.Do(r)
	if err != nil {
		return FileDetailsWeb{}, err
	}
	defer resp.Body.Close()

	return ExtractFileDetailsFromHtml(resp.Body)
}

var tagBody = []byte("body")
var classAttr = []byte("class")
var hrefAttr = []byte("href")
var idAttr = []byte("id")
var workshopItemTitleClass = []byte("workshopItemTitle")
var requiredItemsId = []byte("RequiredItems")

func ExtractFileDetailsFromHtml(r io.Reader) (FileDetailsWeb, error) {
	/*
		Title: .workshopItemTitle Text
		RequiredItems:
			id: #RequiredItems > a[href] extract id param
			title: #RequiredItems > a > .requiredItem Text
	*/
	z := html.NewTokenizer(r)
	var inBody bool
	var hasAttr bool
	var tagName []byte
	state := 0
	const (
		nothing = iota
		nextIsTitle
		inRequiredItems
		inRequiredItem
		stop
	)
	result := FileDetailsWeb{}
	// This requires proper HTML (every tag should be closed)
	var nextRequiredItem FileDetailsRequiredItems
	inRequiredItemsTracker := &depthTracker{}
	inRequiredItemTracker := &depthTracker{}
	trackers := []*depthTracker{
		inRequiredItemTracker,
		inRequiredItemsTracker,
	}

	for {
		if state == stop {
			return result, nil
		}

		tokenType := z.Next()
		switch tokenType {
		case html.ErrorToken:
			if z.Err() == io.EOF {
				return result, nil
			}
			return result, fmt.Errorf("tokenizer error: %w", z.Err())
		case html.StartTagToken:
			tagName, hasAttr = z.TagName()
			if !inBody && bytes.Equal(tagBody, tagName) {
				inBody = true
			} else if !inBody {
				continue
			}

			for _, tracker := range trackers {
				tracker.Increase(tagName)
			}

			if !hasAttr {
				break
			}

			for {
				key, val, moreAttr := z.TagAttr()
				if key == nil || val == nil {
					break
				}

				if bytes.Equal(key, classAttr) && bytes.Equal(val, workshopItemTitleClass) {
					state = nextIsTitle
					break
				} else if bytes.Equal(key, idAttr) && bytes.Equal(val, requiredItemsId) {
					state = inRequiredItems
					inRequiredItemsTracker.Reset(slices.Clone(tagName), func() {
						state = stop
					})
					break
				} else if state == inRequiredItems &&
					len(tagName) == 1 &&
					tagName[0] == 'a' &&
					bytes.Equal(key, hrefAttr) {
					parsedUrl, err := url.Parse(string(val))
					if err != nil {
						continue
					}
					id := parsedUrl.Query().Get("id")
					parseUint, err := strconv.ParseUint(id, 10, 64)
					if err != nil {
						break
					}
					nextRequiredItem.Id = parseUint
					state = inRequiredItem
					inRequiredItemTracker.Reset(slices.Clone(tagName), func() {
						state = inRequiredItems
						result.RequiredItems = append(result.RequiredItems, nextRequiredItem)
						nextRequiredItem = FileDetailsRequiredItems{}
					})
					break
				}

				if !moreAttr {
					break
				}
			}
		case html.EndTagToken:
			tagName, _ = z.TagName()
			for _, tracker := range trackers {
				tracker.Decrease(tagName)
			}
		case html.TextToken:
			switch state {
			case nextIsTitle:
				result.Title = string(bytes.TrimSpace(z.Text()))
				state = nothing
			case inRequiredItem:
				nextRequiredItem.Title += string(bytes.TrimSpace(z.Text()))
			}
		}
	}
}

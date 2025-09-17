package steamworkshop_test

import (
	"bytes"
	"github.com/MatthiasKunnen/boiler/pkg/steamworkshop"
	"github.com/stretchr/testify/assert"
	"testing"
)
import _ "embed"

//go:embed testdata/workshop_2950011244.html
var workshopDetail []byte

func TestExtractFileDetailsFromHtml(t *testing.T) {
	actual, err := steamworkshop.ExtractFileDetailsFromHtml(bytes.NewReader(workshopDetail))
	assert.NoError(t, err)
	expected := steamworkshop.FileDetailsWeb{
		RequiredItems: []steamworkshop.FileDetailsRequiredItems{
			{
				Id:    "450814997",
				Title: "CBA_A3",
			},
			{
				Id:    "2291129343",
				Title: "Improved Melee System",
			},
			{
				Id:    "1726376971",
				Title: "Nassau 1715",
			},
			{
				Id:    "1451755886",
				Title: "Max_Women",
			},
			{
				Id:    "1862880106",
				Title: "POLPOX's Base Functions",
			},
			{
				Id:    "1341387001",
				Title: "POLPOX's Artwork Supporter",
			},
		},
		Title: "Sail to South-Eastern Asia",
	}
	assert.Equal(t, expected, actual)
}

func TestGetFileDetailsWeb(t *testing.T) {
	actual, err := steamworkshop.GetFileDetailsWeb(t.Context(), "2950011244")
	assert.NoError(t, err)
	expected := steamworkshop.FileDetailsWeb{
		RequiredItems: []steamworkshop.FileDetailsRequiredItems{
			{
				Id:    "450814997",
				Title: "CBA_A3",
			},
			{
				Id:    "2291129343",
				Title: "Improved Melee System",
			},
			{
				Id:    "1726376971",
				Title: "Nassau 1715",
			},
			{
				Id:    "1451755886",
				Title: "Max_Women",
			},
			{
				Id:    "1862880106",
				Title: "POLPOX's Base Functions",
			},
			{
				Id:    "1341387001",
				Title: "POLPOX's Artwork Supporter",
			},
		},
		Title: "Sail to South-Eastern Asia",
	}
	assert.Equal(t, expected, actual)
}

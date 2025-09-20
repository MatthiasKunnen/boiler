package steamworkshop_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/MatthiasKunnen/boiler/pkg/steamworkshop"
	"github.com/stretchr/testify/assert"

	_ "embed"
)

//go:embed testdata/file_details_463939057_ace.json
var modAceApiResponse []byte

func TestFileDetailsApiFromReader(t *testing.T) {
	actual, err := steamworkshop.FileDetailsApiFromReader(bytes.NewBuffer(modAceApiResponse), 463939057)
	assert.NoError(t, err)
	expected := []steamworkshop.FileDetailApi{
		{
			Id:          463939057,
			TimeCreated: time.Unix(1434653369, 0),
			TimeUpdated: time.Unix(1752589679, 0),
			Title:       "ace",
		},
	}
	assert.Equal(t, expected, actual)
}

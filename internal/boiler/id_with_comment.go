package boiler

import (
	"fmt"
	"strconv"

	"github.com/go-json-experiment/json/jsontext"
)

// IdWithComment is parsed from a [string, string] or string in JSON.
type IdWithComment struct {
	Id      uint64
	Comment string
}

func (c *IdWithComment) MarshalJSONTo(enc *jsontext.Encoder) error {
	if err := enc.WriteToken(jsontext.BeginArray); err != nil {
		return err
	}
	if err := enc.WriteToken(jsontext.String(strconv.FormatUint(c.Id, 10))); err != nil {
		return err
	}
	if err := enc.WriteToken(jsontext.String(c.Comment)); err != nil {
		return err
	}
	return enc.WriteToken(jsontext.EndArray)
}

func (c *IdWithComment) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	kind := dec.PeekKind()

	switch kind {
	case '[':
		// Handle array case - could be [uint64, string] or just [uint64]
		if _, err := dec.ReadToken(); err != nil { // consume '['
			return err
		}

		// Read the first element (ID)
		if dec.PeekKind() == ']' {
			return fmt.Errorf("empty array not supported")
		}

		token, err := dec.ReadToken()
		if err != nil {
			return err
		}

		if token.Kind() == '"' {
			c.Id, err = strconv.ParseUint(token.String(), 10, 64)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("first array element must be a string")
		}

		// Check if there's a second element (comment)
		if dec.PeekKind() != ']' {
			token, err := dec.ReadToken()
			if err != nil {
				return err
			}

			if token.Kind() == '"' {
				c.Comment = token.String()
			} else {
				return fmt.Errorf("second array element must be a string")
			}
		} else {
			// Default empty comment for a single-element array. Must be set to prevent partial
			// overwrite.
			c.Comment = ""
		}

		// Consume the closing ']'
		if _, err := dec.ReadToken(); err != nil {
			return err
		}

		return nil
	case '"': // Only number
		token, err := dec.ReadToken()
		if err != nil {
			return err
		}

		c.Id, err = strconv.ParseUint(token.String(), 10, 64)
		if err != nil {
			return err
		}

		c.Comment = ""
		return nil
	default:
		return fmt.Errorf("unsupported JSON type for IdWithComment")
	}
}

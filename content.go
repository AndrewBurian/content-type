package content_type

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// ContentType represents a weighted quality mime-type for use in HTTP headers
type ContentType struct {
	MediaType  string            // The full text of the type (ex "application/json")
	Type       string            // The major type (ex "application")
	SubType    string            // The sub type (ex "json")
	Parameters map[string]string // Parameter map of any additional qualities (ex "charset=utf-8"
	Quality    float64           // The quality type (q=1) in numeric form
}

// ContentTypeList is an array of content type objects
type ContentTypeList []*ContentType

// ParseRequests pulls content types from the `Content-Type` and `Accept` headers, reconstructing
// them according to RFC 2616. The return value for content may be nil even if no error was set
func ParseRequest(r *http.Request) (content *ContentType, accepts ContentTypeList, err error) {

	// The Content-Type should only be a single entry, so we take the first and disregard
	// any other entries
	content, err = ParseSingle(r.Header.Get("Content-Type"))
	if err != nil {
		return
	}

	// RFC 2616 sec 4.2 allows headers to be split across multiple entries
	// and specifies they should be reconstructed with a comma separator
	fullType := strings.Join(r.Header["Accept"], ",")
	accepts, err = Parse(fullType)
	if err != nil {
		return
	}

	return
}

// Parse multiple content types out of a single comma separated string
func Parse(data string) (ContentTypeList, error) {
	types := make(ContentTypeList, 0, 1)

	for _, entry := range strings.Split(data, ",") {
		t, err := ParseSingle(entry)
		if err != nil {
			return nil, err
		}

		if t != nil {
			types = append(types, t)
		}
	}

	return types, nil

}

// ParseSingle takes a single content type and assumes it is not comma terminated
func ParseSingle(data string) (*ContentType, error) {
	var qSet bool

	if data == "" {
		return nil, nil
	}

	t := &ContentType{
		Parameters: make(map[string]string),
	}

	// split the content type out from it's parameters
	components := strings.Split(data, ";")
	t.MediaType = strings.TrimSpace(components[0])

	// split the media type into type and subtype
	typeParts := strings.Split(t.MediaType, "/")
	if len(typeParts) != 2 {
		return nil, errors.New("Invalid content type [" + t.MediaType + "]")
	}
	t.Type = typeParts[0]
	t.SubType = typeParts[1]

	// Go through the parameters
	for _, param := range components[1:] {
		values := strings.Split(param, "=")
		if len(values) != 2 {
			return nil, errors.New("Malformed parameter [" + param + "]")
		}
		key := strings.TrimSpace(values[0])
		t.Parameters[key] = strings.TrimSpace(values[1])

		// store quality specially
		if key == "q" {
			qual, err := strconv.ParseFloat(t.Parameters[key], 64)
			if err != nil {
				return nil, errors.New("Malformed quality [" + t.Parameters[key] + "]")
			}
			t.Quality = qual
			qSet = true
		}
	}

	// Default Quality is 1 (RFC 7231 Sec. 5.3.1)
	if !qSet {
		t.Quality = 1
	}

	return t, nil
}

func (t *ContentType) String() string {
	buf := bytes.NewBufferString(t.MediaType)
	for key, val := range t.Parameters {
		buf.WriteString("; ")
		buf.WriteString(key + "=" + val)
	}

	if t.Quality != 1 {
		fmt.Fprintf(buf, "; q=%f", t.Quality)
	}

	return buf.String()
}

func (l ContentTypeList) String() string {
	var buf bytes.Buffer

	if len(l) == 0 {
		return ""
	}

	buf.WriteString(l[0].String())

	for _, t := range l {
		buf.WriteString(", ")
		buf.WriteString(t.String())
	}

	return buf.String()
}

// SupportsType checks if the provided content type t is supported by an entry in this list
func (l ContentTypeList) SupportsType(t *ContentType) bool {
	for _, support := range l {

		// major type must match
		if support.Type != "*" && support.Type != t.Type {
			continue
		}

		// sub type must match
		if support.SubType != "*" && support.SubType != t.SubType {
			continue
		}

		// quality must not be 0
		if support.Quality == 0 {
			continue
		}

		return true
	}

	return false
}

// PreferredMatch chooses the best content type based on quality that is supported in options from the list.
// Returns nil if no types are supported.
func (l ContentTypeList) PreferredMatch(options ContentTypeList) *ContentType {
	candidates := make(ContentTypeList, 0, len(options))

	// get the list of mutually supported types
	for _, option := range options {
		if l.SupportsType(option) {
			candidates = append(candidates, option)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].Quality < candidates[j].Quality
	})

	return candidates[len(candidates)-1]
}

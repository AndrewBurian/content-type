package content_type

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseSingle(t *testing.T) {
	str := "text/html"

	ct, err := ParseSingle(str)
	if err != nil {
		t.Fatal(err)
	}

	if ct.MediaType != "text/html" {
		t.Error("Mismatch Media Type", ct.MediaType)
	}

	if ct.Type != "text" {
		t.Error("Mismatch Type", ct.Type)
	}

	if ct.SubType != "html" {
		t.Error("Mismatch SubType", ct.SubType)
	}

	if ct.Quality != 1 {
		t.Error("Wrong quality", ct.Quality)
	}

}

func TestParseSingle2(t *testing.T) {
	str := "text/html; q=0.5; charset=utf-8"

	ct, err := ParseSingle(str)
	if err != nil {
		t.Fatal(err)
	}

	if ct.MediaType != "text/html" {
		t.Error("Mismatch Media Type", ct.MediaType)
	}

	if ct.Type != "text" {
		t.Error("Mismatch Type", ct.Type)
	}

	if ct.SubType != "html" {
		t.Error("Mismatch SubType", ct.SubType)
	}

	if ct.Quality != 0.5 {
		t.Error("Wrong quality", ct.Quality)
	}

	if ct.Parameters["charset"] != "utf-8" {
		t.Error("Wrong parameters", ct.Parameters)
	}
}

func TestParse(t *testing.T) {
	str := "text/plain; q=0.5, text/html, text/x-dvi; q=0.8, text/x-c"

	list, err := Parse(str)
	if err != nil {
		t.Fatal(err)
	}

	if len(list) != 4 {
		t.Fatal("Length mismatch", len(list))
	}

	if list[0].MediaType != "text/plain" {
		t.Error("Mismatch type", list[0].MediaType, "text/plain")
	}

	if list[1].MediaType != "text/html" {
		t.Error("Mismatch type", list[0].MediaType, "text/html")
	}
}

func TestParseRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("Accept", "text/plain")

	ct, accept, err := ParseRequest(req)
	if err != nil {
		t.Fatal(err)
	}

	if ct != nil {
		t.Error("Should have been nil content type")
	}

	if len(accept) != 1 {
		t.Fatal("Should have received 1 accept type", len(accept))
	}

	if accept[0].MediaType != "text/plain" {
		t.Error("Mismatch", accept[0].MediaType, "text/plain")
	}
}

func TestParseRequest2(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Add("Accept", "text/plain")
	req.Header.Add("Content-Type", "application/json")

	ct, accept, err := ParseRequest(req)

	if err != nil {
		t.Fatal(err)
	}

	if ct == nil {
		t.Error("Should not have been nil content type")
	}

	if ct.MediaType != "application/json" {
		t.Error("Mismatch", ct.MediaType, "application/json")
	}

	if len(accept) != 1 {
		t.Fatal("Should have received 1 accept type", len(accept))
	}

	if accept[0].MediaType != "text/plain" {
		t.Error("Mismatch", accept[0].MediaType, "text/plain")
	}
}

func TestContentTypeList_SupportsType(t *testing.T) {
	list, err := Parse("text/plain; q=0.5, text/html, text/x-dvi; q=0.8, text/x-c")
	if err != nil {
		t.Fatal(err)
	}

	target, err := ParseSingle("application/json")

	if list.SupportsType(target) {
		t.Error("Should not have supported application/json")
	}
}

func TestContentTypeList_SupportsType2(t *testing.T) {
	list, err := Parse("text/plain; q=0.5, text/html, text/x-dvi; q=0.8, text/x-c")
	if err != nil {
		t.Fatal(err)
	}

	target, err := ParseSingle("text/x-c")

	if !list.SupportsType(target) {
		t.Error("Should have supported text/x-c")
	}
}


func TestContentTypeList_SupportsType3(t *testing.T) {
	list, err := Parse("text/plain; q=0.5, text/html, text/x-dvi; q=0.8, text/x-c, application/*")
	if err != nil {
		t.Fatal(err)
	}

	target, err := ParseSingle("application/json")

	if !list.SupportsType(target) {
		t.Error("Should have supported application/json")
	}
}

func TestContentTypeList_PreferredMatch(t *testing.T) {
	list, err := Parse("text/plain; q=0.5, text/html, text/x-dvi; q=0.8, text/x-c")
	if err != nil {
		t.Fatal(err)
	}

	options, err := Parse("application/json, text/plain, text/x-dvi")
	if err != nil {
		t.Fatal(err)
	}

	match := list.PreferredMatch(options)
	if match.MediaType != "text/x-dvi" {
		t.Error("Mismatch", match.MediaType, "text/x-dvi")
	}
}

func TestContentTypeList_PreferredMatch2(t *testing.T) {
	list, err := Parse("text/plain; q=0.5, text/html, text/x-dvi; q=0.8, text/x-c")
	if err != nil {
		t.Fatal(err)
	}

	options, err := Parse("application/json, text/plain, text/x-dvi, text/html")
	if err != nil {
		t.Fatal(err)
	}

	match := list.PreferredMatch(options)
	if match.MediaType != "text/html" {
		t.Error("Mismatch", match.MediaType, "text/html")
	}
}

func TestContentTypeList_PreferredMatch3(t *testing.T) {
	list, err := Parse("text/plain; q=0.5, text/html, text/x-dvi; q=0.8, text/x-c, */*; q=0.001")
	if err != nil {
		t.Fatal(err)
	}

	options, err := Parse("application/nothing")
	if err != nil {
		t.Fatal(err)
	}

	match := list.PreferredMatch(options)
	if match.MediaType != "application/nothing" {
		t.Error("Mismatch", match.MediaType, "application/nothing")
	}
}

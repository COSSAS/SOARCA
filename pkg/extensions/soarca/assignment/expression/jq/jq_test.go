package jq

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

const object = `{
  "library": "Riverside Branch",
  "open": true,
  "books": [
    {
      "id": 1,
      "title": "The Left Hand of Darkness",
      "author": "Ursula K. Le Guin",
      "year": 1969,
      "genres": ["sci-fi", "literary"],
      "copies": 3,
      "checked_out": 2
    },
    {
      "id": 2,
      "title": "Dune",
      "author": "Frank Herbert",
      "year": 1965,
      "genres": ["sci-fi"],
      "copies": 5,
      "checked_out": 5
    },
    {
      "id": 3,
      "title": "Beloved",
      "author": "Toni Morrison",
      "year": 1987,
      "genres": ["literary", "historical"],
      "copies": 2,
      "checked_out": 0
    },
    {
      "id": 4,
      "title": "Neuromancer",
      "author": "William Gibson",
      "year": 1984,
      "genres": ["sci-fi", "cyberpunk"],
      "copies": 4,
      "checked_out": 3
    }
  ]
}
`

func TestName(t *testing.T) {
	jq := New()
	assert.Equal(t, jq.GetExpressionName(), "jq")
}

func TestSimpleQuery(t *testing.T) {
	jq := New()
	err, result := jq.Execute(object, ".library")
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "Riverside Branch")
}

func TestLengthQuery(t *testing.T) {
	jq := New()
	err, result := jq.Execute(object, ".books | length")
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "4")
}

const books = `The Left Hand of Darkness
Dune
Beloved
Neuromancer`

func TestMultiItem(t *testing.T) {

	jq := New()
	err, result := jq.Execute(object, ".books[].title")
	assert.Equal(t, err, nil)
	assert.Equal(t, result, books)
}

func TestMap(t *testing.T) {

	jq := New()
	err, result := jq.Execute(object, ".books | map(.checked_out) | add")
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "10")
}

func TestStringField(t *testing.T) {
	jq := New()
	err, result := jq.Execute(object, ".library")
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "Riverside Branch")
}

func TestBool(t *testing.T) {
	jq := New()
	err, result := jq.Execute(object, ".open")
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "true")
}

func TestNullOnMissingField(t *testing.T) {
	jq := New()
	err, result := jq.Execute(object, ".nonexistent")
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "null")
}

func TestIntFromLength(t *testing.T) {
	jq := New()
	err, result := jq.Execute(object, ".books | length")
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "4")
}

func TestFloatAdd(t *testing.T) {
	jq := New()
	err, result := jq.Execute(object, ".books | map(.checked_out) | add")
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "10")
}

func TestNonIntegerFloat(t *testing.T) {
	jq := New()
	err, result := jq.Execute(object, ".books[0].copies / 2")
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "1.5")
}

func TestArrayValue(t *testing.T) {
	jq := New()
	err, result := jq.Execute(object, ".books[0].genres")
	assert.Equal(t, err, nil)
	assert.Equal(t, result, `["sci-fi","literary"]`)
}

func TestObjectValue(t *testing.T) {
	jq := New()
	err, result := jq.Execute(object, ".books[1] | {title, year}")
	assert.Equal(t, err, nil)
	assert.Equal(t, result, `{"title":"Dune","year":1965}`)
}

func TestMultiItemStrings(t *testing.T) {
	jq := New()
	err, result := jq.Execute(object, ".books[].title")
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "The Left Hand of Darkness\nDune\nBeloved\nNeuromancer")
}

func TestMultiItemNumbers(t *testing.T) {
	jq := New()
	err, result := jq.Execute(object, ".books[].copies")
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "3\n5\n2\n4")
}

func TestSelectFilter(t *testing.T) {
	jq := New()
	err, result := jq.Execute(object, `.books[] | select(.checked_out == 0) | .title`)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "Beloved")
}

func TestFlattenAndUnique(t *testing.T) {
	jq := New()
	err, result := jq.Execute(object, `[.books[].genres[]] | unique`)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, `["cyberpunk","historical","literary","sci-fi"]`)
}

func TestBigInteger(t *testing.T) {
	jq := New()
	err, result := jq.Execute(object, "100000000000000000000")
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "100000000000000000000")
}

func TestInvalidJSON(t *testing.T) {
	jq := New()
	err, result := jq.Execute("{not valid json", ".library")
	assert.NotEqual(t, err, nil)
	assert.Equal(t, result, "")
}

func TestParseError(t *testing.T) {
	jq := New()
	err, result := jq.Execute(object, ".books[")
	assert.NotEqual(t, err, nil)
	assert.Equal(t, result, "")
}

func TestRuntimeError(t *testing.T) {
	jq := New()
	err, _ := jq.Execute(object, ".books | .title")
	assert.NotEqual(t, err, nil)
}

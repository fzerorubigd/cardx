package cards

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Attributes is the list of attributes with proper functions for the template engine
type Attributes struct {
	values map[string]string
}

func cloneAttributes(in Attributes) Attributes {
	attr := Attributes{
		values: make(map[string]string),
	}

	for i := range in.values {
		attr.values[i] = in.values[i]
	}

	return attr
}

func (a *Attributes) ensureMap() {
	if a.values == nil {
		a.values = make(map[string]string)
	}
}

// Append add/replace data in attribute group
func (a *Attributes) Append(key, data string) {
	a.ensureMap()
	a.values[strings.Trim(key, "\n\t ")] = strings.Trim(data, "\n\t ")
}

// Get is a safe function that return the type and the value
func (a *Attributes) Get(key, typ string) (interface{}, error) {
	a.ensureMap()

	switch strings.ToLower(typ) {
	case "int", "int64":
		i64, err := strconv.ParseInt(a.values[key], 10, 0)
		if err != nil {
			return nil, fmt.Errorf("key %q is not an integer: %w", key, err)
		}
		return i64, nil
	case "str", "string":
		return a.values[key], nil
	case "bool", "boolean":
		b, err := strconv.ParseBool(a.values[key])
		if err != nil {
			return nil, fmt.Errorf("key %q is not a boolean: %w", key, err)
		}
		return b, nil
	case "strarray":
		data := strings.Split(a.values[key], "|")
		return data, nil
	default:
		return nil, fmt.Errorf("invalid type %q", key)
	}
}

// Str return a key in string format
func (a *Attributes) Str(key string) (string, error) {
	s, err := a.Get(key, "str")
	if err != nil {
		return "", err
	}

	return s.(string), nil
}

// StrArray try to convert | (pipe) seperated string to array
func (a *Attributes) StrArray(key string) ([]string, error) {
	s, err := a.Get(key, "strarray")
	if err != nil {
		return nil, err
	}

	return s.([]string), nil
}

// Int return the key in the int type
func (a *Attributes) Int(key string) (int64, error) {
	i, err := a.Get(key, "int")
	if err != nil {
		return 0, err
	}

	return i.(int64), nil
}

// Bool return the boolean value of the key
func (a *Attributes) Bool(key string) (bool, error) {
	b, err := a.Get(key, "bool")
	if err != nil {
		return false, err
	}

	return b.(bool), nil
}

// All is a function to get all data, for loops mostly
func (a *Attributes) All() map[string]string {
	a.ensureMap()

	return a.values
}

// Card is a type for a single card in deck
type Card struct {
	Name  string
	Index int
	Attributes
}

// Deck is collection of cards
type Deck struct {
	Cards []*Card `yaml:"-"`

	Name        string
	Copyright   string `yaml:"copyright,omitempty"`
	Version     string
	Author      string `yaml:"author,omitempty"`
	Description string

	Width  int
	Height int

	HasCSSTemplate  bool `yaml:"-"`
	HasCardTemplate bool `yaml:"-"`
}

// LoadCSV is for loading the deck data from a CSV file
func (d *Deck) LoadCSV(r io.Reader) error {
	cr := csv.NewReader(r)
	header, err := cr.Read()
	if err != nil {
		return fmt.Errorf("unexpected error during reading the header from csv: %w", err)
	}

	if len(header) == 0 {
		return errors.New("no column")
	}

	for {
		data, err := cr.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("unexpected error: %w", err)
		}

		if len(header) != len(data) {
			return fmt.Errorf("expected %d column, found %d", len(header), len(data))
		}

		attr := Attributes{}
		var (
			name  string
			count int = 1
		)
		for i := range header {
			attr.Append(header[i], data[i])
			switch header[i] {
			case "_name":
				name = data[i]
			case "_count":
				f, err := strconv.ParseInt(data[i], 10, 0)
				if err != nil {
					return fmt.Errorf("the %q should be an integer: %w", header[i], err)
				}

				if f < 0 {
					return fmt.Errorf("the %q should be >= zero but is %d", header[i], f)
				}

				count = int(f)
			}
		}
		for i := 0; i < count; i++ {
			d.Cards = append(d.Cards, &Card{
				Name:       name,
				Index:      i,
				Attributes: cloneAttributes(attr),
			})
		}
	}

	return nil
}

// CSSClass return the CSS class for the current deck
func (d *Deck) CSSClass() string {
	return fmt.Sprintf("deck-%d-%d", d.Width, d.Height)
}

// Pages split the cards in A4 printable pages
func (d Deck) Pages() []Page {
	// TODO: determine the correct PerPage based on the size
	pp := 6
	var pages []Page

	for i, current, end := 1, 0, pp; ; i, current, end = i+1, current+pp, end+pp {
		pages = append(pages, Page{
			d:          &d,
			PerPage:    pp,
			PageNumber: i,
		})
		if end >= len(d.Cards) {
			break
		}
	}

	return pages
}

// CardsCount return number of cards inside the deck
func (d Deck) CardsCount() int {
	return len(d.Cards)
}

// PagesCount return number of pages inside the deck
func (d Deck) PagesCount() int {
	return len(d.Pages())
}

// NewDeck creates a new deck
func NewDeck(name string, width, height int) *Deck {
	return &Deck{
		Name:   name,
		Width:  width,
		Height: height,
	}
}

// Page is one page in a deck, it is printable page.
type Page struct {
	d          *Deck
	PageNumber int // From 1
	PerPage    int
}

// CSSClass is the css class of the page
func (p *Page) CSSClass() string {
	return fmt.Sprintf("page-%d-%d", p.d.Width, p.d.Height)
}

// Cards return cards in this page
func (p *Page) Cards() []*Card {
	start := (p.PageNumber - 1) * p.PerPage
	end := start + p.PerPage

	if start >= len(p.d.Cards) {
		return nil
	}

	if end >= len(p.d.Cards) {
		return p.d.Cards[start:]
	}

	return p.d.Cards[start:end]
}

// Total number of pages
func (p *Page) Total() int {
	if p.PerPage == 0 {
		return 0
	}
	count, zero := len(p.d.Cards)/p.PerPage, len(p.d.Cards)%p.PerPage
	if zero != 0 {
		count++
	}

	return count
}

// Height Card height
func (p *Page) Height() int {
	return p.d.Height
}

// Width Card Width
func (p *Page) Width() int {
	return p.d.Width
}

// Name of the deck
func (p *Page) Name() string {
	return p.d.Name
}

// Description of the deck
func (p *Page) Description() string {
	return p.d.Description
}

// Version of the deck
func (p *Page) Version() string {
	return p.d.Version
}

// Copyright of the deck
func (p *Page) Copyright() string {
	return p.d.Copyright
}

// Author of the deck
func (p *Page) Author() string {
	return p.d.Author
}

// HasCardTemplate return true if the user template has card section
func (p *Page) HasCardTemplate() bool {
	return p.d.HasCardTemplate
}

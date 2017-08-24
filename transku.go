package transku

import (
	"github.com/WedgeNix/transKU/dictionary"
	"github.com/WedgeNix/transKU/intlprods"
	"github.com/WedgeNix/util"

	"github.com/WedgeNix/chapi"
	"github.com/WedgeNix/gosetta"
	"golang.org/x/text/language"
)

// Type holds transKU controller data.
type Type struct {
	ca    *chapi.CaObj
	prods []chapi.Product
	rose  *gosetta.Rose
}

// New creates a new instance for translating English ChannelAdvisor data.
func New() (*Type, error) {
	ca, err := chapi.New()
	if err != nil {
		return nil, err
	}
	rose, err := gosetta.New(language.English)
	if err != nil {
		return nil, err
	}
	return &Type{ca, nil, rose}, nil
}

// ReadChannelAdvisor reads ChannelAdvisor product information in for parsing.
func (t *Type) ReadChannelAdvisor() error {
	prods, err := t.ca.GetCAData()
	if err != nil {
		return err
	}
	t.prods = prods
	return nil
}

type region struct {
	id    int
	label string
}

var regions = map[language.Tag]region{
	language.Chinese: region{12016078, "Amazon Seller Central - CN"},
	language.French:  region{12015122, "Amazon Seller Central - FR"},
	language.German:  region{12014987, "Amazon Seller Central - DE"},
}

// Translate translates ChannelAdvisor data from English to another language.
func (t Type) Translate(dst language.Tag) (intlprods.Type, error) {
	//
	//
	//
	t.rose.Destination(dst)

	// t.ca.Parent(true)

	dict := dictionary.New()

	done := util.NewLoader("Adding words/phrases to dictionary")

	// add product properties in parts to the dictionary
	dict.GoAdd(t.prods)
	done <- true

	done = util.NewLoader("Translating words/phrases in dictionary")

	// sets all entries in the dictionary to their respective translations
	dict.GoFillAll(t.rose.MustTranslate)
	done <- true

	done = util.NewLoader("Translating products using dictionary")

	// translates any matching word/phrases from products in dictionary
	newProds := dict.GoTransAll(t.prods)
	done <- true

	reg := regions[dst]

	return intlprods.New(newProds, reg.id, reg.label), nil
}

// WriteChannelAdvisor writes to a ChannelAdvisor region database.
func (t Type) WriteChannelAdvisor(ip intlprods.Type) error {
	done := util.NewLoader("Writing binary CSV to ChannelAdvisor")
	defer func() { done <- true }()
	return t.ca.SendBinaryCSV(ip.GetCSVLayout())
}

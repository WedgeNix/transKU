package transku

import (
	"github.com/WedgeNix/transKU/dictionary"
	"github.com/WedgeNix/transKU/intlprods"

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
func (t Type) Translate(dst language.Tag) (*intlprods.Type, error) {
	//
	//
	//

	// t.ca.Parent(true)

	dict := new(dictionary.Type)

	// add product properties in parts to the dictionary
	dict.GoAdd(t.prods)

	// sets all entries in the dictionary to their respective translations
	dict.GoFillAll(t.rose.MustTranslate)

	newProds := dict.GoTransAll(t.prods)

	reg := regions[dst]

	return intlprods.New(newProds, reg.id, reg.label), nil
}

// WriteChannelAdvisor writes to a ChannelAdvisor region database.
func (t Type) WriteChannelAdvisor(ip intlprods.Type) error {
	return t.ca.SendBinaryCSV(ip.GetCSVLayout())
}

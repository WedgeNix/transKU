package transku

import (
	"github.com/WedgeNix/transKU/dictionary"

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

// RegionData holds translate product data and the region number.
type RegionData struct {
	prods  []chapi.Product
	region int
}

var regionLookup = map[language.Tag]int{
	language.Chinese: 0,
	language.French:  0,
	language.German:  0,
}

// Translate translates ChannelAdvisor data from English to another language.
func (t Type) Translate(dst language.Tag) (RegionData, error) {
	// t.ca.Parent(true)

	data := RegionData{region: regionLookup[dst]}

	dict := new(dictionary.Type)
	dict.GoAddAttributes(t.prods)
	dict.GoAddDescriptions(t.prods)
	dict.GoAddTitles(t.prods)

	dict.GoSetAll(t.rose.MustTranslate)

	// desc := prod.Description
	// html := regex.HTML.FindAllString(desc, -1)
	// cleanDesc := htmlExpression.ReplaceAllString(desc, "<>")
	// phrases := phraseExpression.FindAllString(cleanDesc, -1)

	// cleanTrans := regex.Phrase.ReplaceAllStringFunc(cleanDesc, func(phrase string) string {
	// 	return dict[phrase]
	// })
	// util.Log(cleanTrans)

	// util.Log("this is the fully tagged, translated description")
	// fullTrans := htmlExpression.ReplaceAllStringFunc(cleanTrans, func(s string) string {
	// 	tag := html[0]  // head
	// 	html = html[1:] // tail; shift
	// 	return tag
	// })
	// util.Log(fullTrans)

	// charCnt := 0
	// hit := 0
	// util.Log("len(map)=", len(dict), " [vs] hit=", hit)
	// dollarsPerChar := 0.00002
	// util.Log("Translation cost :: ", currency.USD.Amount(dollarsPerChar*float64(charCnt)))

	return data, nil
}

// WriteChannelAdvisor writes to a ChannelAdvisor region database.
func (t Type) WriteChannelAdvisor(data RegionData) error {
	return t.ca.CSVify(data.prods, data.region) // should have a proper region read
}

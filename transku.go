package transku

import (
	"encoding/gob"
	"fmt"
	"os"

	"github.com/WedgeNix/awsapi"
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
	aws   *awsapi.Controller
	rose  *gosetta.Rose
}

type region struct {
	ID    int
	Label string
	Dict  string
}

var regions = map[language.Tag]region{
	language.Chinese: region{12016078, "Amazon Seller Central - CN", "cn.gob"},
	language.French:  region{12015122, "Amazon Seller Central - FR", "fr.gob"},
	language.German:  region{12014987, "Amazon Seller Central - DE", "de.gob"},
}

// InitChapi creates a new instance for translating English ChannelAdvisor data.
func InitChapi() (*Type, error) {

	done := util.NewLoader("Initializing transKU")
	ca, err := chapi.New()
	if err != nil {
		return nil, err
	}
	done <- true

	return &Type{ca: ca}, nil
}

func (t *Type) InitAwsapi() error {

	aws, err := awsapi.New()
	if err != nil {
		return err
	}
	t.aws = aws

	return nil
}

func (t *Type) InitGosetta() error {

	rose, err := gosetta.New(language.English)
	if err != nil {
		return err
	}
	t.rose = rose

	return nil
}

// ReadChannelAdvisor reads ChannelAdvisor product information in for parsing.
func (t *Type) ReadChannelAdvisor() error {

	fnm := "prods.gob"

	f, err := os.Open(fnm)
	if err == nil {

		done := util.NewLoader("Decoding product data from '" + fnm + "'")
		f, err := os.Open(fnm)
		if err != nil {
			return err
		}
		d := gob.NewDecoder(f)
		err = d.Decode(&t.prods)
		if err != nil {
			return err
		}
		done <- true

		// for _, t := range t.prods {
		// 	if t.Sku != "FM4048-RDBK-One Size" {
		// 		continue
		// 	}
		// 	panic("found")
		// }
		// panic("not found")

	} else {

		done := util.NewLoader("Reading product data from ChannelAdvisor")
		prods, err := t.ca.GetCAData()
		if err != nil {
			return err
		}
		t.prods = prods
		done <- true

		f, err = os.Create(fnm)
		if err != nil {
			return err
		}

		done = util.NewLoader("Encoding product data to '" + fnm + "'")
		e := gob.NewEncoder(f)
		err = e.Encode(prods)
		if err != nil {
			return err
		}
		done <- true
	}

	return nil
}

// ReadDictFromAWS reads a cached dictionary from AWS.
// func (t *Type) ReadDictFromAWS(dst language.Tag) error {

// 	fnm := regions[dst].Dict

// 	if _, err := os.Stat(fnm); os.IsExist(err) {

// 		done := util.NewLoader("Decoding dictionary from '" + fnm + "'")
// 		f, err := os.Open(regions[dst].Dict)
// 		if err != nil {
// 			return err
// 		}
// 		d := gob.NewDecoder(f)
// 		err = d.Decode(&t.dict)
// 		if err != nil {
// 			return err
// 		}
// 		done <- true

// 	} else {

// 		f, err := os.Create(fnm)
// 		if err != nil {
// 			return err
// 		}

// 		done := util.NewLoader("Reading dictionary from AWS")
// 		err = t.aws.Read("transku/"+fnm, encdec.Gob, &t.dict)
// 		if err != nil {
// 			return err
// 		}
// 		done <- true

// 		done = util.NewLoader("Encoding dictionary to '" + fnm + "'")
// 		e := gob.NewEncoder(f)
// 		err = e.Encode(t.dict)
// 		if err != nil {
// 			return err
// 		}
// 		done <- true
// 	}

// 	return nil
// }

// WriteDictToAWS writes a cached dictionary to AWS.
// func (t *Type) WriteDictToAWS(dst language.Tag) error {

// 	done := util.NewLoader("Writing dictionary to AWS")
// 	err := t.aws.Write("transku/"+regions[dst].Dict, encdec.Gob, t.dict)
// 	if err != nil {
// 		return err
// 	}
// 	done <- true

// 	return nil
// }

// CreateDict creates and translates a dictionary.
func (t Type) CreateDict(dst language.Tag) (*dictionary.Type, error) {

	fnm := regions[dst].Dict
	dict := dictionary.Lookup{}

	f, err := os.Open(fnm)
	if err == nil {

		done := util.NewLoader("Decoding dictionary from '" + fnm + "'")
		d := gob.NewDecoder(f)
		err = d.Decode(&dict)
		if err != nil {
			return nil, err
		}
		f.Close()
		done <- true

	} else {

		done := util.NewLoader("Reading dictionary from AWS")
		err = t.aws.Read("transku/"+fnm, &dict)
		if err != nil {
			return nil, err
		}
		done <- true
	}

	done := util.NewLoader("Initializing dictionary")
	d := dictionary.New(dict)
	done <- true

	done = util.NewLoader("Adding words/phrases to dictionary")
	d.GoAdd(t.prods)
	done <- true

	fmt.Println(d.GetPrice())

	done = util.NewLoader("Translating words in dictionary")
	t.rose.Destination(dst)
	d.GoFillAll(t.rose.MustTranslate)
	done <- true

	done = util.NewLoader("Encoding dictionary to '" + fnm + "'")
	f, err = os.Create(fnm)
	if err != nil {
		return nil, err
	}
	e := gob.NewEncoder(f)
	err = e.Encode(dict)
	if err != nil {
		return nil, err
	}
	done <- true

	done = util.NewLoader("Writing dictionary to AWS")
	err = t.aws.Write("transku/"+fnm, dict)
	if err != nil {
		return nil, err
	}
	done <- true

	return d, nil
}

// ApplyDict translates ChannelAdvisor data from English to another language.
func (t Type) ApplyDict(dict *dictionary.Type, dst language.Tag) intlprods.Type {

	done := util.NewLoader("Translating products using dictionary")
	newProds := dict.GoTransAll(t.prods)
	done <- true

	reg := regions[dst]

	done = util.NewLoader("Converting translated to international format")
	ip := intlprods.New(newProds, reg.ID, reg.Label)
	done <- true

	return ip
}

// WriteChannelAdvisor writes to a ChannelAdvisor region database.
func (t Type) WriteChannelAdvisor(ip intlprods.Type) error {

	done := util.NewLoader("Writing binary CSV to ChannelAdvisor")
	err := t.ca.SendBinaryCSV(ip.GetCSVLayout())
	if err != nil {
		return err
	}
	done <- true

	return nil
}

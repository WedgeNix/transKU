package transku

import (
	"encoding/gob"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/WedgeNix/awsapi"
	"github.com/WedgeNix/util"

	"github.com/WedgeNix/chapi"
	"github.com/WedgeNix/gosetta"
	"golang.org/x/text/language"
)

// InitChapi creates a new instance for translating English ChannelAdvisor data.
func InitChapi(start time.Time) (*TransKU, error) {
	done := util.NewLoader("Initializing transKU")
	ca, err := chapi.New()
	if err != nil {
		return nil, err
	}
	done <- true

	return &TransKU{ca: ca, createDate: start}, nil
}

// InitAwsapi initializes Awsapi right before point needed.
func (t *TransKU) InitAwsapi() error {
	aws, err := awsapi.New()
	if err != nil {
		return err
	}
	t.aws = aws

	return nil
}

// InitGosetta initializes Gosetta right before point needed.
func (t *TransKU) InitGosetta() error {
	rose, err := gosetta.New(language.English)
	if err != nil {
		return err
	}
	t.rose = rose

	return nil
}

// ReadChannelAdvisor reads ChannelAdvisor product information in for parsing.
func (t *TransKU) ReadChannelAdvisor() error {
	fnm := "prods.gob"

	f, err := os.Open(fnm)
	if err == nil {
		done := util.NewLoader("Decoding product data from '" + fnm + "'")
		d := gob.NewDecoder(f)
		err = d.Decode(&t.prods)
		if err != nil {
			return err
		}
		done <- true
	} else {
		done := util.NewLoader("Reading product data from ChannelAdvisor")
		prods, err := t.ca.GetCAData(t.createDate)
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

// CreateDict creates and translates a Dictionary.
func (t TransKU) CreateDict(r Region) (*Dictionary, error) {
	fnm := strings.ToLower(r.ChannelTag + ".gob")
	dict := lookup{}

	f, err := os.Open(fnm)
	if err == nil {
		done := util.NewLoader("Decoding Dictionary from '" + fnm + "'")
		d := gob.NewDecoder(f)
		err = d.Decode(&dict)
		if err != nil {
			return nil, err
		}
		f.Close()
		done <- true
	} else {
		done := util.NewLoader("Reading Dictionary from AWS")
		err = t.aws.Read("transku/"+fnm, &dict)
		if err != nil {
			return nil, err
		}
		done <- true
	}

	done := util.NewLoader("Initializing Dictionary")
	d := newDictionary(dict)
	done <- true

	done = util.NewLoader("Adding words/phrases to Dictionary")
	d.GoAdd(t.prods)
	done <- true

	fmt.Println(d.GetPrice())

	done = util.NewLoader("Translating words in Dictionary")
	tag, err := language.Parse(r.BCP47)
	if err != nil {
		return nil, err
	}
	t.rose.Destination(tag)
	d.GoFillAll(t.rose.MustTranslate)
	done <- true

	done = util.NewLoader("Encoding Dictionary to '" + fnm + "'")
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

	done = util.NewLoader("Writing Dictionary to AWS")
	err = t.aws.Write("transku/"+fnm, dict)
	if err != nil {
		return nil, err
	}
	done <- true

	return d, nil
}

// ApplyDict translates ChannelAdvisor data from English to another language.
func (t TransKU) ApplyDict(dict *Dictionary, r Region) IntlProds {
	done := util.NewLoader("Translating products using Dictionary")
	newProds := dict.GoTransAll(t.prods)
	done <- true

	done = util.NewLoader("Converting translated to international format")
	ip := newIntlProds(newProds, r.ProfileID, `Amazon Seller Central - `+strings.ToUpper(r.ChannelTag))
	done <- true

	return ip
}

// WriteChannelAdvisor writes to a ChannelAdvisor region database.
func (t TransKU) WriteChannelAdvisor(ip IntlProds) error {
	done := util.NewLoader("Writing binary CSV to ChannelAdvisor")
	err := t.ca.SendBinaryCSV(ip.GetCSVLayout())
	if err != nil {
		return err
	}
	done <- true

	return nil
}

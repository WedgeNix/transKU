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
	util.Log("Initializing transKU" + "...")
	ca, err := chapi.New()
	if err != nil {
		return nil, err
	}
	util.Log("Initializing transKU" + " !")

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
		util.Log("Decoding product data from '" + fnm + "'" + "...")
		d := gob.NewDecoder(f)
		err = d.Decode(&t.prods)
		if err != nil {
			return err
		}
		util.Log("Decoding product data from '" + fnm + "'" + " !")
	} else {
		util.Log("Reading product data from ChannelAdvisor" + "...")
		prods, err := t.ca.GetCAData(t.createDate)
		if err != nil {
			return err
		}
		t.prods = prods
		util.Log("Reading product data from ChannelAdvisor" + " !")

		f, err = os.Create(fnm)
		if err != nil {
			return err
		}

		util.Log("Encoding product data to '" + fnm + "'" + "...")
		e := gob.NewEncoder(f)
		err = e.Encode(prods)
		if err != nil {
			return err
		}
		util.Log("Encoding product data to '" + fnm + "'" + " !")
	}
	f.Close()

	return nil
}

// CreateDict creates and translates a Dictionary.
func (t TransKU) CreateDict(r Region) (*Dictionary, error) {
	fnm := strings.ToLower(r.ChannelTag + ".gob")
	dict := lookup{}

	// f, err := os.Open(fnm)
	// if err == nil {
	// 	util.Log("Decoding Dictionary from '" + fnm + "'" + "...")
	// 	d := gob.NewDecoder(f)
	// 	err = d.Decode(&dict)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	f.Close()
	// 	util.Log("Decoding Dictionary from '" + fnm + "'" + " !")
	// } else {

	util.Log("Reading Dictionary from AWS" + "...")
	err := t.aws.Read("transku/"+fnm, &dict)
	if err != nil {
		return nil, err
	}
	util.Log("Reading Dictionary from AWS" + " !")
	// }

	// fmt.Println("[check your memory usage] aws.Read")
	// time.Sleep(10 * time.Second)

	util.Log("Initializing Dictionary" + "...")
	tag, err := language.Parse(r.BCP47)
	if err != nil {
		return nil, err
	}
	d := newDictionary(tag, dict)
	util.Log("Initializing Dictionary" + " !")

	// fmt.Println("[check your memory usage] newDictionary")
	// time.Sleep(10 * time.Second)

	util.Log("Adding words/phrases to Dictionary" + "...")
	d.GoAdd(t.prods)
	util.Log("Adding words/phrases to Dictionary" + " !")

	// fmt.Println("[check your memory usage] GoAdd")
	// time.Sleep(10 * time.Second)

	fmt.Println(d.GetPrice())

	// fmt.Println("[check your memory usage] GetPrice")
	// time.Sleep(10 * time.Second)

	util.Log("Translating words in Dictionary" + "...")
	t.rose.Destination(tag)
	d.GoFillAll(t.rose.MustTranslate)
	util.Log("Translating words in Dictionary" + " !")

	// fmt.Println("[check your memory usage] GoFillAll")
	// time.Sleep(10 * time.Second)

	// util.Log("Encoding Dictionary to '" + fnm + "'" + "...")
	// f, err := os.Create(fnm)
	// if err != nil {
	// 	return nil, err
	// }
	// e := gob.NewEncoder(f)
	// err = e.Encode(dict)
	// if err != nil {
	// 	return nil, err
	// }
	// util.Log("Encoding Dictionary to '" + fnm + "'" + " !")

	// fmt.Println("[check your memory usage] Encode dict")
	// time.Sleep(10 * time.Second)

	util.Log("Writing Dictionary to AWS" + "...")
	err = t.aws.Write("transku/"+fnm, dict)
	if err != nil {
		return nil, err
	}
	util.Log("Writing Dictionary to AWS" + " !")

	// fmt.Println("[check your memory usage] Write aws")
	// time.Sleep(240 * time.Second)

	return d, nil
}

// ApplyDict translates ChannelAdvisor data from English to another language.
func (t TransKU) ApplyDict(dict *Dictionary, r Region) (IntlProds, error) {
	util.Log("Translating products using Dictionary" + "...")
	newProds := dict.GoTransAll(t.prods)
	util.Log("Translating products using Dictionary" + " !")

	caTag := strings.ToUpper(r.ChannelTag)

	util.Log("Converting translated to international format [" + caTag + "]" + "...")
	lang, err := language.Parse(r.BCP47)
	if err != nil {
		return IntlProds{}, err
	}
	ip := newIntlProds(newProds, r.ProfileID, `Amazon Seller Central - `+caTag, lang)
	util.Log("Converting translated to international format [" + caTag + "]" + " !")

	return ip, nil
}

// WriteChannelAdvisor writes to a ChannelAdvisor region database.
func (t TransKU) WriteChannelAdvisor(ip IntlProds) error {
	util.Log("Writing binary CSV to ChannelAdvisor" + "...")
	err := t.ca.SendBinaryCSV(ip.GetCSVLayout())
	if err != nil {
		return err
	}
	util.Log("Writing binary CSV to ChannelAdvisor" + " !")

	return nil
}

package transku

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"golang.org/x/text/language"

	"github.com/WedgeNix/chapi"
	"github.com/WedgeNix/util"
)

// IntlProds holds info for international products.
type IntlProds struct {
	pres      []PreCSV
	profileID int
}

type attrkv struct {
	name  string
	value string
}

// PreCSV is the look of what will be sent over to ChannelAdvisor regions.
type PreCSV struct {
	InventoryNumber,
	AuctionTitle,
	Brand,
	BuyItNowPrice,
	Classification,
	Description,
	Labels,
	PictureURLs,
	RelationshipName,
	RetailPrice,
	SellerCost,
	UPC,
	VariationParentSKU,
	Weight string

	attributes []attrkv
}

// New creates proper international products.
func newIntlProds(prods []chapi.Product, profileID int, label string, lang language.Tag) (IntlProds, error) {
	ip := IntlProds{profileID: profileID}

	ps := parentSKUs{}
	for i, prod := range prods {

		if lang == language.English {
			prod.Weight /= 2.20462
		}

		if len(prod.Sku) == 0 {
			return ip, errors.New("empty inventory number (sku) @ index-" + strconv.Itoa(i))
		}

		p := PreCSV{
			InventoryNumber:    prod.Sku,
			AuctionTitle:       prod.Title,
			Brand:              prod.Brand,
			BuyItNowPrice:      strconv.FormatFloat(prod.BuyItNowPrice, 'f', 2, 64),
			Classification:     prod.Classification,
			Description:        prod.Description,
			Labels:             label,
			RelationshipName:   prod.RelationshipName,
			RetailPrice:        strconv.FormatFloat(prod.RetailPrice, 'f', 2, 64),
			SellerCost:         strconv.FormatFloat(prod.Cost, 'f', 2, 64),
			UPC:                prod.UPC,
			VariationParentSKU: ps.getVariationParentSKU(prod, prods),
			Weight:             strconv.FormatFloat(prod.Weight, 'f', 2, 64),
		}

		if len(p.AuctionTitle) == 0 {
			return ip, errors.New("empty 'AuctionTitle'")
		}
		if len(p.Brand) == 0 {
			return ip, errors.New("empty 'Brand'")
		}
		if len(p.BuyItNowPrice) == 0 {
			return ip, errors.New("empty 'BuyItNowPrice'")
		}
		if len(p.Classification) == 0 {
			return ip, errors.New("empty 'Classification'")
		}
		if len(p.Description) == 0 {
			return ip, errors.New("empty 'Description'")
		}
		if len(p.Labels) == 0 {
			return ip, errors.New("empty 'Labels'")
		}
		if len(p.RelationshipName) == 0 {
			return ip, errors.New("empty 'RelationshipName'")
		}
		if len(p.RetailPrice) == 0 {
			return ip, errors.New("empty 'RetailPrice'")
		}
		if len(p.SellerCost) == 0 {
			return ip, errors.New("empty 'SellerCost'")
		}
		if len(p.UPC) == 0 {
			return ip, errors.New("empty 'UPC'")
		}
		if len(p.Weight) == 0 {
			return ip, errors.New("empty 'Weight'")
		}
		if len(p.VariationParentSKU) == 0 {
			log.Println("empty 'VariationParentSKU' for " + p.InventoryNumber)
		}

		urls := []string{}
		for _, img := range prod.Images {
			urls = append(urls, img.URL)
		}
		if len(urls) == 0 {
			return ip, errors.New("no 'PictureURLs' found")
		}
		p.PictureURLs = `"` + strings.Join(urls, ",") + `"`

		for _, attr := range prod.Attributes {
			_, exists := FilterAttr[attr.Name]
			if !exists {
				continue
			}
			p.attributes = append(p.attributes, attrkv{attr.Name, attr.Value})
		}

		ip.pres = append(ip.pres, p)
		util.Log(i+1, "/", len(prods))
	}

	return ip, nil
}

// GetCSVLayout formats the data to suit a CSV layout and gives the profileID.
func (ip IntlProds) GetCSVLayout() ([][]string, int) {
	layout := make([][]string, len(ip.pres)+1)
	layout[0] = []string{
		`Inventory Number`,
		`Auction Title`,
		`Brand`,
		`Buy It Now Price`,
		`Classification`,
		`Description`,
		`Labels`,
		`Picture URLs`,
		`Relationship Name`,
		`Retail Price`,
		`Seller Cost`,
		`UPC`,
		`Variation Parent SKU`,
		`Weight`,
	}
	for i := range make([]int, len(FilterAttr)) {
		n := strconv.Itoa(i + 1)
		layout[0] = append(layout[0], []string{
			`Attribute` + n + `Name`,
			`Attribute` + n + `Value`,
		}...)
	}

	work := sync.WaitGroup{}
	work.Add(len(ip.pres))

	for i, pre := range ip.pres {
		go func(i int, pre PreCSV) {
			defer work.Done()
			layout[i] = []string{
				pre.InventoryNumber,
				pre.AuctionTitle,
				pre.Brand,
				pre.BuyItNowPrice,
				pre.Classification,
				pre.Description,
				pre.Labels,
				pre.PictureURLs,
				pre.RelationshipName,
				pre.RetailPrice,
				pre.SellerCost,
				pre.UPC,
				pre.VariationParentSKU,
				pre.Weight,
			}
			for _, attr := range pre.attributes {
				layout[i] = append(layout[i], []string{attr.name, attr.value}...)
			}
		}(i+1, pre)
	}
	work.Wait()

	return layout, ip.profileID
}

type parentSKUs map[int]string

func (ps parentSKUs) getVariationParentSKU(prod chapi.Product, prods []chapi.Product) string {
	// t := time.Now()

	if prod.IsParent {
		// util.Log("[getVariationParentSKU] IsParent")
		return "Parent"
	}
	sku, exists := ps[prod.ParentProductID]
	if exists {
		// util.Log("[getVariationParentSKU] CacheHit")
		return sku
	}

	skuc := make(chan string)

	var misswg sync.WaitGroup
	var miss uint64

	for _, p := range prods {
		misswg.Add(1)
		p := p
		go func() {
			if prod.ParentProductID != p.ID {
				atomic.AddUint64(&miss, 1)
				misswg.Done()
				return
			}
			misswg.Done()
			skuc <- p.Sku
		}()
	}

	misswg.Wait()
	if int(atomic.LoadUint64(&miss)) == len(prods) {
		sku = ""
		ps[prod.ParentProductID] = sku
		// util.Log("[getVariationParentSKU] [miss] ", time.Since(t).Nanoseconds(), "ns")
		return sku
	}

	sku = <-skuc
	ps[prod.ParentProductID] = sku

	// util.Log("[getVariationParentSKU] ", time.Since(t).Nanoseconds(), "ns")
	return sku
}

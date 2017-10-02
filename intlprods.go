package transku

import (
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

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
	ASIN,
	AuctionTitle,
	Brand,
	BuyItNowPrice,
	Classification,
	Description,
	InventoryNumber,
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
func newIntlProds(prods []chapi.Product, profileID int, label string) IntlProds {
	ip := IntlProds{profileID: profileID}

	ps := parentSKUs{}
	for i, prod := range prods {

		p := PreCSV{
			ASIN:               prod.ASIN,
			AuctionTitle:       prod.Title,
			Brand:              prod.Brand,
			BuyItNowPrice:      strconv.FormatFloat(prod.BuyItNowPrice, 'f', 2, 64),
			Classification:     prod.Classification,
			Description:        prod.Description,
			InventoryNumber:    prod.Sku,
			Labels:             label,
			RelationshipName:   prod.RelationshipName,
			RetailPrice:        strconv.FormatFloat(prod.RetailPrice, 'f', 2, 64),
			SellerCost:         strconv.FormatFloat(prod.Cost, 'f', 2, 64),
			UPC:                prod.UPC,
			VariationParentSKU: ps.getVariationParentSKU(prod, prods),
			Weight:             strconv.FormatFloat(prod.Weight, 'f', 2, 64),
		}

		urls := []string{}
		for _, img := range prod.Images {
			urls = append(urls, img.URL)
		}
		p.PictureURLs = `"` + strings.Join(urls, ",") + `"`

		for _, attr := range prod.Attributes {
			_, exists := FilterAttr[attr.Name]
			if !exists {
				continue
			}
			if attr.Name == `AMZTitle` {
				println(`intlprods.go / newIntlProds /`, attr.Value)
			}
			p.attributes = append(p.attributes, attrkv{attr.Name, attr.Value})
		}

		ip.pres = append(ip.pres, p)
		util.Log(i+1, "/", len(prods))
	}

	return ip
}

// GetCSVLayout formats the data to suit a CSV layout and gives the profileID.
func (ip IntlProds) GetCSVLayout() ([][]string, int) {
	layout := make([][]string, len(ip.pres)+1)
	layout[0] = []string{
		`ASIN`,
		`Auction Title`,
		`Brand`,
		`Buy It Now Price`,
		`Classification`,
		`Description`,
		`Inventory Number`,
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
				pre.ASIN,
				pre.AuctionTitle,
				pre.Brand,
				pre.BuyItNowPrice,
				pre.Classification,
				pre.Description,
				pre.InventoryNumber,
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

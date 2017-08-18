package intlprods

import (
	"strconv"
	"strings"

	"golang.org/x/text/language"

	"github.com/WedgeNix/chapi"
)

type region struct {
	id    int
	label string
}

var lookup = map[language.Tag]region{
	language.Chinese: region{0, "Amazon Seller Central - CN"},
	language.French:  region{0, "Amazon Seller Central - FR"},
	language.German:  region{0, "Amazon Seller Central - DE"},
}

type Type struct {
	prods         []chapi.Product
	pres          []PreCSV
	parentSkuDict map[int]string
	ID            int
}

type PreCSV struct {
	AuctionTitle        string
	InventoryNumber     string
	ItemCreateDate      string
	Weight              string
	UPC                 string
	ASIN                string
	Description         string
	Brand               string
	SellerCost          string
	BuyItNowPrice       string
	RetailPrice         string
	PictureURLs         string
	ReceivedInInventory string
	RelationshipName    string
	VariationParentSKU  string
	Labels              string
	Classification      string
	AMZCategory         string
	AMZColorMap         string
	AMZItemType         string
	AMZProductIDType    string
	AMZClothingType     string
	AMZColor            string
	AMZDepartment       string
	AMZDescription      string
	AMZSize             string
	AMZTitle            string
	ApparelClosureType  string
	ArmLength           string
	BandMaterial        string
	BottomStyle         string
	BottomsSizeMens     string
	BottomsSizeWomens   string
	BoysClothingType    string
	ClothingType        string
	FeatureBullet1      string
	FeatureBullet2      string
	FeatureBullet3      string
	FeatureBullet4      string
	FeatureBullet5      string
	FeatureBullet6      string
}

func New(prods []chapi.Product, dst language.Tag) *Type {
	t := new(Type)
	t.prods = prods

	for _, prod := range prods {
		p := PreCSV{
			AuctionTitle:        prod.Title,
			InventoryNumber:     prod.Sku,
			ItemCreateDate:      prod.CreateDateUtc.String(),
			Weight:              strconv.FormatFloat(prod.Weight, 'f', 2, 64),
			UPC:                 prod.UPC,
			ASIN:                prod.ASIN,
			Description:         prod.Description,
			Brand:               prod.Brand,
			SellerCost:          strconv.FormatFloat(prod.Cost, 'f', 2, 64),
			BuyItNowPrice:       strconv.FormatFloat(prod.BuyItNowPrice, 'f', 2, 64),
			RetailPrice:         strconv.FormatFloat(prod.RetailPrice, 'f', 2, 64),
			ReceivedInInventory: prod.ReceivedDateUtc.String(),
			RelationshipName:    prod.RelationshipName,
			VariationParentSKU:  t.getVariationParentSKU(prod),
			Labels:              lookup[dst].label,
			Classification:      prod.Classification,
		}
		urls := []string{}
		for _, img := range prod.Images {
			urls = append(urls, img.URL)
		}
		p.PictureURLs = `"` + strings.Join(urls, ",") + `"`

		t.pres = append(t.pres, p)
	}

	return t
}

func (t Type) getVariationParentSKU(prod chapi.Product) string {
	sku, exists := t.parentSkuDict[prod.ParentProductID]
	if exists {
		return sku
	}

	skuc := make(chan string)

	for _, p := range t.prods {
		go func(p chapi.Product) {
			if prod.ParentProductID != p.ID {
				return
			}

			skuc <- p.Sku
			close(skuc)
		}(p)
	}

	return <-skuc
}

func (t Type) csv() {

}

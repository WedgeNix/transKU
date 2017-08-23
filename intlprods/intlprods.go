package intlprods

import (
	"strconv"
	"strings"
	"sync"

	"github.com/WedgeNix/chapi"
)

// Type holds info for international products.
type Type struct {
	pres   []PreCSV
	region int
}

// PreCSV is the look of what will be sent over to ChannelAdvisor regions.
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

// New creates proper international products.
func New(prods []chapi.Product, region int, label string) *Type {
	t := new(Type)
	t.region = region

	ps := parentSKUs{}

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
			VariationParentSKU:  ps.getVariationParentSKU(prod, prods),
			Labels:              label,
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

// GetCSVLayout formats the data to suit a CSV layout and gives the region.
func (t Type) GetCSVLayout() ([][]string, int) {
	layout := make([][]string, len(t.pres)+1)
	layout[0] = []string{
		`Auction Title`,
		`Inventory Number`,
		`Item Create Date`,
		`Weight`,
		`UPC`,
		`ASIN`,
		`Description`,
		`Brand`,
		`Seller Cost`,
		`Buy It Now Price`,
		`Retail Price`,
		`Picture URLs`,
		`Received In Inventory`,
		`Relationship Name`,
		`Variation Parent SKU`,
		`Labels`,
		`Classification`,
		`AMZ_Category`,
		`AMZ_Color_Map`,
		`AMZ_Item_Type`,
		`AMZ_ProductIDType`,
		`AMZClothingType`,
		`AMZColor`,
		`AMZDepartment`,
		`AMZDescription`,
		`AMZSize`,
		`AMZTitle`,
		`Apparel-Closure-Type`,
		`Arm Length`,
		`Band Material`,
		`Bottom Style`,
		`Bottoms Size (Men's)`,
		`Bottoms Size (Women's)`,
		`Boy's Clothing Type`,
		`Clothing Type`,
		`FeatureBullet1`,
		`FeatureBullet2`,
		`FeatureBullet3`,
		`FeatureBullet4`,
		`FeatureBullet5`,
		`FeatureBullet6`,
	}

	work := sync.WaitGroup{}
	work.Add(len(t.pres))

	for i, pre := range t.pres {
		go func(i int, pre PreCSV) {
			defer work.Done()
			layout[i] = []string{
				pre.AuctionTitle,
				pre.InventoryNumber,
				pre.ItemCreateDate,
				pre.Weight,
				pre.UPC,
				pre.ASIN,
				pre.Description,
				pre.Brand,
				pre.SellerCost,
				pre.BuyItNowPrice,
				pre.RetailPrice,
				pre.PictureURLs,
				pre.ReceivedInInventory,
				pre.RelationshipName,
				pre.VariationParentSKU,
				pre.Labels,
				pre.Classification,
				pre.AMZCategory,
				pre.AMZColorMap,
				pre.AMZItemType,
				pre.AMZProductIDType,
				pre.AMZClothingType,
				pre.AMZColor,
				pre.AMZDepartment,
				pre.AMZDescription,
				pre.AMZSize,
				pre.AMZTitle,
				pre.ApparelClosureType,
				pre.ArmLength,
				pre.BandMaterial,
				pre.BottomStyle,
				pre.BottomsSizeMens,
				pre.BottomsSizeWomens,
				pre.BoysClothingType,
				pre.ClothingType,
				pre.FeatureBullet1,
				pre.FeatureBullet2,
				pre.FeatureBullet3,
				pre.FeatureBullet4,
				pre.FeatureBullet5,
				pre.FeatureBullet6,
			}
		}(i+1, pre)
	}
	work.Wait()

	return layout, t.region
}

type parentSKUs map[int]string

func (ps parentSKUs) getVariationParentSKU(prod chapi.Product, prods []chapi.Product) string {
	sku, exists := ps[prod.ParentProductID]
	if exists {
		return sku
	}

	skuc := make(chan string)

	go func() {
		for _, p := range prods {
			go func(p chapi.Product) {
				if prod.ParentProductID != p.ID {
					return
				}
				skuc <- p.Sku
				close(skuc)
			}(p)
		}
	}()

	return <-skuc
}

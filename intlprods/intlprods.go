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
func New(prods []chapi.Product, region int, label string) Type {
	t := Type{region: region}

	ps := parentSKUs{}

	for _, prod := range prods {
		p := PreCSV{
			AuctionTitle:        prod.Title,
			InventoryNumber:     prod.Sku,
			ItemCreateDate:      prod.CreateDateUtc.Format("01/02/2006"),
			Weight:              strconv.FormatFloat(prod.Weight, 'f', 2, 64),
			UPC:                 prod.UPC,
			ASIN:                prod.ASIN,
			Description:         prod.Description,
			Brand:               prod.Brand,
			SellerCost:          strconv.FormatFloat(prod.Cost, 'f', 2, 64),
			BuyItNowPrice:       strconv.FormatFloat(prod.BuyItNowPrice, 'f', 2, 64),
			RetailPrice:         strconv.FormatFloat(prod.RetailPrice, 'f', 2, 64),
			ReceivedInInventory: prod.ReceivedDateUtc.Format("01/02/2006"),
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

		for _, attr := range prod.Attributes {
			switch attr.Name {
			case `AMZ_Category`:
				p.AMZCategory = attr.Value
			case `AMZ_Color_Map`:
				p.AMZColorMap = attr.Value
			case `AMZ_Item_Type`:
				p.AMZItemType = attr.Value
			case `AMZ_ProductIDType`:
				p.AMZProductIDType = attr.Value
			case `AMZClothingType`:
				p.AMZClothingType = attr.Value
			case `AMZColor`:
				p.AMZColor = attr.Value
			case `AMZDepartment`:
				p.AMZDepartment = attr.Value
			case `AMZDescription`:
				p.AMZDescription = attr.Value
			case `AMZSize`:
				p.AMZSize = attr.Value
			case `AMZTitle`:
				p.AMZTitle = attr.Value
			case `Apparel-Closure-Type`:
				p.ApparelClosureType = attr.Value
			case `Arm Length`:
				p.ArmLength = attr.Value
			case `Band Material`:
				p.BandMaterial = attr.Value
			case `Bottom Style`:
				p.BottomStyle = attr.Value
			case `Bottoms Size (Men's)`:
				p.BottomsSizeMens = attr.Value
			case `Bottoms Size (Women's)`:
				p.BottomsSizeWomens = attr.Value
			case `Boy's Clothing Type`:
				p.BoysClothingType = attr.Value
			case `Clothing Type`:
				p.ClothingType = attr.Value
			case `FeatureBullet1`:
				p.FeatureBullet1 = attr.Value
			case `FeatureBullet2`:
				p.FeatureBullet2 = attr.Value
			case `FeatureBullet3`:
				p.FeatureBullet3 = attr.Value
			case `FeatureBullet4`:
				p.FeatureBullet4 = attr.Value
			case `FeatureBullet5`:
				p.FeatureBullet5 = attr.Value
			case `FeatureBullet6`:
				p.FeatureBullet6 = attr.Value
			}
		}

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
		// `Item Create Date`,
		`Weight`,
		`UPC`,
		`ASIN`,
		`Description`,
		`Brand`,
		`Seller Cost`,
		`Buy It Now Price`,
		`Retail Price`,
		`Picture URLs`,
		// `Received In Inventory`,
		`Relationship Name`,
		`Variation Parent SKU`,
		`Labels`,
		`Classification`,
		`Attribute1Name`,
		`Attribute1Value`,
		`Attribute2Name`,
		`Attribute2Value`,
		`Attribute3Name`,
		`Attribute3Value`,
		`Attribute4Name`,
		`Attribute4Value`,
		`Attribute5Name`,
		`Attribute5Value`,
		`Attribute6Name`,
		`Attribute6Value`,
		`Attribute7Name`,
		`Attribute7Value`,
		`Attribute8Name`,
		`Attribute8Value`,
		`Attribute9Name`,
		`Attribute9Value`,
		`Attribute10Name`,
		`Attribute10Value`,
		`Attribute11Name`,
		`Attribute11Value`,
		`Attribute12Name`,
		`Attribute12Value`,
		`Attribute13Name`,
		`Attribute13Value`,
		`Attribute14Name`,
		`Attribute14Value`,
		`Attribute15Name`,
		`Attribute15Value`,
		`Attribute16Name`,
		`Attribute16Value`,
		`Attribute17Name`,
		`Attribute17Value`,
		`Attribute18Name`,
		`Attribute18Value`,
		`Attribute19Name`,
		`Attribute19Value`,
		`Attribute20Name`,
		`Attribute20Value`,
		`Attribute21Name`,
		`Attribute21Value`,
		`Attribute22Name`,
		`Attribute22Value`,
		`Attribute23Name`,
		`Attribute23Value`,
		`Attribute24Name`,
		`Attribute24Value`,
	}

	work := sync.WaitGroup{}
	work.Add(len(t.pres))

	for i, pre := range t.pres {
		go func(i int, pre PreCSV) {
			defer work.Done()
			layout[i] = []string{
				pre.AuctionTitle,
				pre.InventoryNumber,
				// pre.ItemCreateDate,
				pre.Weight,
				pre.UPC,
				pre.ASIN,
				pre.Description,
				pre.Brand,
				pre.SellerCost,
				pre.BuyItNowPrice,
				pre.RetailPrice,
				pre.PictureURLs,
				// pre.ReceivedInInventory,
				pre.RelationshipName,
				pre.VariationParentSKU,
				pre.Labels,
				pre.Classification,
				`AMZ_Category`,
				pre.AMZCategory,
				`AMZ_Color_Map`,
				pre.AMZColorMap,
				`AMZ_Item_Type`,
				pre.AMZItemType,
				`AMZ_ProductIDType`,
				pre.AMZProductIDType,
				`AMZClothingType`,
				pre.AMZClothingType,
				`AMZColor`,
				pre.AMZColor,
				`AMZDepartment`,
				pre.AMZDepartment,
				`AMZDescription`,
				pre.AMZDescription,
				`AMZSize`,
				pre.AMZSize,
				`AMZTitle`,
				pre.AMZTitle,
				`Apparel-Closure-Type`,
				pre.ApparelClosureType,
				`Arm Length`,
				pre.ArmLength,
				`Band Material`,
				pre.BandMaterial,
				`Bottom Style`,
				pre.BottomStyle,
				`Bottoms Size (Men's)`,
				pre.BottomsSizeMens,
				`Bottoms Size (Women's)`,
				pre.BottomsSizeWomens,
				`Boy's Clothing Type`,
				pre.BoysClothingType,
				`Clothing Type`,
				pre.ClothingType,
				`FeatureBullet1`,
				pre.FeatureBullet1,
				`FeatureBullet2`,
				pre.FeatureBullet2,
				`FeatureBullet3`,
				pre.FeatureBullet3,
				`FeatureBullet4`,
				pre.FeatureBullet4,
				`FeatureBullet5`,
				pre.FeatureBullet5,
				`FeatureBullet6`,
				pre.FeatureBullet6,
			}
		}(i+1, pre)
	}
	work.Wait()

	return layout, t.region
}

type parentSKUs map[int]string

func (ps parentSKUs) getVariationParentSKU(prod chapi.Product, prods []chapi.Product) string {
	if prod.IsParent {
		return "Parent"
	}
	sku, exists := ps[prod.ParentProductID]
	if exists {
		return sku
	}

	skuc := make(chan string)

	var misswg sync.WaitGroup
	miss := make(chan int, 1)
	miss <- 0

	for _, p := range prods {
		p := p
		misswg.Add(1)
		go func() {
			if prod.ParentProductID != p.ID {
				miss <- 1 + <-miss
				misswg.Done()
				return
			}
			misswg.Done()
			skuc <- p.Sku
		}()
	}

	misswg.Wait()
	if len(prods) == <-miss {
		return ""
	}

	return <-skuc
}

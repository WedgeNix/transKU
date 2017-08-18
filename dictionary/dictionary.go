package dictionary

import (
	"strings"
	"sync"
	"sync/atomic"

	"golang.org/x/text/currency"

	"github.com/WedgeNix/transKU/regex"

	"github.com/WedgeNix/chapi"
)

// Type holds the dictionary information.
type Type struct {
	jobs   sync.WaitGroup
	lock   sync.RWMutex
	lookup map[string]string
}

var (
	transAttr = map[string]bool{
		`AMZ_Category`:           false,
		`AMZ_Color_Map`:          false,
		`AMZ_Item_Type`:          false,
		`AMZ_ProductIDType`:      false,
		`AMZClothingType`:        false,
		`AMZColor`:               true,
		`AMZDepartment`:          false,
		`AMZDescription`:         true,
		`AMZSize`:                false,
		`AMZTitle`:               true,
		`Apparel-Closure-Type`:   false,
		`Arm Length`:             false,
		`Band Material`:          false,
		`Bottom Style`:           false,
		`Bottoms Size (Men's)`:   false,
		`Bottoms Size (Women's)`: false,
		`Boy's Clothing Type`:    false,
		`Clothing Type`:          false,
		`FeatureBullet1`:         true,
		`FeatureBullet2`:         true,
		`FeatureBullet3`:         true,
		`FeatureBullet4`:         true,
		`FeatureBullet5`:         true,
		`FeatureBullet6`:         true,
	}
)

// GoAdd adds specific product fields into the dictionary (concurrently).
func (dict *Type) GoAdd(prods []chapi.Product) {
	dict.jobs.Add(len(prods))

	for _, prod := range prods {
		go func(prod chapi.Product) {
			defer dict.jobs.Done()

			// dict.stripAndAddText(prod.Title, prod)
			// dict.stripAndAddText(prod.Sku, prod)
			// dict.stripAndAddText(prod.CreateDateUtc, prod)
			// dict.stripAndAddText(prod.Weight, prod)
			// dict.stripAndAddText(prod.UPC, prod)
			// dict.stripAndAddText(prod.ASIN, prod)
			// dict.stripAndAddText(prod.Description, prod)
			// dict.stripAndAddText(prod.Brand, prod)
			// dict.stripAndAddText(prod.Cost, prod)
			// dict.stripAndAddText(prod.BuyItNowPrice, prod)
			// dict.stripAndAddText(prod.RetailPrice, prod)
			// dict.stripAndAddText(prod.Images, prod)
			// dict.stripAndAddText(prod.ReceivedDateUtc, prod)
			// dict.stripAndAddText(prod.RelationshipName, prod)
			// dict.stripAndAddText(prod.ParentProductID, prod)
			// dict.stripAndAddText(prod.Labels, prod)
			// dict.stripAndAddText(prod.Classification, prod)

			for _, attr := range prod.Attributes {
				if transAttr[attr.Name] {
					dict.stripAndAddText(attr.Value, prod)
				}
			}
		}(prod)
	}
}

func (dict *Type) stripAndAddText(text string, prod chapi.Product) {
	brandless := strings.Replace(text, prod.Brand, "", -1)
	tagless := regex.HTML.ReplaceAllString(brandless, "<>")
	phrases := regex.Phrase.FindAllString(tagless, -1)

	for _, phrase := range phrases {
		dict.lock.RLock()
		_, exists := dict.lookup[phrase]
		dict.lock.RUnlock()

		if exists {
			continue
		}

		dict.lock.Lock()
		dict.lookup[phrase] = phrase
		dict.lock.Unlock()
	}
}

// GoSetAll sets every entry in the dictionary using the passed-in function (concurrently).
func (dict *Type) GoSetAll(f func(string) string) {
	dict.jobs.Wait()

	dict.jobs.Add(len(dict.lookup))
	for word := range dict.lookup {
		go func(word string) {
			defer dict.jobs.Done()
			dict.lookup[word] = f(word)
		}(word)
	}
	dict.jobs.Wait()
}

// GoTransAll combs through products and fills up a new version with specifics translated.
func (dict *Type) GoTransAll(prods []chapi.Product) []chapi.Product {
	dict.jobs.Wait()

	newProds := []chapi.Product{}
	// for _, prod := range prods {
	// 	newProd := chapi.Product{
	// 		Title:            prod.Title,
	// 		Sku:              prod.Sku,
	// 		CreateDateUtc:    prod.CreateDateUtc,
	// 		Weight:           prod.Weight,
	// 		UPC:              prod.UPC,
	// 		ASIN:             prod.ASIN,
	// 		Description:      prod.Description,
	// 		Brand:            prod.Brand,
	// 		Cost:             prod.Cost,
	// 		BuyItNowPrice:    prod.BuyItNowPrice,
	// 		RetailPrice:      prod.RetailPrice,
	// 		Images:           prod.Images,
	// 		ReceivedDateUtc:  prod.ReceivedDateUtc,
	// 		RelationshipName: prod.RelationshipName,
	// 		ParentProductID:  prod.ParentProductID,
	// 		Labels:           prod.Labels,
	// 		Classification:   prod.Classification,
	// 	}
	// 	newProds = append(newProds, newProd)
	// }

	dict.jobs.Add(len(dict.lookup))
	for word := range dict.lookup {
		go func(word string) {
			defer dict.jobs.Done()
			// dict.lookup[word] = f(word)
		}(word)
	}
	dict.jobs.Wait()

	return newProds
}

// GoGetPrice counts all dictionary word lengths, returning the cost exchange.
func (dict *Type) GoGetPrice() currency.Amount {
	dict.jobs.Wait()

	charCnt := uint64(0)

	for word := range dict.lookup {
		go func(word string) {
			atomic.AddUint64(&charCnt, uint64(len(word)))
		}(word)
	}

	return currency.USD.Amount(0.00002 * float64(atomic.LoadUint64(&charCnt)))
}

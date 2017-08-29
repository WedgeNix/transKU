package dictionary

import (
	"fmt"
	"strings"
	"sync"

	"golang.org/x/text/currency"

	"github.com/WedgeNix/transKU/regex"

	"github.com/WedgeNix/chapi"
)

// Lookup is a dictionary file type.
type Lookup map[string]string

// Type holds the dictionary information.
type Type struct {
	jobs         sync.WaitGroup
	lock         sync.RWMutex
	cache        Lookup
	cacheCharCnt int
}

// New creates a new dictionary controller.
func New(cache ...Lookup) *Type {
	t := &Type{cache: Lookup{}}
	if len(cache) > 0 {
		t.cache = cache[0]
		t.cacheCharCnt = t.getCharCnt()
	}
	return t
}

func (dict *Type) String() string {
	dict.jobs.Wait()
	return fmt.Sprintln(dict.cache)
}

var (
	// FilterAttr is a mapping of to-translate attributes.
	FilterAttr = map[string]bool{
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

func (dict *Type) filter(p *chapi.Product) []*string {
	fill := []*string{
	// &p.Title,
	// &p.Sku,
	// &p.CreateDateUtc,
	// &p.Weight,
	// &p.UPC,
	// &p.ASIN,
	// &p.Description,
	// &p.Brand,
	// &p.Cost,
	// &p.BuyItNowPrice,
	// &p.RetailPrice,
	// &p.Images,
	// &p.ReceivedDateUtc,
	// &p.RelationshipName,
	// &p.ParentProductID,
	// &p.Labels,
	// &p.Classification,
	}
	for i := range p.Attributes {
		if FilterAttr[p.Attributes[i].Name] {
			fill = append(fill, &p.Attributes[i].Value)
		}
	}
	return fill
}

// GoAdd adds specific product fields into the dictionary (concurrently).
func (dict *Type) GoAdd(prods []chapi.Product) {
	dict.jobs.Add(len(prods))

	for _, prod := range prods {
		if !prod.IsParent {
			dict.jobs.Done()
			continue
		}
		go func(prod chapi.Product) {
			defer dict.jobs.Done()

			fields := dict.filter(&prod)

			for _, field := range fields {
				dict.stripAndAddText(*field, prod)
			}
		}(prod)
	}
}

type bag struct {
	items []string
	tok   string
	tlate bool
}

func strip(text string, prod chapi.Product, onlyPhrases bool) (tags, brands, phrases bag, toks string) {
	noTags := regex.HTML.ReplaceAllString(text, "<>")
	noTagsAndBrand := strings.Replace(noTags, prod.Brand, "[]", -1)

	if !onlyPhrases { // efficiency
		tags = bag{items: regex.HTML.FindAllString(text, -1), tok: "<>"}
		brandCnt := strings.Count(noTags, prod.Brand)
		brndz := []string{}
		brandRng := make([]int, brandCnt)
		for range brandRng {
			brndz = append(brndz, prod.Brand)
		}
		brands = bag{items: brndz, tok: "[]"}
		toks = regex.Phrase.ReplaceAllString(noTagsAndBrand, "{}")
	}

	phrases = bag{items: regex.Phrase.FindAllString(noTagsAndBrand, -1), tok: "{}", tlate: true}
	return
}

func (dict *Type) stripAndAddText(text string, prod chapi.Product) {
	_, _, phrases, _ := strip(text, prod, true)

	for _, phrase := range phrases.items {
		dict.lock.RLock()
		_, exists := dict.cache[phrase]
		dict.lock.RUnlock()

		if exists {
			continue
		}

		dict.lock.Lock()
		dict.cache[phrase] = ""
		dict.lock.Unlock()
	}
}

// GoFillAll sets every entry in the dictionary using the passed-in function (concurrently).
func (dict *Type) GoFillAll(cacheMiss func(string) string) {
	dict.jobs.Wait()

	newEntries := make(chan Lookup, 1)
	newEntries <- Lookup{}

	fmt.Println(`len(dict.cache)=`, len(dict.cache))
	for word, tlate := range dict.cache {
		if len(tlate) > 0 {
			// fmt.Print(`O`)
			continue
		}

		word := word

		dict.jobs.Add(1)
		go func() {
			defer dict.jobs.Done()
			// fmt.Print(`X`)

			tlate := cacheMiss(word)
			// fmt.Print(word + `>>` + tlate)

			e := <-newEntries
			e[word] = tlate
			newEntries <- e
		}()
	}
	dict.jobs.Wait()

	for word, tlate := range <-newEntries {
		dict.cache[word] = tlate
	}
}

func (dict *Type) swapNShift(toks string, b bag) string {
	for _, itm := range b.items {
		if b.tlate {
			dict.lock.RLock()
			itm = dict.cache[itm]
			dict.lock.RUnlock()
		}
		toks = strings.Replace(toks, b.tok, itm, 1)
	}
	return toks
}

// GoTransAll combs through products and fills up a new version with specifics translated.
func (dict *Type) GoTransAll(prods []chapi.Product) []chapi.Product {
	dict.jobs.Wait()

	newProds := make([]chapi.Product, len(prods))

	dict.jobs.Add(len(prods))
	for i, prod := range prods {
		go func(i int, prod chapi.Product) {
			defer dict.jobs.Done()

			fields := dict.filter(&prod)

			for _, field := range fields {
				tags, brands, phrases, toks := strip(*field, prod, false)

				toks = dict.swapNShift(toks, tags)
				toks = dict.swapNShift(toks, brands)
				toks = dict.swapNShift(toks, phrases)

				*field = toks
			}

			newProds[i] = prod
		}(i, prod)
	}
	dict.jobs.Wait()

	return newProds
}

// SHOULD CORRECT FOR DIFFERENCE BETWEEN LOADED CACHE AND NEW ENTRIES.
// SHOULD CORRECT FOR DIFFERENCE BETWEEN LOADED CACHE AND NEW ENTRIES.
// SHOULD CORRECT FOR DIFFERENCE BETWEEN LOADED CACHE AND NEW ENTRIES.

func (dict *Type) getNewCharCnt() int {
	return dict.getCharCnt() - dict.cacheCharCnt
}

// GetPrice counts all dictionary word lengths, returning the cost exchange.
func (dict *Type) GetPrice() currency.Amount {
	return currency.USD.Amount(0.00002 * float64(dict.getNewCharCnt()))
}

func (dict *Type) getCharCnt() int {
	dict.jobs.Wait()

	charCnt := 0

	for word := range dict.cache {
		charCnt += len(word)
	}

	return charCnt
}

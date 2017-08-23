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
	filterAttr = map[string]bool{
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

func (dict *Type) filter(p chapi.Product) []*string {
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
	for _, attr := range p.Attributes {
		if filterAttr[attr.Name] {
			fill = append(fill, &attr.Value)
		}
	}
	return fill
}

// GoAdd adds specific product fields into the dictionary (concurrently).
func (dict *Type) GoAdd(prods []chapi.Product) {
	dict.jobs.Add(len(prods))

	for _, prod := range prods {
		go func(prod chapi.Product) {
			defer dict.jobs.Done()

			fields := dict.filter(prod)

			for _, field := range fields {
				dict.stripAndAddText(*field, prod)
			}
		}(prod)
	}
}

type bag struct {
	items []string
	tok   string
}

// type safebag struct {
// 	unsafe_lock sync.Mutex
// 	unsafe_bag  []string
// 	tok         string
// }

// func (b *safebag) pop() string {
// 	b.unsafe_lock.Lock()
// 	defer b.unsafe_lock.Unlock()
// 	head, newB := b.unsafe_bag[0], b.unsafe_bag[1:]
// 	b.unsafe_bag = newB
// 	return head
// }

// type safetoks struct {
// 	unsafe_work sync.WaitGroup
// 	unsafe_lock sync.Mutex
// 	unsafe_toks string
// }

// func (t *safetoks) goswap(b *safebag) {
// 	bc := len(b.unsafe_bag)
// 	t.unsafe_work.Add(bc)

// 	bags := make([]int, bc)

// 	for range bags {
// 		go func() {
// 			defer t.unsafe_work.Done()
// 			t.unsafe_lock.Lock()
// 			defer t.unsafe_lock.Unlock()
// 			t.unsafe_toks = strings.Replace(t.unsafe_toks, b.tok, b.pop(), 1)
// 		}()
// 	}
// }
// func (t *safetoks) get() string {
// 	t.unsafe_work.Wait()
// 	return t.unsafe_toks
// }

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

	phrases = bag{items: regex.Phrase.FindAllString(noTagsAndBrand, -1), tok: "{}"}
	return
}

func (dict *Type) stripAndAddText(text string, prod chapi.Product) {
	_, _, phrases, _ := strip(text, prod, true)

	for _, phrase := range phrases.items {
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

// GoFillAll sets every entry in the dictionary using the passed-in function (concurrently).
func (dict *Type) GoFillAll(f func(string) string) {
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

func swapNShift(toks string, b bag) string {
	for _, itm := range b.items {
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

			fields := dict.filter(prod)

			for _, field := range fields {
				tags, brands, phrases, toks := strip(*field, prod, false)

				toks = swapNShift(toks, tags)
				toks = swapNShift(toks, brands)
				toks = swapNShift(toks, phrases)

				*field = dict.lookup[toks]
			}

			newProds[i] = prod
		}(i, prod)
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

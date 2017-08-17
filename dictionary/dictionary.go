package dictionary

import (
	"strings"
	"sync"

	"github.com/WedgeNix/transKU/regex"

	"github.com/WedgeNix/chapi"
)

// Type holds the dictionary information.
type Type struct {
	jobs   sync.WaitGroup
	lock   sync.RWMutex
	lookup map[string]string
}

// GoAddTitles adds product titles into the dictionary (concurrently).
func (dict *Type) GoAddTitles(prods []chapi.Product) {
	dict.jobs.Add(len(prods))
	for _, prod := range prods {
		go func(prod chapi.Product) {
			defer dict.jobs.Done()

			title := strings.Replace(prod.Title, prod.Brand, "", -1)

			dict.lock.RLock()
			_, exists := dict.lookup[title]
			dict.lock.RUnlock()

			if exists {
				return
			}

			dict.lock.Lock()
			dict.lookup[title] = title
			dict.lock.Unlock()
		}(prod)
	}
}

// GoAddAttributes adds specific product attributes into the dictionary (concurrently).
func (dict *Type) GoAddAttributes(prods []chapi.Product) {
	dict.jobs.Add(len(prods))
	for _, prod := range prods {
		go func(prod chapi.Product) {
			defer dict.jobs.Done()

			for _, attr := range prod.Attributes {
				if !strings.Contains(attr.Name, "FeatureBullet") {
					continue
				}

				dict.lock.RLock()
				_, exists := dict.lookup[attr.Value]
				dict.lock.RUnlock()

				if exists {
					continue
				}

				dict.lock.Lock()
				dict.lookup[attr.Value] = attr.Value
				dict.lock.Unlock()
			}
		}(prod)
	}
}

// GoAddDescriptions adds all product descriptions into the dictionary (concurrently).
func (dict *Type) GoAddDescriptions(prods []chapi.Product) {
	dict.jobs.Add(len(prods))
	for _, prod := range prods {
		go func(prod chapi.Product) {
			defer dict.jobs.Done()

			desc := prod.Description
			cleanDesc := regex.HTML.ReplaceAllString(desc, "<>")
			phrases := regex.Phrase.FindAllString(cleanDesc, -1)

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
		}(prod)
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

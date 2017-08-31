package transku

import (
	"sync"

	"github.com/WedgeNix/awsapi"
	"github.com/WedgeNix/chapi"
	"github.com/WedgeNix/gosetta"
)

// Region holds data needed for dynamic region integration.
type Region struct {
	BCP47      string
	ChannelTag string
	ProfileID  int
}

// TransKU holds transKU controller data.
type TransKU struct {
	ca    *chapi.CaObj
	prods []chapi.Product
	aws   *awsapi.Controller
	rose  *gosetta.Rose
}

// Dictionary holds the dictionary information.
type Dictionary struct {
	jobs         sync.WaitGroup
	lock         sync.RWMutex
	cache        lookup
	cacheCharCnt int
}

type lookup map[string]string

type bag struct {
	items []string
	tok   string
	tlate bool
}

package net

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/wuyazero/Elastos.ELA.SPV/log"
)

const (
	CachedAddrsFile = "addrs.cache"
)

type AddrManager struct {
	sync.RWMutex
	seeds     []string
	cached    []string
	connected map[string]byte
}

func newAddrManager(seeds []string) *AddrManager {
	am := &AddrManager{
		seeds:     make([]string, 0),
		cached:    make([]string, 0),
		connected: make(map[string]byte),
	}

	// Read seed list from config file
	for _, addr := range seeds {
		am.seeds = append(am.seeds, addr)
	}

	// Read cached addresses from file
	data, err := ioutil.ReadFile(CachedAddrsFile)
	if err != nil {
		return am
	}
	addrs := strings.Split(strings.TrimSpace(string(data)), "\n")

	for _, addr := range addrs {
		if len(strings.TrimSpace(addr)) != 0 {
			am.cached = append(am.cached, addr)
		}
	}

	return am
}

func (am *AddrManager) GetIdleAddrs(count int) []string {
	addrMap := make(map[string]string)

	for _, seed := range am.seeds {
		if am.isConnected(seed) {
			continue
		}
		addrMap[seed] = seed
	}

	for _, cache := range am.cached {
		if am.isConnected(cache) {
			continue
		}
		addrMap[cache] = cache
	}

	totalAddrs := len(addrMap)
	if count > totalAddrs {
		count = totalAddrs
	}

	randAddrs := make([]string, count)
	for addr := range addrMap {
		count--
		randAddrs[count] = addr
		if count == 0 {
			break
		}
	}

	return randAddrs
}

func (am *AddrManager) AddAddr(addr string) {
	am.Lock()
	defer am.Unlock()

	am.connected[addr] = 'c'

	if !am.isSeed(addr) && !am.isCached(addr) {
		am.cached = append(am.cached, addr)
		am.saveCached()
	}
}

func (am *AddrManager) DisconnectedAddr(addr string) {
	am.Lock()
	defer am.Unlock()

	delete(am.connected, addr)
}

func (am *AddrManager) DiscardAddr(addr string) {
	am.Lock()
	defer am.Unlock()

	log.Info("AddrManager discard addr:", addr)
	for i, cache := range am.cached {
		if cache == addr {
			am.cached = append(am.cached[:i], am.cached[i+1:]...)
			am.saveCached()
			return
		}
	}
}

func (am *AddrManager) isSeed(addr string) bool {
	for _, seed := range am.seeds {
		if seed == addr {
			return true
		}
	}
	return false
}

func (am *AddrManager) isCached(addr string) bool {
	for _, cached := range am.cached {
		if cached == addr {
			return true
		}
	}
	return false
}

func (am *AddrManager) isConnected(addr string) bool {
	_, ok := am.connected[addr]
	return ok
}

func (am *AddrManager) saveCached() {
	var cached string
	for _, addr := range am.cached {
		cached += string(addr)
		cached += "\n"
	}

	file, err := os.OpenFile(CachedAddrsFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Println("Open cached addresses failed")
		return
	}

	_, err = file.Write([]byte(cached))
	if err != nil {
		fmt.Println("Write cached addresses failed")
		return
	}
}

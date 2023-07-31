package common

import (
	"fmt"
	"time"

	"github.com/eoscanada/eos-go"
)

var endpoints = []string{
	"https://wax.greymass.com", // Priority endpoint
	"https://wax.eosdac.io",
	"https://wax.pink.gg",
	"https://wax.api.eosnation.io",
	"https://wax.eosrio.io",
	"https://wax.eosusa.io",
	"https://wax.eu.eosamsterdam.net",
	"https://api.wax.bountyblok.io",
	"https://api.waxsweden.org",
	"https://api.wax.alohaeos.com",
	"https://waxapi.ledgerwise.io",
	"https://api.wax.detroitledger.tech",
	"https://wax.eosphere.io",
	"https://api-wax.eosauthority.com",
	"https://wax-public1.neftyblocks.com",
	"https://wax-public2.neftyblocks.com",
	"https://apiwax.3dkrender.com",
	"https://aa.wax.blacklusion.io",
	"https://hyperion.wax.blacklusion.io",
	"https://wax.blacklusion.io",
	"https://wax.hivebp.io",
}

func GetAPIEndpoint(failureCount int) string {
	if failureCount >= len(endpoints) {
		fmt.Println("🔁 Reached max endpoint switches. 🔁 \n Waiting for 30 seconds...")
		time.Sleep(30 * time.Second)
		return endpoints[0] // Reset to the first endpoint
	} else if failureCount > 0 {
		fmt.Println("🔀 Non-preferential API request detected. 🔀 \n Waiting for 2 seconds...")
		time.Sleep(2 * time.Second)
	}
	return endpoints[failureCount]
}

func RemoveStakeFromChain(poolId uint64, assetIds []uint64) (eos.Checksum256, error) {
	// Remove the stake from the chain
	// EOS logic should be implemented here
	return eos.Checksum256{}, nil
}

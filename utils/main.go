package utils

import (
	"github.com/ipfs/go-cid"
)

// isIPFS checks if a given string `str` contains an IPFS CID or not.
func isIPFS(str string) string {
	c, err := cid.Parse(str)
	if err != nil {
		return ""
	}

	return c.String()
}

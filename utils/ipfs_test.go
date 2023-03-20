package utils

import (
	"testing"
)

func TestIsIPFS(t *testing.T) {
	s := isIPFS("https://snapshot.mypinata.cloud/ipfs/bafkreiap5ojtthjysktrjzxrowk6xufex3jqunkgkydwyrxv7ff3dfceiu")
	cid := "bafkreiap5ojtthjysktrjzxrowk6xufex3jqunkgkydwyrxv7ff3dfceiu"

	if s != cid {
		t.Fatal("unable to parse IPFS link", "expected", cid, "got", s)
	}

	s = isIPFS("https://snapshot.mypinata.cloud/ipfs/Qafkreiap5ojtthjysktrjzxrowk6xufex3jqunkgkydwyrxv7ff3dfceiu")
	cid = ""
	if s != cid {
		t.Fatal("unable to parse IPFS link", "expected", cid, "got", s)
	}
}

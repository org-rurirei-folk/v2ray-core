package fakedns

import (
	"golang.org/x/net/dns/dnsmessage"

	"github.com/v2fly/v2ray-core/v4/common"
)

type SniffHeader struct {
	domain string
}

func (h *SniffHeader) Protocol() string {
	return "fakedns"
}

func (h *SniffHeader) Domain() string {
	return h.domain
}

func SniffFakeDNS(b []byte) (*SniffHeader, error) {
	h := &SniffHeader{}

	isIPQuery, domain, _, _ := ParseIPQuery(b)

	if isIPQuery {
		h.domain = domain
		return h, nil
	}

	return nil, common.ErrNoClue
}

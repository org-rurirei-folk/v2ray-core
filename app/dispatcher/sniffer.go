// +build !confonly

package dispatcher

import (
	"context"

	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/protocol/bittorrent"
	"github.com/v2fly/v2ray-core/v4/common/protocol/http"
	"github.com/v2fly/v2ray-core/v4/common/protocol/tls"
)

type SniffResult interface {
	Protocol() string
	Domain() string
}

type protocolSniffer func(context.Context, []byte) (SniffResult, error)

type snifferWithMetadata struct {
	// domainSniffer domainSniffer
	protocolSniffer protocolSniffer
	// A Metadata sniffer will be invoked on connection establishment only, with nil body,
	// for both TCP and UDP connections
	// It will not be shown as a traffic type for routing unless there is no other successful sniffing.
	metadataSniffer bool
}

type Sniffer struct {
	sniffer []snifferWithMetadata
}

func NewSniffer(ctx context.Context) *Sniffer {
	if alternativeSniffer, metadata, err := UseAlternativeSniffer(); err == nil && alternativeSniffer != nil {
		return &Sniffer{
			sniffer: []snifferWithMetadata{
				{alternativeSniffer, metadata},
			},
		}
	}

	ret := &Sniffer{
		sniffer: []snifferWithMetadata{
			{func(c context.Context, b []byte) (SniffResult, error) { return http.SniffHTTP(b) }, false},
			{func(c context.Context, b []byte) (SniffResult, error) { return tls.SniffTLS(b) }, false},
			{func(c context.Context, b []byte) (SniffResult, error) { return bittorrent.SniffBittorrent(b) }, false},
		},
	}
	if sniffer, err := newFakeDNSSniffer(ctx); err == nil {
		others := ret.sniffer
		ret.sniffer = append(ret.sniffer, sniffer)
		fakeDNSThenOthers, err := newFakeDNSThenOthers(ctx, sniffer, others)
		if err == nil {
			ret.sniffer = append([]snifferWithMetadata{fakeDNSThenOthers}, ret.sniffer...)
		}
	}
	return ret
}

var UseAlternativeSniffer = func() (protocolSniffer, bool, error) {
	return nil, false, nil
}

var errUnknownContent = newError("unknown content")

func (s *Sniffer) Sniff(c context.Context, payload []byte, metadataSniffer bool) (SniffResult, error) {
	var pendingSniffer []snifferWithMetadata
	for _, si := range s.sniffer {
		s := si.protocolSniffer
		if si.metadataSniffer != metadataSniffer {
			continue
		}
		result, err := s(c, payload)
		if err == common.ErrNoClue {
			pendingSniffer = append(pendingSniffer, si)
			continue
		}

		if err == nil && result != nil {
			return result, nil
		}
	}

	if len(pendingSniffer) > 0 {
		s.sniffer = pendingSniffer
		return nil, common.ErrNoClue
	}

	return nil, errUnknownContent
}

func CompositeResult(domainResult SniffResult, protocolResult SniffResult) SniffResult {
	return &compositeResult{domainResult: domainResult, protocolResult: protocolResult}
}

type compositeResult struct {
	domainResult   SniffResult
	protocolResult SniffResult
}

func (c compositeResult) Protocol() string {
	return c.protocolResult.Protocol()
}

func (c compositeResult) Domain() string {
	return c.domainResult.Domain()
}

func (c compositeResult) ProtocolForDomainResult() string {
	return c.domainResult.Protocol()
}

type SnifferResultComposite interface {
	ProtocolForDomainResult() string
}

type SnifferIsProtoSubsetOf interface {
	IsProtoSubsetOf(protocolName string) bool
}

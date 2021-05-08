package dns

import (
	"context"
	"io"
	"sync"

	"golang.org/x/net/dns/dnsmessage"

	core "github.com/v2fly/v2ray-core/v4"
	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/buf"
	"github.com/v2fly/v2ray-core/v4/common/net"
	dns_proto "github.com/v2fly/v2ray-core/v4/common/protocol/dns"
	"github.com/v2fly/v2ray-core/v4/common/session"
	"github.com/v2fly/v2ray-core/v4/common/task"
	"github.com/v2fly/v2ray-core/v4/features/dns"
	"github.com/v2fly/v2ray-core/v4/transport"
	"github.com/v2fly/v2ray-core/v4/transport/internet"
)

func ParseIPQuery(b []byte) (r bool, domain string, id uint16, qType dnsmessage.Type) {
	var parser dnsmessage.Parser
	header, err := parser.Start(b)
	if err != nil {
		newError("parser start").Base(err).WriteToLog()
		return
	}

	id = header.ID
	q, err := parser.Question()
	if err != nil {
		newError("question").Base(err).WriteToLog()
		return
	}
	qType = q.Type
	if qType != dnsmessage.TypeA && qType != dnsmessage.TypeAAAA {
		return
	}

	domain = q.Name.String()
	r = true
	return
}

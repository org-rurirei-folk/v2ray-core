package net

// ParseNetwork converts a network from its string presentation.
func ParseNetwork(network string) (Network, error) {
	switch network {
	case "tcp":
		return Network_TCP, nil
	case "udp":
		return Network_UDP, nil
	default:
		return Network_Unknown, newError("unsupported network ", network)
	}
}

func (n Network) SystemString() string {
	switch n {
	case Network_TCP:
		return "tcp"
	case Network_UDP:
		return "udp"
	case Network_UNIX:
		return "unix"
	default:
		return "unknown"
	}
}

// HasNetwork returns true if the network list has a certain network.
func HasNetwork(list []Network, network Network) bool {
	for _, value := range list {
		if value == network {
			return true
		}
	}
	return false
}

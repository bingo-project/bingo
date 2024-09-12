package ip

import "net"

func Contains(cidr, ip string) bool {
	if cidr == ip {
		return true
	}

	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}

	if ipNet.Contains(net.ParseIP(ip)) {
		return true
	}

	return false
}

func ContainsInCIDR(cidr []string, ip string) bool {
	for _, item := range cidr {
		if Contains(item, ip) {
			return true
		}
	}

	return false
}

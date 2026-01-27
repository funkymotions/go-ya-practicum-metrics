package utils

import "net"

func ParseCIDRString(cidrStr string) (*net.IPNet, error) {
	_, ipnet, err := net.ParseCIDR(cidrStr)
	if err != nil {
		return nil, err
	}

	return ipnet, nil
}

func IsIPInCIDR(clientIP net.IP, cidr *net.IPNet) bool {
	return cidr.Contains(clientIP)
}

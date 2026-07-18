package helpers

import (
	"fmt"
	"net/netip"
)

const RealIpHeader = "X-Real-IP"

func ValidateMaskCIDR(cidr string) error {

	_, err := netip.ParsePrefix(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR %q: %w", cidr, err)
	}
	return nil

}

func ValidateIPByCIDR(value string, cidr string) error {
	ip, err := netip.ParseAddr(value)
	if err != nil {
		return fmt.Errorf("invalid IP %q: %w", value, err)
	}

	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR %q: %w", cidr, err)
	}

	if !prefix.Contains(ip) {
		return fmt.Errorf("IP %q is not within CIDR %q", value, cidr)
	}

	return nil
}

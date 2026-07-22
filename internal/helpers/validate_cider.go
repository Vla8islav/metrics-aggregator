package helpers

import (
	"fmt"
	"net/netip"
	"strings"
)

const RealIpHeader = "X-Real-IP"

func ValidateMaskCIDR(cidr string) error {

	_, err := netip.ParsePrefix(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR %q: %w", cidr, err)
	}
	return nil

}

func ValidateIPByCIDR(value string, subnetsCIDR string) error {
	ip, err := netip.ParseAddr(value)
	if err != nil {
		return fmt.Errorf("invalid IP %q: %w", value, err)
	}

	for _, rawSubnet := range strings.Split(subnetsCIDR, ",") {
		subnet := strings.TrimSpace(rawSubnet)
		prefix, errPrefix := netip.ParsePrefix(subnet)
		if errPrefix != nil {
			return fmt.Errorf("invalid CIDR %q: %w", subnetsCIDR, err)
		}

		if prefix.Contains(ip) {
			return nil
		}
	}

	return fmt.Errorf("IP %q is not within CIDRs %q", value, subnetsCIDR)
}

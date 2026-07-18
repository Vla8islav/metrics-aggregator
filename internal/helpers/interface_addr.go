package helpers

import (
	"fmt"
	"net"
)

func GetInterfaceAddr() (string, error) {
	interfaces, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range interfaces {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}
		ip := ipNet.IP.To4()
		if ip != nil {
			return ip.String(), nil
		}
	}
	return "", fmt.Errorf("no valid IP address found")
}

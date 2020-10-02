package system

import (
	"log"
	"net"
	"strings"
)

func GetNetInterfaceAddresses() []string {

	interfaces, err := net.Interfaces()
	if err != nil {
		log.Printf("Error, no interfaces: %s\n", err)
		return []string{}
	}
	var rv = []string{}
	for _, iface := range interfaces {
		if strings.HasPrefix(iface.Name, "lo") {
			continue
		}
		addrs, err := iface.Addrs()

		if err != nil {
			log.Printf(" %s. %s\n", iface.Name, err)
			continue
		}
		for _, a := range addrs {
			addr := a.String()
			if !strings.Contains(addr, ":") {
				if strings.HasPrefix(addr, "127.") || strings.HasPrefix(addr, "169.254.") {
					continue
				}
				rv = append(rv, addr[:strings.Index(addr, "/")])
			}
		}
	}
	return rv
}

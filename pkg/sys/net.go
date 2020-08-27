package sys

import (
	"bufio"
	"net"
	"os"
	"strings"
)

type NetworkInterface struct {
	Name   string
	HWAddr string
	IPAddr []net.IP
}

func ProbeNetworkInterface() (NetworkInterface, error) {
	var (
		eth  NetworkInterface
		file = "/proc/net/route"
	)

	i, err := net.Interfaces()
	fd, err := os.Open(file)
	if err != nil {
		return NetworkInterface{}, err
	}
	defer fd.Close()
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		scanner.Scan()
		tokens := strings.Split(scanner.Text(), "\t")
		eth.Name = tokens[0]
		for _, d := range i {
			if d.Name == tokens[0] {
				addrs, _ := d.Addrs()
				for _, addr := range addrs {
					ip, _, _ := net.ParseCIDR(addr.String())
					eth.IPAddr = append(eth.IPAddr, ip)
				}
				eth.HWAddr = d.HardwareAddr.String()
			}
		}
		break
	}
	return eth, nil
}

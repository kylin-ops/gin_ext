package netif

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strings"
)

func selsectDnsServer() []string {
	rgx := regexp.MustCompile("^nameserver.*")
	var dnsServers []string
	if f, err := ioutil.ReadFile("/etc/resolv.conf"); err == nil {
		for _, row := range bytes.Split(f, []byte("\n")) {
			if ok := rgx.MatchString(string(row)); ok {
				line := strings.Split(string(row), " ")
				if len(line) > 1 {
					dnsServers = append(dnsServers, line[len(line)-1])
				}
			}
		}
	}
	dnsServers = append(dnsServers, []string{"114.114.114.114", "8.8.8.8"}...)
	fmt.Println(dnsServers)
	return dnsServers
}

// network  tcp|udp
// address  netif:port
func GetIpFromDial(network, address string) (string, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	ip := strings.Split(conn.LocalAddr().String(), ":")[0]
	return ip, nil
}

func GetIpFromNetIf() string {
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() {
			ipv4 := ipNet.IP.To4()
			if ipv4 != nil {
				return ipv4.String()
			}
		}
	}
	return ""
}

func GetIp() string {
	for _, dnsServer := range selsectDnsServer() {
		if ip, err := GetIpFromDial("udp", dnsServer+":53"); err == nil {
			return ip
		}
	}
	return GetIpFromNetIf()
}

func GetHostName() string {
	name, _ := os.Hostname()
	return name
}

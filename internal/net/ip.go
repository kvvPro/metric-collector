package ip

import (
	"log"
	"net"
)

// Get preferred outbound ip of this machine
func GetOutboundIP(serverIP string) net.IP {
	conn, err := net.Dial("udp", serverIP) // указывать адрес сервера на который будем стучаться
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func CheckIPInSubnet(ip string, subnet string) (bool, error) {
	_, subnetCIDR, err := net.ParseCIDR(subnet)
	if err != nil {
		return false, err
	}
	ipCIDR := net.ParseIP(ip)

	return subnetCIDR.Contains(ipCIDR), nil
}

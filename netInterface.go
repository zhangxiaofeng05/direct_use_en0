package direct_use_en0

import (
	"log"
	"net"
)

func PrintInterfaceIndex() {
	vpnIndex, err := GetInterfaceIndex("utun0")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("vpnIndex: %d", vpnIndex)
	enIndex, err := GetInterfaceIndex("en0")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("enIndex: %d", enIndex)
}

func GetInterfaceIndex(name string) (int, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return 0, err
	}
	return iface.Index, nil
}

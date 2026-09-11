package direct_use_en0

import "fmt"

const (
	En0InterfaceName   = "en0"
	Utun0InterfaceName = "utun0"
	Utun1InterfaceName = "utun1"
	Utun2InterfaceName = "utun2"
	Utun3InterfaceName = "utun3"
	Utun4InterfaceName = "utun4"
	Utun5InterfaceName = "utun5"
	Utun6InterfaceName = "utun6"

	// listen ip, only for local access
	Ip = "127.0.0.1"
)

// Addr returns the listen address for the given port.
func Addr(port int) string {
	return fmt.Sprintf("%s:%d", Ip, port)
}

package conf

import (
	"fmt"
	"net"
)

type Network struct {
	Interface_ string         `yaml:"interface"`
	LocalAddr_ string         `yaml:"local_addr"`
	RouterMac_ string         `yaml:"router_mac"`
	PCAP       PCAP           `yaml:"pcap"`
	TCP        TCP            `yaml:"tcp"`
	Interface  *net.Interface `yaml:"-"`
	Router     *net.Interface `yaml:"-"`
	LocalAddr  net.UDPAddr    `yaml:"-"`
}

func (n *Network) setDefaults(role string) {
	n.PCAP.setDefaults(role)
	n.TCP.setDefaults()
}

func (n *Network) validate() []error {
	var errors []error

	if n.Interface_ == "" {
		errors = append(errors, fmt.Errorf("network interface is required"))
	}
	if len(n.Interface_) > 15 {
		errors = append(errors, fmt.Errorf("network interface name too long (max 15 characters): '%s'", n.Interface_))
	}
	lIface, err := net.InterfaceByName(n.Interface_)
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to find network interface %s: %v", n.Interface_, err))
	}
	n.Interface = lIface

	l, err := validateAddr(n.LocalAddr_, false)
	if err != nil {
		errors = append(errors, err)
	}
	n.LocalAddr = *l

	if n.RouterMac_ == "" {
		errors = append(errors, fmt.Errorf("MAC address is required"))
	}

	hwAddr, err := net.ParseMAC(n.RouterMac_)
	if err != nil {
		errors = append(errors, fmt.Errorf("invalid MAC address '%s': %v", n.RouterMac_, err))
	}
	n.Router = &net.Interface{HardwareAddr: hwAddr}

	errors = append(errors, n.PCAP.validate()...)
	errors = append(errors, n.TCP.validate()...)

	return errors
}

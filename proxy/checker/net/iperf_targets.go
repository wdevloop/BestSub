package net

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bestruirui/bestsub/proxy/checker"
)

// IperfTarget represents an iperf3 test server
type IperfTarget struct {
	Code    int    `json:"code"`
	Server  string `json:"server"`
	PortL   int    `json:"portl"`
	PortU   int    `json:"portu"`
	City    string `json:"city"`
	CityZh  string `json:"cityzh"`
}

// LoadIperfTargets loads iperf test targets from resource file
func LoadIperfTargets() ([]IperfTarget, error) {
	iperfPath, err := checker.GetResourcePath("iperf.json")
	if err != nil {
		return nil, fmt.Errorf("failed to get iperf.json path: %w", err)
	}
	
	data, err := os.ReadFile(iperfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read iperf file: %w", err)
	}
	
	var targets []IperfTarget
	if err := json.Unmarshal(data, &targets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal iperf data: %w", err)
	}
	
	return targets, nil
}

// GetDefaultIperfTargets returns default iperf targets if resource file is unavailable
func GetDefaultIperfTargets() []IperfTarget {
	return []IperfTarget{
		{Code: 11, Server: "iperf3.he.net", PortL: 5201, PortU: 5209, City: "Fremont", CityZh: "弗里蒙特"},
		{Code: 12, Server: "lon.speedtest.clouvider.net", PortL: 5200, PortU: 5209, City: "London", CityZh: "伦敦"},
		{Code: 21, Server: "fra.speedtest.clouvider.net", PortL: 5200, PortU: 5209, City: "Frankfurt", CityZh: "法兰克福"},
		{Code: 22, Server: "nyc.speedtest.clouvider.net", PortL: 5200, PortU: 5209, City: "New York", CityZh: "纽约"},
		{Code: 31, Server: "dal.speedtest.clouvider.net", PortL: 5200, PortU: 5209, City: "Dallas", CityZh: "达拉斯"},
		{Code: 32, Server: "la.speedtest.clouvider.net", PortL: 5200, PortU: 5209, City: "Los Angeles", CityZh: "洛杉矶"},
		{Code: 41, Server: "speedtest.fra1.de.leaseweb.net", PortL: 5201, PortU: 5210, City: "Frankfurt", CityZh: "法兰克福"},
		{Code: 42, Server: "speedtest.wdc1.us.leaseweb.net", PortL: 5201, PortU: 5210, City: "Washington", CityZh: "华盛顿"},
		{Code: 43, Server: "speedtest.ams1.nl.leaseweb.net", PortL: 5201, PortU: 5210, City: "Amsterdam", CityZh: "阿姆斯特丹"},
		{Code: 51, Server: "speedtest.tokyo2.linode.com", PortL: 5201, PortU: 5210, City: "Tokyo", CityZh: "东京"},
		{Code: 52, Server: "speedtest.singapore.linode.com", PortL: 5201, PortU: 5210, City: "Singapore", CityZh: "新加坡"},
	}
} 
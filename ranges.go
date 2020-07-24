package awsranges

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/user"
	"path"
	"time"
)

const (
	awsRangesURL  string = "https://ip-ranges.amazonaws.com/ip-ranges.json"
	cacheFileName string = ".aws-ranges.json"
)

// Prefix is a representation of given IP prefix, region and service
type Prefix struct {
	IP      string `json:"ip_prefix"`
	Region  string
	Service string
}

// Ranges contains the entire list of AWS Prefixes and an HTTP client
// used to pull data down from AWS
type Ranges struct {
	Prefixes []Prefix
	Client   *http.Client
}

// CheckAddress checks if a given address is owned by AWS
func (r *Ranges) CheckAddress(address string) (bool, error) {
	parsedAddr := net.ParseIP(address)

	for _, prefix := range r.Prefixes {
		_, network, _ := net.ParseCIDR(prefix.IP)
		if network.Contains(parsedAddr) {
			return true, nil
		}
	}

	return false, nil
}

// CheckCIDR checks if a given network is owned by AWS
func (r *Ranges) CheckCIDR(cidr string) (bool, error) {
	cidrFirstDigit := cidr[0]
	for _, prefix := range r.Prefixes {
		if cidrFirstDigit != prefix.IP[0] {
			continue
		}
		if prefix.IP == cidr {
			return true, nil
		}
		ip, _, _ := net.ParseCIDR(cidr)
		_, prefixNetwork, _ := net.ParseCIDR(prefix.IP)
		if prefixNetwork.Contains(ip) {
			return true, nil
		}
	}
	return false, nil
}

// ServicesResponse contains the region and services assigned to an IP/network
type ServicesResponse struct {
	Region   string
	Services []string
}

// CheckServices determines what services and region an IP address is assigned to
func (r *Ranges) CheckServices(address string) (*ServicesResponse, error) {
	parsedAddr := net.ParseIP(address)

	var answer Prefix
	var services []string
	for _, prefix := range r.Prefixes {
		_, network, _ := net.ParseCIDR(prefix.IP)
		if network.Contains(parsedAddr) {
			answer = prefix
			services = append(services, prefix.Service)
		}
	}

	if answer.Service != "" {
		if len(services) > 1 {
			return &ServicesResponse{
				Region:   answer.Region,
				Services: services,
			}, nil
		}
		return &ServicesResponse{
			Region:   answer.Region,
			Services: []string{answer.Service},
		}, nil
	}

	return &ServicesResponse{}, nil
}

// New returns a new instance of the Ranges object
func New() (*Ranges, error) {
	useCache := false

	u, err := user.Current()
	if err != nil {
		return nil, err
	}

	cachedFile := path.Join(u.HomeDir, cacheFileName)
	if fileExists(cachedFile) {
		useCache = true
	}

	client := httpClient()
	var ranges Ranges
	ranges.Client = client

	var data []byte
	if useCache {
		data, err = readFromCache(cachedFile)
		if err != nil {
			return nil, err
		}
	} else {
		res, err := ranges.Client.Get(awsRangesURL)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		data, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
	}

	err = json.Unmarshal(data, &ranges)
	if err != nil {
		return nil, err
	}
	err = writeToCache(data, cachedFile)
	if err != nil {
		return nil, err
	}

	return &ranges, nil
}

func httpClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConnsPerHost:   -1,
			DisableKeepAlives:     true,
		},
	}
}

func fileExists(f string) bool {
	_, err := os.Stat(f)
	if os.IsNotExist(err) || err != nil {
		return false
	}
	return true
}

func readFromCache(cacheFile string) ([]byte, error) {
	fileReader, err := os.Open(cacheFile)
	if err != nil {
		return nil, fmt.Errorf("unable to open cached file: %+v", err)
	}
	defer fileReader.Close()

	return ioutil.ReadAll(fileReader)
}

func writeToCache(data []byte, cacheFile string) error {
	fileReader, err := os.Open(cacheFile)
	if err != nil {
		return fmt.Errorf("unable to open cached file: %+v", err)
	}
	defer fileReader.Close()

	_, err = io.Copy(fileReader, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("unable to copy contents to cached file: %+v", err)
	}
	return nil
}

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package staticdiscovery

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/trustbloc/trustbloc-did-method/pkg/vdri/trustbloc/config"
	"github.com/trustbloc/trustbloc-did-method/pkg/vdri/trustbloc/endpoint"
)

// DiscoveryService implements a static discovery service
type DiscoveryService struct {
	httpClient *http.Client
}

// NewService return static discovery service
func NewService() *DiscoveryService {
	return &DiscoveryService{httpClient: &http.Client{
		// TODO add tls config
		// TODO !!!!!!!remove InsecureSkipVerify after configure tls for http client
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint: gosec
		}}}
}

// GetEndpoints discover endpoints from domain
func (ds *DiscoveryService) GetEndpoints(domain string) ([]*endpoint.Endpoint, error) {
	configData, err := ds.getConsortium(domain, "consortium")
	if err != nil {
		return nil, err
	}

	var endpoints []*endpoint.Endpoint

	for _, s := range configData.Config.Stakeholders {
		stakeholderConfig, err := ds.getStakeholder(s.Domain, s.Domain)
		if err != nil {
			return nil, err
		}

		for _, ep := range stakeholderConfig.Config.Endpoints {
			endpoints = append(endpoints, &endpoint.Endpoint{
				URL:    ep,
				Domain: s.Domain,
			})
		}
	}

	return endpoints, nil
}

const consortiumURLInfix = "/.well-known/did-trustbloc/"
const consortiumURLSuffix = ".json"

func configURL(urlDomain, consortiumDomain string) string {
	prefix := ""
	if !strings.HasPrefix(urlDomain, "http://") && !strings.HasPrefix(urlDomain, "https://") {
		prefix = "https://"
	}

	return prefix + urlDomain + consortiumURLInfix + consortiumDomain + consortiumURLSuffix
}

// getConsortiumFileData fetches and parses the consortium file at the given domain
func (ds *DiscoveryService) getConsortium(url, domain string) (*config.ConsortiumFileData, error) {
	res, err := ds.httpClient.Get(configURL(url, domain))
	if err != nil {
		return nil, err
	}

	// nolint: errcheck
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		// TODO retry
		return nil, fmt.Errorf("consortium config request failed: error %d, `%s`", res.StatusCode, string(body))
	}

	return config.ParseConsortium(body)
}

// getStakeholder fetches and parses the stakeholder file at the given domain
func (ds *DiscoveryService) getStakeholder(url, domain string) (*config.StakeholderFileData, error) {
	res, err := ds.httpClient.Get(configURL(url, domain))
	if err != nil {
		return nil, err
	}

	// nolint: errcheck
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		// TODO retry
		return nil, fmt.Errorf("stakeholder config request failed: error %d, `%s`", res.StatusCode, string(body))
	}

	return config.ParseStakeholder(body)
}

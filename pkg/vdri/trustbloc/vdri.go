/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package trustbloc

import (
	"errors"
	"fmt"
	"strings"

	docdid "github.com/hyperledger/aries-framework-go/pkg/doc/did"
	vdriapi "github.com/hyperledger/aries-framework-go/pkg/framework/aries/api/vdri"
	"github.com/hyperledger/aries-framework-go/pkg/vdri/httpbinding"

	"github.com/trustbloc/trustbloc-did-method/pkg/vdri/trustbloc/discovery/staticdiscovery"
	"github.com/trustbloc/trustbloc-did-method/pkg/vdri/trustbloc/endpoint"
	"github.com/trustbloc/trustbloc-did-method/pkg/vdri/trustbloc/selection/staticselection"
)

type discovery interface {
	GetEndpoints(domain string) ([]*endpoint.Endpoint, error)
}

type selection interface {
	SelectEndpoints(endpoints []*endpoint.Endpoint) ([]*endpoint.Endpoint, error)
}

type vdri interface {
	Build(pubKey *vdriapi.PubKey, opts ...vdriapi.DocOpts) (*docdid.Doc, error)
	Read(did string, opts ...vdriapi.ResolveOpts) (*docdid.Doc, error)
}

// VDRI bloc
type VDRI struct {
	resolverURL string
	discovery   discovery
	selection   selection
	getHTTPVDRI func(url string) (vdri, error) // needed for unit test
}

// New creates new bloc vdri
func New(opts ...Option) *VDRI {
	vdri := &VDRI{discovery: staticdiscovery.NewService(), selection: staticselection.NewService(),
		getHTTPVDRI: func(url string) (vdri, error) {
			return httpbinding.New(url)
		}}

	for _, opt := range opts {
		opt(vdri)
	}

	return vdri
}

// Accept did method
func (v *VDRI) Accept(method string) bool {
	return method == "trustbloc"
}

// Close vdri
func (v *VDRI) Close() error {
	return nil
}

// Store did doc
func (v *VDRI) Store(doc *docdid.Doc, by *[]vdriapi.ModifiedBy) error {
	return nil
}

// Build did doc
func (v *VDRI) Build(pubKey *vdriapi.PubKey, opts ...vdriapi.DocOpts) (*docdid.Doc, error) {
	return nil, fmt.Errorf("build method not supported for did bloc")
}

func (v *VDRI) getEndpoints(domain string) ([]*endpoint.Endpoint, error) {
	endpoints, err := v.discovery.GetEndpoints(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to discover endpoints: %w", err)
	}

	selectedEndpoints, err := v.selection.SelectEndpoints(endpoints)
	if err != nil {
		return nil, fmt.Errorf("failed to select endpoints: %w", err)
	}

	if len(selectedEndpoints) == 0 {
		return nil, errors.New("list of endpoints is empty")
	}

	return selectedEndpoints, nil
}

func (v *VDRI) Read(did string, opts ...vdriapi.ResolveOpts) (*docdid.Doc, error) {
	if v.resolverURL != "" {
		resolver, err := v.getHTTPVDRI(v.resolverURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create new sidetree vdri: %w", err)
		}

		doc, err := resolver.Read(did, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve did: %w", err)
		}

		return doc, nil
	}

	// parse did
	didParts := strings.Split(did, ":")
	if len(didParts) != 4 {
		return nil, fmt.Errorf("wrong did %s", did)
	}

	endpoints, err := v.getEndpoints(didParts[2])
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoints: %w", err)
	}

	var doc *docdid.Doc

	for _, e := range endpoints {
		sideTreeVDRI, err := v.getHTTPVDRI(e.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to create new sidetree vdri: %w", err)
		}

		resp, err := sideTreeVDRI.Read(did, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve did: %w", err)
		}

		// TODO add logic to compare response from each endpoint
		doc = resp
	}

	return doc, nil
}

// Option configures the bloc vdri
type Option func(opts *VDRI)

// WithResolverURL option is setting resolver url
func WithResolverURL(resolverURL string) Option {
	return func(opts *VDRI) {
		opts.resolverURL = resolverURL
	}
}

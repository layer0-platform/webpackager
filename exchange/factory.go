// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package exchange

import (
	"errors"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/WICG/webpackage/go/signedexchange"
	"github.com/layer0-platform/webpackager/certchain/certchainutil"
)

// Factory produces and verifies signed exchanges.
type Factory struct {
	Config
}

// FactoryProvider provides Factory.
type FactoryProvider interface {
	Get() (*Factory, error)
}

// NewFactory creates and initializes a new Factory. It panics if c.CertChain
// or c.PrivateKey is nil.
func NewFactory(c Config) *Factory {
	c.populateDefaults()
	return &Factory{c}
}

// NewExchange generates a signed exchange from resp, vp, and validityURL.
func (fty *Factory) NewExchange(resp *Response, vp ValidPeriod, validityURL *url.URL) (*signedexchange.Exchange, error) {
	u := resp.Request.URL

	e := signedexchange.NewExchange(
		fty.Version,
		u.String(),
		resp.Request.Method,
		resp.Request.Header,
		resp.StatusCode,
		resp.GetFullHeader(fty.Config.KeepNonSXGPreloads),
		resp.Payload)
	if err := e.MiEncodePayload(fty.MIRecordSize); err != nil {
		return nil, err
	}

	signer := &signedexchange.Signer{
		Date:        vp.Date(),
		Expires:     vp.Expires(),
		Certs:       fty.CertChain.Certs,
		CertUrl:     u.ResolveReference(fty.CertURL),
		ValidityUrl: validityURL,
		PrivKey:     fty.PrivateKey,
	}
	if err := e.AddSignatureHeader(signer); err != nil {
		return nil, err
	}

	return e, nil
}

// Verify validates the provided signed exchange e at the provided date.
// It returns the payload decoded from e on success.
func (fty *Factory) Verify(e *signedexchange.Exchange, date time.Time) ([]byte, error) {
	var logText strings.Builder

	payload, ok := e.Verify(
		date,
		certchainutil.WrapToCertFetcher(fty.CertChain),
		log.New(&logText, "", 0))
	if !ok {
		return nil, errors.New(logText.String())
	}

	return payload, nil
}

// Get returns fty. It implements FactoryProvider and allows Factory to be
// set directory to ExchangeFactory in webpackager.Config.
func (fty *Factory) Get() (*Factory, error) { return fty, nil }

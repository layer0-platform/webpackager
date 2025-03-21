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

package validity_test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/layer0-platform/webpackager/exchange"
	"github.com/layer0-platform/webpackager/exchange/exchangetest"
	"github.com/layer0-platform/webpackager/validity"
)

func TestAppendExtDotUnixTime(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		header http.Header
		rule   validity.URLRule
		vp     exchange.ValidPeriod
		want   string
	}{
		{
			name: "LastModified_Correct",
			url:  "https://example.com/index.html",
			header: http.Header{
				"Last-Modified": []string{"Mon, 01 Jul 2019 12:34:56 GMT"},
				"Content-Type":  []string{"text/html; charset=utf-8"},
			},
			rule: validity.AppendExtDotLastModified(".validity"),
			vp: exchange.NewValidPeriodWithLifetime(
				time.Unix(1561939200, 0), 24*time.Hour),
			want: "https://example.com/index.html.validity.1561984496",
		},
		{
			name: "LastModified_Missing",
			url:  "https://example.com/index.html",
			header: http.Header{
				"Content-Type": []string{"text/html; charset=utf-8"},
			},
			rule: validity.AppendExtDotLastModified(".validity"),
			vp: exchange.NewValidPeriodWithLifetime(
				time.Unix(1561939200, 0), 24*time.Hour),
			want: "https://example.com/index.html.validity.1561939200",
		},
		{
			name: "LastModified_Invalid",
			url:  "https://example.com/index.html",
			header: http.Header{
				"Last-Modified": []string{"COMPLETELY_BROKEN_DATE_STRING"},
				"Content-Type":  []string{"text/html; charset=utf-8"},
			},
			rule: validity.AppendExtDotLastModified(".validity"),
			vp: exchange.NewValidPeriodWithLifetime(
				time.Unix(1561939200, 0), 24*time.Hour),
			want: "https://example.com/index.html.validity.1561939200",
		},
		{
			name: "LastModified_QueryGone",
			url:  "https://example.com/index.php?id=42",
			header: http.Header{
				"Last-Modified": []string{"Mon, 01 Jul 2019 12:34:56 GMT"},
				"Content-Type":  []string{"text/html; charset=utf-8"},
			},
			rule: validity.AppendExtDotLastModified(".validity"),
			vp: exchange.NewValidPeriodWithLifetime(
				time.Unix(1561939200, 0), 24*time.Hour),
			want: "https://example.com/index.php.validity.1561984496",
		},
		{
			name: "ExchangeDate",
			url:  "https://example.com/index.html",
			header: http.Header{
				"Last-Modified": []string{"Mon, 01 Jul 2019 12:34:56 GMT"},
				"Content-Type":  []string{"text/html; charset=utf-8"},
			},
			rule: validity.AppendExtDotExchangeDate(".validity"),
			vp: exchange.NewValidPeriodWithLifetime(
				time.Unix(1561939200, 0), 24*time.Hour),
			want: "https://example.com/index.html.validity.1561939200",
		},
	}

	for _, test := range tests {
		arg, err := url.Parse(test.url)
		if err != nil {
			t.Fatal(err)
		}
		resp := exchangetest.MakeEmptyResponse(test.url)
		resp.Header = test.header
		got, err := test.rule.Apply(arg, resp, test.vp)
		if err != nil {
			t.Fatalf("got error(%q), want success", err)
		}
		if got.String() != test.want {
			t.Errorf("got %q, want %q", got, test.want)
		}
	}
}

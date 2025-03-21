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

package htmlproc_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/layer0-platform/webpackager/exchange"
	"github.com/layer0-platform/webpackager/exchange/exchangetest"
	"github.com/layer0-platform/webpackager/processor/htmlproc"
	"github.com/layer0-platform/webpackager/processor/htmlproc/htmldoc"
	"github.com/layer0-platform/webpackager/processor/htmlproc/htmltask"
	"github.com/layer0-platform/webpackager/resource/preload"
	"github.com/layer0-platform/webpackager/resource/preload/preloadtest"
)

func makeResponse(url, html string) *exchange.Response {
	resp := fmt.Sprint(
		"HTTP/1.1 200 OK\r\n",
		"Cache-Control: public, max-age=604800\r\n",
		"Content-Length: ", len(html), "\r\n",
		"Content-Type: text/html;charset=utf-8\r\n",
		"\r\n",
		html)
	return exchangetest.MakeResponse(url, resp)
}

func TestHTMLProcessor_Presets(t *testing.T) {
	pl := preloadtest.NewPreloadForRawLink

	tests := []struct {
		name  string
		html  string
		tasks []htmltask.HTMLTask
		want  []*preload.Preload
	}{
		{
			name: "ConservativeTaskSet",
			html: fmt.Sprint(
				`<!doctype html>`,
				`<link rel="preload" href="icons.svg" as="image">`,
				`<link rel="stylesheet" href="style.css">`,
				`<script src="script.js"></script>`,
			),
			tasks: htmltask.ConservativeTaskSet,
			want: []*preload.Preload{
				pl(`<https://example.com/icons.svg>;rel="preload";as="image"`),
			},
		},
		{
			name: "AggressiveTaskSet",
			html: fmt.Sprint(
				`<!doctype html>`,
				`<link rel="preload" href="icons.svg" as="image">`,
				`<link rel="stylesheet" href="style.css">`,
				`<script src="script.js"></script>`,
			),
			tasks: htmltask.AggressiveTaskSet,
			want: []*preload.Preload{
				pl(`<https://example.com/icons.svg>;rel="preload";as="image"`),
				pl(`<https://example.com/style.css>;rel="preload";as="style"`),
				pl(`<https://example.com/script.js>;rel="preload";as="script"`),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			proc := htmlproc.NewHTMLProcessor(htmlproc.Config{
				TaskSet: test.tasks,
			})
			resp := makeResponse("https://example.com/test.html", test.html)

			if err := proc.Process(resp); err != nil {
				t.Errorf("got error(%v), want success", err)
			}

			if diff := cmp.Diff(test.want, resp.Preloads); diff != "" {
				t.Errorf("resp.Preloads mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHTMLProcessor_Success(t *testing.T) {
	called := ""

	proc := htmlproc.NewHTMLProcessor(htmlproc.Config{
		TaskSet: []htmltask.HTMLTask{
			AsHTMLTask(func(*htmldoc.HTMLResponse) error {
				called += "Task1;"
				return nil
			}),
			AsHTMLTask(func(*htmldoc.HTMLResponse) error {
				called += "Task2;"
				return nil
			}),
			AsHTMLTask(func(*htmldoc.HTMLResponse) error {
				called += "Task3;"
				return nil
			}),
		},
	})
	html := `<!doctype html><p>Hello, world.</p>`
	resp := makeResponse("https://example.com/test.html", html)

	if err := proc.Process(resp); err != nil {
		t.Errorf("got error(%q), want success", err)
	}
	if called != "Task1;Task2;Task3;" {
		t.Errorf("called = %q, want %q", called, "Task1;Task2;Task3;")
	}
}

func TestHTMLProcessor_Error(t *testing.T) {
	errDummy := errors.New("errDummy")
	called := ""

	proc := htmlproc.NewHTMLProcessor(htmlproc.Config{
		TaskSet: []htmltask.HTMLTask{
			AsHTMLTask(func(*htmldoc.HTMLResponse) error {
				called += "Task1;"
				return nil
			}),
			AsHTMLTask(func(*htmldoc.HTMLResponse) error {
				called += "Task2;"
				return errDummy
			}),
			AsHTMLTask(func(*htmldoc.HTMLResponse) error {
				called += "Task3;"
				return nil
			}),
		},
	})
	html := `<!doctype html><p>Hello, world.</p>`
	resp := makeResponse("https://example.com/test.html", html)

	if err := proc.Process(resp); err != errDummy {
		if err != nil {
			t.Errorf("got error(%q), want errDummy", err)
		} else {
			t.Error("got success, want errDummy")
		}
	}
	if called != "Task1;Task2;" {
		t.Errorf("called = %q, want %q", called, "Task1;Task2;")
	}
}

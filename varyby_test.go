package throttled_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/humans-group/throttled/v2"
)

func TestVaryBy(t *testing.T) {
	u, err := url.Parse("http://localhost/test/path?q=s")
	if err != nil {
		panic(err)
	}
	ck := &http.Cookie{Name: "ssn", Value: "test"}
	cases := []struct {
		vb *throttled.VaryBy
		r  *http.Request
		k  string
	}{
		0: {nil, &http.Request{}, ""},
		1: {
			&throttled.VaryBy{RemoteAddr: true},
			&http.Request{RemoteAddr: "1.2.3.4:1234"}, // IPv4
			"1.2.3.4\n",
		},
		2: {
			&throttled.VaryBy{RemoteAddr: true},
			&http.Request{RemoteAddr: "[::]:1234"}, // IPv6
			"[::]\n",
		},
		3: {
			// Don't panic in case RemoteAddr doesn't contain a port.
			&throttled.VaryBy{RemoteAddr: true},
			&http.Request{RemoteAddr: "1.2.3.4"},
			"1.2.3.4\n",
		},
		4: {
			&throttled.VaryBy{Method: true, Path: true},
			&http.Request{Method: "POST", URL: u},
			"post\n/test/path\n",
		},
		5: {
			&throttled.VaryBy{Headers: []string{"Content-length"}},
			&http.Request{Header: http.Header{"Content-Type": []string{"text/plain"}, "Content-Length": []string{"123"}}},
			"123\n",
		},
		6: {
			&throttled.VaryBy{Separator: ",", Method: true, Headers: []string{"Content-length"}, Params: []string{"q", "user"}},
			&http.Request{Method: "GET", Header: http.Header{"Content-Type": []string{"text/plain"}, "Content-Length": []string{"123"}}, Form: url.Values{"q": []string{"s"}, "pwd": []string{"secret"}, "user": []string{"test"}}},
			"get,123,s,test,",
		},
		7: {
			&throttled.VaryBy{Cookies: []string{"ssn"}},
			&http.Request{Header: http.Header{"Cookie": []string{ck.String()}}},
			"test\n",
		},
		8: {
			&throttled.VaryBy{Cookies: []string{"ssn"}, RemoteAddr: true, Custom: func(r *http.Request) string {
				return "blah"
			}},
			&http.Request{Header: http.Header{"Cookie": []string{ck.String()}}},
			"blah",
		},
	}
	for i, c := range cases {
		got := c.vb.Key(c.r)
		if got != c.k {
			t.Errorf("%d: expected '%s' (%d), got '%s' (%d)", i, c.k, len(c.k), got, len(got))
		}
	}
}

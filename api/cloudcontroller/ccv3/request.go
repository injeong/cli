package ccv3

import (
	"io"
	"net/http"
	"net/url"
)

// requestOptions contains all the options to create an HTTP request.
type requestOptions struct {
	// URIParams are the list URI route parameters
	URIParams map[string]string

	// Query is a list of HTTP query parameters. Query will overwrite any
	// existing query string in the URI. If you want to preserve the query
	// string in URI make sure Query is nil.
	Query url.Values

	// RequestName is the name of the request (see routes)
	RequestName string

	// URL is the request path.
	URL string
	// Method is the HTTP method.
	Method string
	// Body is the content of the request.
	Body io.Reader
}

// newHTTPRequest returns a constructed HTTP.Request with some defaults.
// Defaults are applied when Request options are not filled in.
func (client *Client) newHTTPRequest(passedRequest requestOptions) (*http.Request, error) {
	var request *http.Request
	var err error

	if passedRequest.URL != "" {
		request, err = http.NewRequest(
			passedRequest.Method,
			passedRequest.URL,
			passedRequest.Body,
		)
	} else {
		request, err = client.router.CreateRequest(
			passedRequest.RequestName,
			passedRequest.URIParams,
			passedRequest.Body,
		)
	}
	if err != nil {
		return nil, err
	}

	if passedRequest.Query != nil {
		request.URL.RawQuery = passedRequest.Query.Encode()
	}

	request.Header = http.Header{}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", client.userAgent)

	return request, nil
}

// soaptrip is an HTTP Transport wrapper for parsing SOAP Faults
package soaptrip

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var (
	// DefaultSoapTrip is a new SoapRoundTripper using http.DefaultTransport
	DefaultSoapTrip = New(http.DefaultTransport)
)

// New returns a new SoapRoundTripper from an existing http.RoundTripper
func New(rt http.RoundTripper) http.RoundTripper {
	return &SoapRoundTripper{rt}
}

// SoapRoundTripper is a wrapper for an existing http.RoundTripper
type SoapRoundTripper struct {
	rt http.RoundTripper
}

// RoundTrip will call the original http.RoundTripper. Upon an error of the original RoundTripper it will return,
// otherwise it will copy the response body and attempt to parse it for a fault. If one is found a SoapFault
// will be returned as the error.
func (st *SoapRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// call original round tripper
	resp, err := st.rt.RoundTrip(req)

	// return on error
	if err != nil {
		return nil, err
	}

	// parse resp for soap faults
	err = ParseFault(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// SoapFault is  an implementation of error for soap faults. Due to a change in Go 1.4, a non nil http.Response and
// error can not be returned at the same time. SoapFault contains the original http.Response so that it can be read.
type SoapFault struct {
	FaultCode   string
	FaultString string
	Response    *http.Response
}

func (sf SoapFault) Error() string {
	return fmt.Sprintf("FaultCode: '%s' FaultString: '%s'", sf.FaultCode, sf.FaultString)
}

// ParseFault attempts to parse a Soap Fault from an http.Response. If a fault is found, it will return an error
// of type SoapFault, otherwise it will return nil
func ParseFault(resp *http.Response) error {
	var buf bytes.Buffer
	d := xml.NewDecoder(io.TeeReader(resp.Body, &buf))

	var start xml.StartElement
	fault := &SoapFault{Response: resp}
	found := false
	depth := 0

	// iterate through the tokens
	for {
		tok, _ := d.Token()
		if tok == nil {
			break
		}

		// switch on token type
		switch t := tok.(type) {
		case xml.StartElement:
			start = t.Copy()
			depth++
			if depth > 2 { // don't descend beyond Envelope>Body>Fault
				break
			}
		case xml.EndElement:
			start = xml.StartElement{}
			depth--
		case xml.CharData:
			// fault was found, capture the values and mark as found
			switch strings.ToLower(start.Name.Local) {
			case "faultcode":
				found = true
				fault.FaultCode = string(t)
			case "faultstring":
				found = true
				fault.FaultString = string(t)
			}
		}
	}

	resp.Body = struct {
		io.Reader
		io.Closer
	}{io.MultiReader(bytes.NewReader(buf.Bytes()), resp.Body), resp.Body}

	if found {
		fault.Response = resp
		return fault
	}

	return nil
}

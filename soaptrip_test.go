package soaptrip

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	validFault = `<env:Envelope xmlns:env="http://schemas.xmlsoap.org/soap/envelope/">
	<env:Body>
		<env:Fault>
			<faultcode>fcode</faultcode>
			<faultstring>valid fault string</faultstring>
		</env:Fault>
	</env:Body>
</env:Envelope>`

	validResp = `<env:Envelope xmlns:env="http://schemas.xmlsoap.org/soap/envelope/">
	<env:Body>
		<foo />
	</env:Body>
</env:Envelope>`
)

// staticTransport is a http.RoundTripper that returns a static
// http.Response and error that are defined within it
type staticTransport struct {
	resp *http.Response
	err  error
}

// RoundTrip overwrites the most likely erroneous http.Response with a static one from staticTransport
func (st *staticTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return st.resp, st.err
}

func TestNew(t *testing.T) {
	Convey("Given an http.Client with a SoapTrip transport that wraps a static response transport", t, func() {
		c := &http.Client{}
		st := &staticTransport{}
		c.Transport = New(st)

		Convey("Given a nil response and non nil error of 'foo'", func() {
			st.err = errors.New("foo")

			Convey("A post call should return the original error from staticTranport", func() {
				_, err := c.Post("http://localhost:99999", "text/xml", strings.NewReader("soap envelope here"))
				So(err.Error(), ShouldEqual, "Post http://localhost:99999: foo")
			})
		})

		Convey("Given a response with a valid SOAP fault and nil error", func() {
			st.resp = &http.Response{}
			st.resp.Body = ioutil.NopCloser(strings.NewReader(validFault))

			Convey("A post call should return the new SoapFault from SoapRoundTripper", func() {
				_, err := c.Post("http://localhost:99999", "text/xml", strings.NewReader("soap envelope here"))
				So(err.Error(), ShouldEqual, "Post http://localhost:99999: FaultCode: 'fcode' FaultString: 'valid fault string'")

				Convey("The original http.Response should be available", func() {
					So(err.(*url.Error).Err.(*SoapFault).Response, ShouldNotBeNil)
				})
			})
		})

		Convey("Given a response with a valid soap envelope and no fault", func() {
			st.resp = &http.Response{}
			st.resp.Body = ioutil.NopCloser(strings.NewReader(validResp))

			Convey("A post call should return a nil error", func() {
				_, err := c.Post("http://localhost:99999", "text/xml", strings.NewReader("soap envelope here"))
				So(err, ShouldBeNil)
			})

			Convey("A post call should return a valid response", func() {
				resp, _ := c.Post("http://localhost:99999", "text/xml", strings.NewReader("soap envelope here"))
				b, _ := ioutil.ReadAll(resp.Body)
				So(string(b), ShouldEqual, validResp)
			})
		})
	})
}

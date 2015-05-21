soaptrip [![Build Status](https://travis-ci.org/kylewolfe/soaptrip.svg?branch=master)](https://travis-ci.org/kylewolfe/soaptrip) [![Coverage Status](https://coveralls.io/repos/kylewolfe/soaptrip/badge.svg)](https://coveralls.io/r/kylewolfe/soaptrip) [![GoDoc](http://godoc.org/github.com/kylewolfe/soaptrip?status.svg)](http://godoc.org/github.com/kylewolfe/soaptrip) 
=========

soaptrip is an HTTP Transport wrapper for parsing SOAP Faults

## Usage

```go
// create a client
c := &http.Client{Transport: soaptrip.DefaultSoapTrip}

// make a call that can respond with a SOAP Fault
resp, err := c.Post("http://localhost/soap/login", "text/xml", soapEnvelopeReader)
if err != nil {
	log.Println(err.Error()) // Post http://localhost/soap/login: FaultCode: 'faultcode' FaultString: 'fault string'

	// resp will always == nil if err != nil, however, you can retrieve the original resp
	urlError := err.(*url.Error)
	switch urlError.Err.(type) {
	case *soaptrip.SoapFault:
		resp = urlError.Err.Response
	default:
		// this is not a soap fault error
	}
}
```
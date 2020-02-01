package requests

import (
	"errors"
	"github.com/alessiosavi/Requests/datastructure"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Remove comment for set the log at debug level
var req Request // = InitDebugRequest()

func TestCreateHeaderList(t *testing.T) {
	// Create a simple headers
	headersKey := `Content-Type`
	headersValue := `application/json`

	err := req.CreateHeaderList(headersKey, headersValue)
	if err == nil {
		t.Error("Error, request is not initialized!")
	}

	request, err := InitRequest("http://", "POST", nil, false, true)
	if err != nil {
		t.Error("Error!: ", err)
	}

	err = request.CreateHeaderList(headersKey, headersValue)
	if err != nil {
		t.Error("Error!", err)
	}
	if strings.Compare(request.Req.Header.Get(headersKey), headersValue) != 0 {
		t.Error("Headers key mismatch!")
	}

}

func TestSendRequest(t *testing.T) {
	var resp *datastructure.Response
	resp = makeBadRequestURL1()
	if resp == nil || resp.Error == nil {
		t.Fail()
	} else {
		t.Log("makeBadRequestURL1 Passed!")
	}
	// t.Log(resp.Dump())

	resp = makeBadRequestURL2()
	if resp == nil || resp.Error == nil {
		t.Fail()
	} else {
		t.Log("makeBadRequestURL2 Passed!")
	}
	// t.Log(resp.Dump())

	resp = makeOKRequestURL3()
	if resp == nil || resp.Error != nil || resp.StatusCode != 200 {
		t.Fail()
	} else {
		t.Log("makeOKRequestURL3 Passed!")
	}
	// t.Log(resp.Dump())
}

func BenchmarkRequestGETWithoutTLS(t *testing.B) {
	var r Request
	for i := 0; i < t.N; i++ {
		r.SendRequest("http://127.0.0.1:9999", "GET", nil, false)
	}
}

func BenchmarkRequestPOSTWithoutTLS(t *testing.B) {
	var r Request
	for i := 0; i < t.N; i++ {
		r.SendRequest("http://127.0.0.1:9999", "POST", []byte{}, false)
	}
}

func BenchmarkParallelRequestGETWithoutTLS(t *testing.B) {
	var n int = t.N
	var requests []Request = make([]Request, n)
	for i := 0; i < n; i++ {
		req, err := InitRequest("http://127.0.0.1:9999", "GET", nil, true, false)
		if err == nil && req != nil {
			requests[i] = *req
		} else if err != nil {
			t.Error("error: ", err)
		}
	}
	for i := 0; i < t.N; i++ {
		ParallelRequest(requests, 100)
	}
}

func BenchmarkParallelRequestPOSTWithoutTLS(t *testing.B) {
	var n int = t.N
	var requests []Request = make([]Request, n)
	for i := 0; i < n; i++ {
		req, err := InitRequest("http://127.0.0.1:9999", "POST", []byte{}, true, false)
		if err == nil && req != nil {
			requests[i] = *req
		} else if err != nil {
			t.Error("error: ", err)
		}
	}
	for i := 0; i < t.N; i++ {
		ParallelRequest(requests, 100)
	}
}

func makeBadRequestURL1() *datastructure.Response {
	return req.SendRequest("tcp://google.it", "GET", nil, true)
}
func makeBadRequestURL2() *datastructure.Response {
	return req.SendRequest("google.it", "GET", nil, true)
}
func makeOKRequestURL3() *datastructure.Response {
	return req.SendRequest("https://google.it", "GET", nil, true)
}

type headerTestCase struct {
	input    []string
	expected bool
	number   int
}

func TestRequest_CreateHeaderList(t *testing.T) {
	var request *Request
	request, err := InitRequest("http://", "POST", nil, false, true)
	if err != nil {
		t.Error("Error!", err)
	}
	cases := []headerTestCase{
		{input: []string{"Content-Type", "text/plain"}, expected: true, number: 1},
		{input: []string{"Content-Type"}, expected: false, number: 2},
		{input: []string{"Content-Type", "text/plain", "Error"}, expected: false, number: 3},
	}
	for _, c := range cases {
		err := request.CreateHeaderList(c.input...)
		if (c.expected && err != nil) || (!c.expected && err == nil) {
			t.Errorf("Expected %v for input %v [test n. %d]", c.expected, c.input, c.number)
		}
	}
}

type requestTestCase struct {
	host     string
	method   string
	body     []byte
	skipTLS  bool
	expected error
	number   int
}

func TestRequest_SendRequest(t *testing.T) {
	var request Request

	// create a listener with the desired port.
	l, err := net.Listen("tcp", "127.0.0.1:8081")
	if err != nil {
		log.Fatal(err)
	}

	ts := httptest.NewUnstartedServer(nil)
	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	_ = ts.Listener.Close()
	ts.Listener = l

	// Start the server.
	ts.Start()

	cases := []requestTestCase{

		// GET
		{host: "http://localhost:8081/", method: "GET", body: nil, skipTLS: false, expected: nil, number: 1},
		{host: "http://localhost:8081/", method: "GET", body: nil, skipTLS: true, expected: nil, number: 2},
		{host: "localhost:8081/", method: "GET", body: nil, skipTLS: false, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 3},
		// POST
		{host: "localhost:8081/", method: "POST", body: []byte{}, skipTLS: true, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 4},
		{host: "localhost:8081/", method: "POST", body: nil, skipTLS: true, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 5},
		{host: "http://localhost:8081/", method: "HEAD", body: nil, skipTLS: false, expected: errors.New("HTTP_METHOD_NOT_MANAGED"), number: 6},
		{host: "http://localhost:8080/", method: "GET", body: nil, skipTLS: false, expected: errors.New("ERROR_SENDING_REQUEST"), number: 7},
	}

	for _, c := range cases {
		resp := request.SendRequest(c.host, c.method, c.body, c.skipTLS)
		if c.expected != resp.Error {
			if c.expected != nil && resp.Error != nil {
				if !strings.Contains(resp.Error.Error(), c.expected.Error()) {
					t.Errorf("Expected %v, received %v [test n. %d]", c.expected, resp.Error, c.number)
				}
			} else {
				t.Error("Url not reachable! Spawn a simple server (python3 -m http.server 8081 || python -m SimpleHTTPServer 8081)")
			}
		}
	}

	// Cleanup.
	ts.Close()
}

func
TestRequest_InitRequest(t *testing.T) {

	cases := []requestTestCase{

		// GET
		{host: "http://localhost:8081/", method: "GET", body: nil, skipTLS: false, expected: nil, number: 1},
		{host: "http://localhost:8081/", method: "GET", body: nil, skipTLS: true, expected: nil, number: 2},
		{host: "localhost:8081/", method: "GET", body: nil, skipTLS: false, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 3},
		// POST
		{host: "localhost:8081/", method: "POST", body: []byte{}, skipTLS: true, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 4},
		{host: "localhost:8081/", method: "POST", body: nil, skipTLS: true, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 5},
		{host: "http://localhost:8081/", method: "HEAD", body: nil, skipTLS: false, expected: errors.New("HTTP_METHOD_NOT_MANAGED"), number: 6},
	}

	for _, c := range cases {
		_, err := InitRequest(c.host, c.method, c.body, c.skipTLS, false)
		if c.expected != err {
			if c.expected.Error() != err.Error() {
				t.Errorf("Expected %v, received %v [test n. %d]", c.expected, err.Error(), c.number)
			}
		}
	}
}

func
TestRequest_ExecuteRequest(t *testing.T) {
	// create a listener with the desired port.
	l, err := net.Listen("tcp", "127.0.0.1:8081")
	if err != nil {
		log.Fatal(err)
	}

	ts := httptest.NewUnstartedServer(nil)
	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	ts.Listener.Close()
	ts.Listener = l

	// Start the server.
	ts.Start()

	cases := []requestTestCase{
		// GET
		{host: "http://localhost:8081/", method: "GET", body: nil, skipTLS: false, expected: nil, number: 1},
		{host: "http://localhost:8081/", method: "GET", body: nil, skipTLS: true, expected: nil, number: 2},
		{host: "localhost:8081/", method: "GET", body: nil, skipTLS: false, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 3},
		// POST
		{host: "localhost:8081/", method: "POST", body: []byte{}, skipTLS: true, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 4},
		{host: "localhost:8081/", method: "POST", body: nil, skipTLS: true, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 5},
		{host: "http://localhost:8081/", method: "HEAD", body: nil, skipTLS: false, expected: errors.New("HTTP_METHOD_NOT_MANAGED"), number: 6},
		{host: "http://localhost:8080/", method: "GET", body: nil, skipTLS: false, expected: errors.New("ERROR_SENDING_REQUEST"), number: 7},
	}

	client := &http.Client{}
	for _, c := range cases {
		req, err := InitRequest(c.host, c.method, c.body, c.skipTLS, false)
		if err == nil {
			resp := req.ExecuteRequest(client)

			if c.expected != nil && resp.Error != nil {
				if !strings.Contains(resp.Error.Error(), c.expected.Error()) {
					t.Errorf("Expected %v, received %v [test n. %d]", c.expected, resp.Error, c.number)
				}
			}
		}
	}
	// Cleanup.
	ts.Close()
}

type timeoutTestCase struct {
	host    string
	method  string
	body    []byte
	skipTLS bool
	time    int
	number  int
}

func TestRequest_Timeout(t *testing.T) {
	// Need to run the server present in example/server_example.py
	cases := []timeoutTestCase{
		// GET
		{host: "https://localhost:5000/timeout", method: "GET", body: nil, skipTLS: true, time: 11, number: 1},
	}

	for _, c := range cases {
		var req Request // = InitDebugRequest()
		req.SetTimeout(time.Second * time.Duration(c.time))
		start := time.Now()
		resp := req.SendRequest(c.host, c.method, c.body, c.skipTLS)
		elapsed := time.Since(start)
		if resp.Error != nil {
			t.Errorf("Received an error -> %v [test n. %d].\n Be sure that the python server on ./example folder is up and running", resp.Error, c.number)
		}
		if time.Duration(c.time)*time.Second < elapsed {
			t.Error("Error timeout")
		}
	}
}

func TestParallelRequest(t *testing.T) {
	start := time.Now()
	// This array will contains the list of request
	var reqs []Request
	// This array will contains the response from the given request
	var response []datastructure.Response

	// Set to run at max 100 request in parallel (use CPU count for best effort)
	var N = 10
	// Create the list of request
	for i := 0; i < 100; i++ {
		// Run against the `server_example.py` present in this folder
		req, err := InitRequest("https://127.0.0.1:5000", "GET", nil, true, false) // Alternate cert validation
		if err != nil {
			t.Error("Error request [", i, "]. Error: ", err)
		} else {
			req.SetTimeout(1000 * time.Millisecond)
			reqs = append(reqs, *req)
		}
	}

	// Run the request in parallel
	response = ParallelRequest(reqs, N)

	elapsed := time.Since(start)

	for i := range response {
		if response[i].Error != nil {
			t.Error("Error request [", i, "]. Error: ", response[i].Error)
		}
	}
	log.Printf("Requests took %s", elapsed)
}

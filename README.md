Logging server [![Build Status](https://travis-ci.org/chlu/logging-server.png)](https://travis-ci.org/chlu/logging-server) 
==============

A logging server with simple mock handlers for quick HTTP API stubs, written in Google Go (golang).
See all requests to an HTTP API and provide simple rules to respond to requests.

Installation
------------

You need Google Go (1.0) in order to install this package (see http://golang.org/doc/install).
After installing the Go SDK it should be as easy as:

    go get github.com/chlu/logging-server

Running
-------

Start a mock server *just for logging*:

    logging-server

Start a mock server on a special address and port:

    logging-server -address 0.0.0.0 -port 3000

Use a mock configuration for rules:

    logging-server mock.json

Example mock configuration
--------------------------

    [{
      "Method": "POST",
      "Path": "^/service/test/(.*)/action",
      "Response": {
        "Body": {"value": "9999999", "arbitrary_json": ["1234"]},
        "Defer": "375ms"
      }
    }, {
      "Method": "GET",
      "Path": "/",
      "Response": {
        "ContentType": "text/html",
        "Body": "<html><body><h1>Hello world!</h1></body></html>"
      }
    }]

Limitations
-----------

The logging server is most useful for JSON APIs right now.
As you can see in the example mock configuration any text content type is also supported.

Some ideas for improvement:

* Implement handling binary data in response
* Support expressions in the response based on the request

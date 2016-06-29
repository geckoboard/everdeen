# Everdeen

Everdeen is an HTTP proxy server for mocking outgoing traffic.
She's controllable via a simple HTTP API (or a [Ruby Gem](https://rubygems.org/gems/everdeen)), has a small memory footprint, and boots in under a second.

Everdeen gets her name from [Katniss Everdeen](https://en.wikipedia.org/wiki/Katniss_Everdeen), the "**Mocking**-jay" of the Hunger Games trilogy. :books:

## Usage

Everdeen was created to mock HTTP traffic in the automated acceptance test suite at [Geckoboard](https://www.geckoboard.com/), our test suite is written in Ruby and uses RSpec; so the simplest way to use everdeen is to make use of the [Ruby Gem](https://rubygems.org/gems/everdeen):


```ruby
require 'everdeen'

# Starting the server
server = Everdeen::Server.start(
  proxy_port: 4321,
  control_port: 4322
)

# Stopping the server
server.stop
```

To use the Everdeen server standalone (for use cases where the Ruby Gem is not appropriate) simply download the [binary for your platform](https://github.com/geckoboard/everdeen/releases) and run it like so:

```
$ ./everdeen_0.1.0_linux-amd64 -proxy-addr=":4321" -control-addr=":4322"
```

When the Everdeen server is running we can proxy HTTP traffic through it like so:

```
$ http_proxy=http://127.0.0.1:4321 \
  curl http://www.geckoboard.com/
```

By default all requests going through Everdeen will be blocked, to mock responses or allow requests to reach external servers we must create "expectations" using the HTTP API.

### Expectations

Expectations are used to control the way that Everdeen responds to requests based on a set of criteria, here's an example of an expectation that responds to all `GET` requests to `www.geckoboard.com` with the message `"Hello World"`:

```ruby
everdeen.create_expectations([
  Everdeen::Expectation.new(
    request_criteria: [
      { type: :method, value: :GET },
      { type: :host, value: 'www.geckoboard.com' }
    ],
    response: {
      status: 201,
      body: 'Hello World',
      headers: {
        'X-Custom-Header' => 'Custom Header'
      }
    }
  )
])
```

To create the same expectation by interacting directly with the API you'd `POST` the following JSON to the `/expectations` endpoint:

```json
{
  "expectations": [
    {
      "request_criteria": [
        {
          "type": "method",
          "value": "GET"
        },
        {
          "type": "host",
          "value": "www.geckoboard.com"
        }
      ],
      "respond_with": {
        "status": 201,
        "body": "Hello World",
        "headers": {
          "X-Custom-Header": "Custom Header"
        }
      }
    }
  ]
}
```

#### Criteria types

Here are all the things that Everdeen can match requests based on:

- Method (exact matches only)
- Host (exact and regex matches)
- Path (exact and regex matches)
- Request Headers (exact and regex matches)
- Request Body (exact and regex matches)
- Query String Parameters (exact and regex matches)

#### Matching with regex

To create request criteria that match with regex, simply supply the `match_type` parameter:

```ruby
Everdeen::Expectation.new(
  request_criteria: [
    { type: :host, match_type: :regex, value: '.*\.geckoboard\.com' }
  ]
)
```

#### Responding with binary data

Sometimes it may be desirable to respond to a request with the contents of a binary file (e.g. an image), creating this expectation using the API may be problematic because JSON can only work with unicode characters (not arbitrary strings of bytes).

To work around this you can supply your binary data to Everdeen as Base64:

```ruby
Everdeen::Expectation.new(
  response: {
    body: Base64.encode64(binary_data),
    body_encoding: 'base64'
  }
)
```

#### Allowing requests through the proxy

Everdeen can also allow requests through the proxy unaltered, to do so simply set the `pass_through` attribute on your expectation instead of supplying a response:

```ruby
Everdeen::Expectation.new(
  pass_through: true
)
```

#### Storing matching requests

Sometimes it is useful to retrieve information about requests that have been handled by the Everdeen proxy,
one such example in the Geckoboard test suite is a test case where we assert that a request was made
to a third party API to delete information related to a user when they close their account.

When you want to store a request for future retrieval you can set the `store_matching_requests` attribute on an expectation.

```ruby
Everdeen::Expectation.new(
  store_matching_requests: true
)
```

When you call the API to create expectations it will return the expectations with UUIDs assigned, these UUIDs are important as they will be used to retrieve the requests matching each expectation.

```json
[
    {
      "uuid": "586dc37e-6421-461a-8302-e57d6cdf9e2e",
      "store_matching_requests": true,
      "request_criteria": [
        ...
      ],
      "respond_with": {
        ...
      }
    }
]
```

#### Retrieving requests for an expectation

Now that you have registered your expectation you will want to query the requests that have matched that expectation. So with your expectation UUID you can just do

```ruby
expectation = server.create_expectations(expectations).first
requests = server.requests(expectation.uuid)
```

Details of all requests matching that expectation will be returned. Note that in all cases the body returned for the request will be Base64 encoded. This ensures consistency especially if there are any requests that are binary data.

```json
{
    "body_base64": "SGVsbG8gV29ybGQ=",
    "headers": {
        "Accept": [
            "*/*"
        ],
        "Accept-Encoding": [
            "gzip;q=1.0,deflate;q=0.6,identity;q=0.3"
        ],
        "Content-Length": [
            "11"
        ],
        "Content-Type": [
            "application/x-www-form-urlencoded"
        ],
        "User-Agent": [
            "Ruby"
        ]
    },
    "method": "POST",
    "url": "https://geckoboard.com"
}
```

To retrieve the raw body content just decode with Base64 but when using the ruby gem calling the `#body` method will return the decoded body content.

```ruby
requests = server.requests(expectation.uuid)
requests.first.body
=> "Hello World"
```

#### Proxying HTTPS Traffic

Due to the secure nature of TLS; HTTPS requests can't be proxied transparently. To overcome this problem, the Everdeen proxy will act as a [Certificate Authority](https://en.wikipedia.org/wiki/Certificate_authority) and decrypt / re-encrypt traffic using it's own self-signed certificates.

Practically this means that you need to trust the Everdeen proxy's certificate to generate / sign certificates.

By default the Everdeen proxy will use the bundled [goproxy certificate](https://github.com/elazarl/goproxy/blob/52c137b4b19acaf8cde1d9e0579d928519918919/certs.go#L24-L38), which you *could* add to your operating system or browser's trust store but this is **highly** discouraged, due to the fact that the private key is available on the public internet.

##### Generating a certificate

It is recommended that you generate your own certificate / key pair like so:

```
$ ./everdeen_0.1.0_linux-amd64 -generate-ca-cert
```

This will generate a `cert.pem` and `key.pem` file in your current working directory.

:warning: Make sure you keep your `key.pem` file safe, as once you trust the `cert.pem` as a Certificate Authority the owner of this file can sign their own certificates and do very nasty things (e.g. pretend to be your bank).

If you're on Ubuntu Linux you can add your newly generated certificate to the trust store like so:

```
$ sudo cp cert.pem /usr/local/share/ca-certificates/everdeen.crt
$ sudo update-ca-certificates
```

##### Using the certificate

Now you have generated the certificate, you must tell the Everdeen proxy to use it:

```ruby
server = Everdeen::Server.start(
  ca_cert_path: '/path/to/the/cert.pem'
  ca_key_path: '/path/to/the/key.pem'
)
```

Here's how you would do so if using the standalone binary:

```
$ ./everdeen_0.1.0_linux-amd64 -ca-cert-path="/path/to/the/cert.pem" ca-key-path="/path/to/the/key.pem"
```

## Similar Projects

- [Puffing Billy] (https://github.com/oesmith/puffing-billy)
- [Mockserver] (https://github.com/jamesdbloom/mockserver)
- [Hoverfly] (https://github.com/SpectoLabs/hoverfly)

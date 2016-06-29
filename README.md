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

#### Storing matching requests

Sometimes it is useful to retrieve information about requests that have been handled by the Everdeen proxy,
one such example in the Geckoboard test suite is a test case where we assert that a request was made
to a third party API to delete information related to a user when they close their account

Therefore when you want to store a request for future retrieval you can set the store_matching_requests attribute on an expectation.

```ruby
Everdeen::Expectation.new(
  store_matching_requests: true
)
```

When you make that request to create the expectation it returns all the expectations in the original call with the uuid generated on the server side. This uuid is important as it will be required to retrieve the requests for that expectation.

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

Now that you have registered your expectation you will want to query the requests that have matched for that expectation. So with your expectation uuid you can just do

```ruby
expectation = server.create_expectations(expectations).first
requests = server.requests(expectation.uuid)
```

Details of all requests matching that expectation will be returned. Note that in all cases the body returned for the request will be base64 encoded. This ensures consistenty especially if there are any requests that are binary data.

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

To retrieve the raw body content just decode with base64 but when using the ruby gem calling the body method will return the decoded body content implicitly

```ruby
requests = server.requests(expectation.uuid)
requests.first.body
=> "Hello World"
```

#### Certificates

Everdeen permits overriding the [goproxy](https://github.com/elazarl/goproxy) certificates. By default it will use the goproxy certificate and will get some warnings. However it is recommended for security reasons to use your own generated cerifications.

##### Generating your certificates

Everdeen supports generating a new set of certificates saved to disk. To do this pass the `generate-ca-cert` to the everdeen server standalone.

```
$ ./everdeen_0.1.0_linux-amd64 -generate-ca-cert
```

That will generate a new key.pem and cert.pem - don't forget to install and trust the cert.pem so that when MITM requests come in they are trusted.

##### Using the certificates

Now you have generated the certificates put them in some directory to reference them like the below examples.

Ruby gem

```ruby
# Passing custom certs
server = Everdeen::Server.start(
  ca_cert_path: File.join('/usr','local', 'share','ca-certificates', 'everdeen.crt'),
  ca_key_path: File.join('/etc','default', 'everdeen_private', 'key.pem')
)

```

Standalone

```
$ ./everdeen_0.1.0_linux-amd64 -ca-cert-path="/usr/local/share/ca-certificates/everdeen.crt" ca-key-path="/etc/default/everdeen_private/key.pem"
```

## Similar Projects

- [Puffing Billy] (https://github.com/oesmith/puffing-billy)
- [Mockserver] (https://github.com/jamesdbloom/mockserver)
- [Hoverfly] (https://github.com/SpectoLabs/hoverfly)

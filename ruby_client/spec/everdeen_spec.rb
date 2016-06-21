require 'spec_helper'

RSpec.describe Everdeen, type: :integration do
  let(:server) { start_server }

  before { server }
  after  { server.stop }

  example 'starting the server and creating expectations' do

    expect {
      server.create_expectations([
        Everdeen::Expectation.new(
          request_criteria: [
            {
              type: 'method',
              value: 'GET'
            }
          ],

          response: {
            status: 200,
            body: 'Hello World'
          }
        )
      ])
    }.to change { server.registered_expectations.count }.from(0).to 1
  end

  example 'requesting matching processed requests' do
    expectations = server.create_expectations([
      Everdeen::Expectation.new(
        store_matching_requests: true,
        request_criteria: [
          {
            type: 'method',
            value: "POST"
          }
        ],

        response: {
          status: 200,
          body: 'Hello World'
        }
      )
    ])

    Net::HTTP.new('127.0.0.1', 4321, nil, nil).start do |http|
      request = Net::HTTP::Post.new 'https://example.com/test'
      request.body = "Hello World"
      http.request request
    end

    requests = server.requests(expectations.first.uuid)

    expect(requests.size).to eq 1
    request = requests.first

    expect(request.url).to eq('https://example.com/test')
    expect(request.headers).to eq({
      "Accept"=>["*/*"],
      "Accept-Encoding"=>["gzip;q=1.0,deflate;q=0.6,identity;q=0.3"],
      "Content-Length"=>["11"],
      "Content-Type"=>["application/x-www-form-urlencoded"],
      "User-Agent"=>["Ruby"]
    })

    expect(request.method).to eq 'POST'
    expect(request.body_base64).to eq "SGVsbG8gV29ybGQ="
  end

  def start_server
    Everdeen::Server.start(
      proxy_port: 4321,
      control_port: 4322,
      store_requests: true
    )
  end
end

module Everdeen
  class Client
    attr_reader :control_addr

    def initialize(control_addr)
      @control_addr = control_addr
    end

    def ping
      Net::HTTP.get build_uri('/ping')
    end

    def registered_expectations
      response = Net::HTTP.get build_uri('/expectations')
      JSON.parse(response)
    end

    def requests(request_criteria)
      uri = build_uri('/requests')

      request = Net::HTTP::Post.new(uri, { 'Content-Type' => 'application/json' })
      request.body = { request_criteria: request_criteria.to_hash }.to_json

      Net::HTTP.start(uri.host, uri.port) do |http|
        JSON.parse(http.request(request).body)
      end
    end

    def create_expectations(expectations)
      uri = build_uri('/expectations')

      request = Net::HTTP::Post.new(uri, { 'Content-Type' => 'application/json' })
      request.body = {
        expectations: expectations.map(&:to_hash)
      }.to_json

      Net::HTTP.start(uri.host, uri.port) do |http|
        http.request(request)
      end
    end

    private

    def build_uri(path)
      uri = control_addr.dup
      uri.path = path
      uri
    end
  end
end

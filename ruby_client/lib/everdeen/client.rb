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

    def requests(expectation_uuid)
      response = Net::HTTP.get build_uri("/expectations/#{expectation_uuid}/requests")
      JSON.parse(response)
    end

    def create_expectations(expectations)
      uri = build_uri('/expectations')

      request = Net::HTTP::Post.new(uri, { 'Content-Type' => 'application/json' })
      request.body = {
        expectations: expectations.map(&:to_hash)
      }.to_json

      Net::HTTP.start(uri.host, uri.port) do |http|
        JSON.parse(http.request(request).body)
      end
    end

    def reset_all
      uri = build_uri('/reset/all')
      req = Net::HTTP::Delete.new(uri.path)

      Net::HTTP.start(uri.host, uri.port) { |http| http.request(req) }
    end

    private

    def build_uri(path)
      uri = control_addr.dup
      uri.path = path
      uri
    end
  end
end

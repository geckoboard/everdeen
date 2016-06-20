module Everdeen
  class Expectation
    attr_reader :uuid, :max_matches, :response, :request_criteria

    def initialize(args = {})
      args.each do |key, value|
        next if ['response', 'request_criteria'].include?(key.to_s)
        instance_variable_set("@#{key}", value)
      end

      add_response(args[:response] || args['response'])
      add_request(args[:request_criteria] || args['request_criteria'])
    end

    def pass_through
      !!@pass_through
    end

    def store_matching_requests
      !!@store_matching_requests
    end

    def to_hash
      {
        store_matching_requests: store_matching_requests,
        max_matches: max_matches,
        pass_through: pass_through,
        request_criteria: request_criteria.to_hash,
        respond_with: response.to_hash
      }
    end

    private

    def add_response(response_attr)
      @response = Response.new(response_attr)
    end

    def add_request(request_attrs)
      @request_criteria = RequestCriteria.new(request_attrs)
    end
  end
end

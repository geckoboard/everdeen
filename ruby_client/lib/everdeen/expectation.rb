module Everdeen
  class Expectation
    attr_accessor :max_matches, :pass_through
    attr_reader :response, :request_criteria

    def initialize(args = {})
      self.max_matches = args[:max_matches]
    end

    def max_matches
      @max_matches.to_i == 0 ? 1 : @max_matches
    end

    def pass_through
      @pass_through.nil? ? false : true
    end

    def add_response(response_attr)
      @response = Response.new(response_attr)
    end

    def add_request(request_attrs)
      @request_criteria = RequestCriteria.new(request_attrs)
    end

    def to_hash
      {
        max_matches: max_matches,
        pass_through: pass_through,
        request_criteria: request_criteria.to_hash,
        respond_with: response.to_hash
      }
    end
  end
end

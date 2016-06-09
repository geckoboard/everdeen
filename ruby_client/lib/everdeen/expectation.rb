module Everdeen
  class Expectation
    attr_reader :max_matches, :response, :request_criteria

    def initialize(args = {})
      @max_matches = args[:max_matches]
      @pass_through = args[:pass_through]

      add_response(args[:response])
      add_request(args[:request_criteria])
    end

    def pass_through
      !!@pass_through
    end

    def to_hash
      {
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

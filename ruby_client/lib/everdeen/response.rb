module Everdeen
  class Response
    attr_reader :status, :headers, :body

    def initialize(args = {})
      args.each do |key, value|
        instance_variable_set("@#{key}", value)
      end
    end

    def body_encoding
      @body_encoding.to_s
    end

    def to_hash
      {
        status: status,
        headers: headers,
        body: body,
        body_encoding: body_encoding
      }
    end
  end
end

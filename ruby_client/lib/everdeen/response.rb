module Everdeen
  class Response
    attr_accessor :status, :headers, :body, :body_encoding

    def initialize(args = {})
      args.each do |key, value|
        self.send("#{key}=", value)
      end
    end

    def body_encoding
      @body_encoding.to_s
    end

    def as_hash
      {
        status: status,
        headers: headers,
        body: body,
        body_encoding: body_encoding
      }
    end
  end
end

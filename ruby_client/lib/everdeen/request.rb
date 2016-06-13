require 'base64'

module Everdeen
  class Request
    attr_reader :body_base64, :headers, :method, :url

    def initialize(args = {})
      @body_base64 = args['body_base64']
      @headers = args['headers']
      @method = args['method']
      @url = args['url']
    end

    def body
      Base64.decode64(body_base64)
    end
  end
end

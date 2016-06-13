require 'uri'
require 'net/http'
require 'json'

require 'everdeen/version'
require 'everdeen/criterion'
require 'everdeen/response'
require 'everdeen/request_criteria'
require 'everdeen/expectation'
require 'everdeen/server'
require 'everdeen/client'
require 'everdeen/request'

module Everdeen
  SUPPORTED_PLATFORMS = %w[x86_64-linux-gnu]
  BINARIES_DIR        = File.expand_path('../../binaries', __FILE__)

  class UnsupportedPlatformError < StandardError
    def initialize
      super "Everdeen does not support this architecture: #{RUBY_PLATFORM}"
    end
  end

  def self.bin_path
    raise UnsupportedPlatformError unless SUPPORTED_PLATFORMS.include? RUBY_PLATFORM
    File.join(BINARIES_DIR, "everdeen_#{Everdeen::SERVER_VERSION}_#{RUBY_PLATFORM}")
  end
end

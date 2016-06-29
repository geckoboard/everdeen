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
  BINARIES_DIR = File.expand_path('../../binaries', __FILE__)

  class UnsupportedPlatformError < StandardError
    def initialize(platform)
      super "Everdeen does not support this architecture (#{platform})"
    end
  end

  def self.bin_path
    platform = `uname -sm`

    platform_extension = case platform
                         when /^Darwin/    then 'darwin-amd64'
                         when /^Linux.*64/ then 'linux-amd64'
                         when /^Linux.*/   then 'linux-386'
                         else raise UnsupportedPlatformError, platform
                         end

    File.join(BINARIES_DIR, "everdeen_#{SERVER_VERSION}_#{platform_extension}")
  end
end

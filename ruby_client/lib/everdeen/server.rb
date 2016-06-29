require 'tempfile'

module Everdeen
  class Server
    StartTimeout = Class.new(StandardError)

    def self.start(opts = {})
      server = new(opts)
      server.start
      server
    end

    attr_reader :proxy_port, :control_port, :stderr, :ca_cert_path, :ca_key_path

    def initialize(opts = {})
      @proxy_port     = opts.fetch(:proxy_port)
      @control_port   = opts.fetch(:control_port)
      @stderr         = opts.fetch(:stderr) { Tempfile.new("everdeen") }
      @ca_cert_path   = opts.fetch(:ca_cert_path, nil)
      @ca_key_path    = opts.fetch(:ca_key_path, nil)
    end

    def start
      pid = spawn_process.pid
      wait_until_ready
      pid
    end

    def create_expectations(expectations)
      client.create_expectations(expectations).collect { |exp| Expectation.new(exp) }
    end

    def registered_expectations
      client.registered_expectations
    end

    def requests(expectation_id)
      client.requests(expectation_id)['requests'].collect { |req| Request.new(req) }
    end

    def stop
      Process.kill(:INT, @pipe.pid)
    end

    def ping
      client.ping
    end

    private

    def spawn_process
      @pipe = IO.popen([
        Everdeen.bin_path,
        "-proxy-addr=#{proxy_addr}",
        "-control-addr=#{control_addr}",
        "-ca-cert-path=#{ca_cert_path}",
        "-ca-key-path=#{ca_key_path}"
      ], err: stderr)
    end

    def wait_until_ready
      10.times do
        return if ready?
        sleep 0.1
      end

      stop
      raise StartTimeout
    end

    def ready?
      ping == 'PONG'
    rescue Errno::ECONNREFUSED
      false
    end

    def proxy_addr
      "127.0.0.1:#{proxy_port}"
    end

    def control_addr
      "127.0.0.1:#{control_port}"
    end

    def client
      @_client ||= Client.new(
        URI.parse("http://#{control_addr}")
      )
    end
  end
end

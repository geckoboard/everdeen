require 'spec_helper'

RSpec.describe Everdeen::Server, type: :unit do
  subject { Everdeen::Server.new(proxy_port: 1090, control_port: 1091, store_requests: true) }

  describe '#initialize' do
    it 'assigns proxy_port and control ports' do
      expect(subject.proxy_port).to eq 1090
      expect(subject.control_port).to eq 1091
    end
  end

  describe '#start' do
    let(:server) { Everdeen::Server.new(proxy_port: 8080, control_port: 8081) }
    let(:host) { '127.0.0.1' }

    it 'spawns a spawn process with user set args' do
      allow(server).to receive(:stop).and_return(true)
      allow(server).to receive(:wait_until_ready).and_return(true)

      expect(IO).to receive(:popen).with([
        Everdeen.bin_path,
        "-proxy-addr=#{host}:8080",
        "-control-addr=#{host}:8081"
      ])

      server.start
    end
  end
end

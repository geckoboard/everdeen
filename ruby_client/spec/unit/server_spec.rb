require 'spec_helper'

RSpec.describe Everdeen::Server, type: :unit do
  let(:stderr) { File.new('everdeen.log', "w") }
  subject { Everdeen::Server.new(proxy_port: 1090, control_port: 1091, store_requests: true, stderr: stderr) }

  describe '#initialize' do
    it 'assigns proxy_port and control ports' do
      expect(subject.proxy_port).to eq 1090
      expect(subject.control_port).to eq 1091
      expect(subject.stderr).to eq stderr
    end

    it 'defaults the stderr to tempfile' do
      subject = Everdeen::Server.new(proxy_port: 1090, control_port: 1091)
      expect(subject.stderr).to be_instance_of(Tempfile)
    end
  end

  describe '#start' do
    let(:host) { '127.0.0.1' }
    let(:io_instance) { instance_double("IO") }

    before do
      allow(subject).to receive(:stop).and_return(true)
      allow(subject).to receive(:wait_until_ready).and_return(true)
      allow(io_instance).to receive(:pid).and_return(23010)
    end

    it 'spawns a spawn process with user set args' do
      expect(IO).to receive(:popen).with([
        Everdeen.bin_path,
        "-proxy-addr=#{host}:1090",
        "-control-addr=#{host}:1091"
      ], err: stderr).and_return(io_instance)

      subject.start
    end

    it 'returns the pid' do
      allow(subject).to receive(:spawn_process).and_return(io_instance)
      expect(subject.start).to eq 23010
    end
  end
end

require 'spec_helper'

RSpec.describe Everdeen::Server, type: :unit do
  let(:stderr)  { double }

  subject do
    Everdeen::Server.new(
      proxy_port: 1090,
      control_port: 1091,
      stderr: stderr,
      ca_cert_path: "/tmp/ca_cert.crt",
      ca_key_path: "/tmp/ca_key"
    )
  end

  describe '#initialize' do
    it 'assigns proxy_port and control ports' do
      expect(subject.proxy_port).to eq 1090
      expect(subject.control_port).to eq 1091
      expect(subject.stderr).to eq stderr
      expect(subject.ca_cert_path).to eq '/tmp/ca_cert.crt'
      expect(subject.ca_key_path).to eq '/tmp/ca_key'
    end


    describe 'defaults' do
      subject { Everdeen::Server.new(proxy_port: 1090, control_port: 1091) }

      it 'defaults the stderr to tempfile' do
        expect(subject.stderr).to be_instance_of(Tempfile)
      end

      it 'defaults ca_cert and ca_key to empty string' do
        expect(subject.ca_cert_path).to be_nil
        expect(subject.ca_key_path).to be_nil
      end
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
        "-control-addr=#{host}:1091",
        "-ca-cert-path=#{subject.ca_cert_path}",
        "-ca-key-path=#{subject.ca_key_path}"
      ], err: stderr).and_return(io_instance)

      subject.start
    end

    it 'returns the pid' do
      allow(subject).to receive(:spawn_process).and_return(io_instance)
      expect(subject.start).to eq 23010
    end
  end
end

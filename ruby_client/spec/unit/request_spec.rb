require 'spec_helper'

RSpec.describe Everdeen::Request, type: :unit do
  subject do Everdeen::Request.new(
    'body_base64' => 'SGVsbG8gV29ybGQK',
    'headers'     => { 'Content-Type' => 'application/json' },
    'method'      => 'GET',
    'url'         =>'https://google.com/test')
  end

  describe '#initialize' do
    it 'assigns values to instance variables' do
      expect(subject.body_base64).to eq 'SGVsbG8gV29ybGQK'
      expect(subject.headers).to eq('Content-Type' => 'application/json')
      expect(subject.method).to eq 'GET'
      expect(subject.url).to eq 'https://google.com/test'
    end
  end

  describe '#body' do
    it 'returns unmarshalled base64 body' do
      expect(subject.body).to eq "Hello World\n"
    end
  end
end

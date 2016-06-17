require 'spec_helper'

RSpec.describe Everdeen::Request, type: :unit do
  let(:attributes) do
    {
      'url'         =>'https://google.com/test',
      'method'      => 'GET',
      'headers'     => { 'Content-Type' => 'application/json' },
      'body_base64' => 'SGVsbG8gV29ybGQK'
    }
  end

  subject { Everdeen::Request.new(attributes) }

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

  describe '#attributes' do
    it 'returns a hash representation of the object attributes' do
      expect(subject.attributes).to eq(
        :url         =>'https://google.com/test',
        :method      => 'GET',
        :headers     => { 'Content-Type' => 'application/json' },
        :body_base64 => 'SGVsbG8gV29ybGQK'
      )
    end
  end

  describe '#==' do
    let(:matching_request)  { Everdeen::Request.new(attributes) }
    let(:not_equal_request) { Everdeen::Request.new(attributes.merge('url' => 'https://google.com/tes')) }

    it 'returns false when the object isnt an Everdeen::Request' do
      expect(subject).not_to eq(Object.new)
    end

    it 'returns true when the two hashes match' do
      expect(subject).to eq(matching_request)
    end

    it 'returns false when the two hashes dont match' do
      expect(subject).not_to eq(not_equal_request)
    end
  end
end

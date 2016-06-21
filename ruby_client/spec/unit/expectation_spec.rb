require 'spec_helper'

module Everdeen
  RSpec.describe Expectation, type: :unit do
    subject { Expectation.new(store_matching_requests: true, max_matches: 3, pass_through: true, request_criteria: [], response: {}) }

    describe '#initialize' do
      let(:expected_response) {{ status: 200, body: 'Hello World', headers: nil, body_encoding: '' }}
      let(:expected_request) do
        [
          {key: nil, type: 'method', value: 'POST', match_type: 'exact'},
          {key: 'Host', type: 'header', value: 'example.com', match_type: 'exact'}
        ]
      end

      subject do
        Expectation.new(
          store_matching_requests: true,
          max_matches: 3,
          pass_through: true,
          request_criteria: [
            {type: 'method', value: 'POST'},
            {key: 'Host', type: 'header', value: 'example.com'}
          ],
          response: {status: 200, body: 'Hello World'}
        )
      end

      it 'assigns max_matches and pass_through and calls add_re(quest|sponse)' do
        expect(subject.max_matches).to eq 3
        expect(subject.store_matching_requests).to eq true
        expect(subject.pass_through).to eq true
        expect(subject.response.to_hash).to eq expected_response
        expect(subject.request_criteria.to_hash).to eq expected_request
      end
    end

    describe '#pass_through' do
      it 'defaults to false when nil' do
        subject = Expectation.new(pass_through: nil)
        expect(subject.pass_through).to be false
      end

      it 'returns true otherwise disregarding user input' do
        subject = Expectation.new(pass_through: 'not nil or true')
        expect(subject.pass_through).to be_truthy
      end
    end

    describe '#store_matching_requests' do
      it 'defaults to false when nil' do
        subject = Expectation.new(store_matching_requests: nil)
        expect(subject.store_matching_requests).to be false
      end

      it 'returns true otherwise disregarding user input' do
        subject = Expectation.new(store_matching_requests: 'not nil or true')
        expect(subject.store_matching_requests).to be_truthy
      end
    end

    describe '#add_response' do
      let(:response_hash) {{status: 200, body: 'Hello World' }}
      subject { Expectation.new(response: response_hash).response }

      it 'creates a new response object' do
        expect(subject.class).to eq Response
        expect(subject.status).to eq 200
        expect(subject.body).to eq 'Hello World'
      end
    end

    describe '#add_request' do
      let(:request_hash) {{ request_criteria:[{ type: 'method', value: 'GET' }]}}
      subject { Expectation.new(request_hash).request_criteria }

      it 'creates a new request criteria' do
        expect(subject.class).to eq RequestCriteria
        expect(subject.criteria.size).to eq 1
      end
    end

    describe '#to_hash' do
      let(:response) { Response.new(status: 200) }
      let(:request) { RequestCriteria.new(type: 'method', value: 'GET') }

      before do
        allow(subject).to receive(:request_criteria).and_return(request)
        allow(subject).to receive(:response).and_return(response)
      end

      it 'returns hash of the response defined attributes' do
        expect(subject.to_hash).to eq(
          store_matching_requests: true,
          max_matches: 3,
          pass_through: true,
          request_criteria: request.to_hash,
          respond_with: response.to_hash
        )
      end
    end
  end
end

RSpec.describe 'Expectation', type: :unit do
  subject { Everdeen::Expectation.new(max_matches: 3, pass_through: true, request_criteria: [], response: {}) }

  describe '#initialize' do
    let(:expected_response) {{ status: 200, body: 'Hello World', headers: nil, body_encoding: '' }}
    let(:expected_request) do
      [
        {key: nil, type: 'Method', value: 'POST', match_type: 'exact'},
        {key: 'Host', type: 'Header', value: 'example.com', match_type: 'exact'}
      ]
    end

    subject do
      Everdeen::Expectation.new(
        max_matches: 3,
        pass_through: true,
        request_criteria: [
          {type: 'Method', value: 'POST'},
          {key: 'Host', type: 'Header', value: 'example.com'}
        ],
        response: {status: 200, body: 'Hello World'}
      )
    end

    it 'assigns max_matches and pass_through and calls add_re(quest|sponse)' do
      expect(subject.max_matches).to eq 3
      expect(subject.pass_through).to eq true
      expect(subject.response.to_hash).to eq expected_response
      expect(subject.request_criteria.to_hash).to eq expected_request
    end


    describe 'with missing keys' do
      subject {}

      it "doesn't call add_response when no response key exists" do
        subject = Everdeen::Expectation.new(request_criteria: [])
        expect(subject).not_to receive(:add_response)
      end

      it "doesn't call add_request when no request_criteria key exists" do
        subject = Everdeen::Expectation.new(response: {})
        expect(subject).not_to receive(:add_request)
      end
    end
  end

  describe '#max_matches' do
    [{expected: 1, input: 0}, {expected: 1, input: nil}, {expected:3, input: 3}].each do |tc|
      it "returns #{tc[:expected]} when #{tc[:input].inspect}" do
        subject = Everdeen::Expectation.new(max_matches: tc[:max_matches])
        expect(subject.max_matches).to eq 1
      end
    end
  end

  describe '#pass_through' do
    it 'defaults to false when nil' do
      subject = Everdeen::Expectation.new(pass_through: nil)
      expect(subject.pass_through).to be false
    end

    it 'returns true otherwise disregarding user input' do
      subject = Everdeen::Expectation.new(pass_through: 'not nil or true')
      expect(subject.pass_through).to be_truthy
    end
  end

  describe '#add_response' do
    let(:response_hash) {{status: 200, body: 'Hello World' }}
    let(:response) { subject.add_response(response_hash) }

    it 'creates a new response object' do
      expect(response.class).to eq Everdeen::Response
      expect(response.status).to eq 200
      expect(response.body).to eq 'Hello World'
    end
  end

  describe '#add_request' do
    let(:request_hash) {{ request_criteria:[{ type: 'Method', value: 'GET' }]}}
    subject { Everdeen::Expectation.new(request_hash).request_criteria }

    it 'creates a new request criteria' do
      expect(subject.class).to eq Everdeen::RequestCriteria
      expect(subject.criteria.size).to eq 1
    end
  end

  describe '#to_hash' do
    let(:response) { Everdeen::Response.new(status: 200) }
    let(:request) { Everdeen::RequestCriteria.new(type: 'Method', value: 'GET') }

    before do
      allow(subject).to receive(:request_criteria).and_return(request)
      allow(subject).to receive(:response).and_return(response)
    end

    it 'returns hash of the response defined attributes' do
      expect(subject.to_hash).to eq(
        max_matches: 3,
        pass_through: true,
        request_criteria: request.to_hash,
        respond_with: response.to_hash
      )
    end
  end
end
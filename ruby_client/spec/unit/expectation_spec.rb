RSpec.describe 'Expectation', type: :unit do
  subject { Everdeen::Expectation.new(max_matches: 3) }

  describe '#initialize' do
    it 'assigns max matches' do
      expect(subject.max_matches).to eq 3
    end
  end

  describe '#max_matches' do
    it 'defaults it to 1 when 0' do
      subject.max_matches = 0
      expect(subject.max_matches).to eq 1
    end

    it 'defaults it to 1 when nil' do
      subject.max_matches = nil
      expect(subject.max_matches).to eq 1
    end

    it 'returns the user set value' do
      subject.max_matches = 3
      expect(subject.max_matches).to eq 3
    end
  end

  describe '#pass_through' do
    it 'defaults to false when nil' do
      subject.pass_through = nil
      expect(subject.pass_through).to be false
    end

    it 'returns true otherwise disregarding user input' do
      subject.pass_through = 'not nil but not true either'
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
    let(:request_hash) {{type: 'Method', value: 'GET' }}
    let(:request) { subject.add_request(request_hash) }

    it 'creates a new request criteria' do
      expect(request.class).to eq Everdeen::RequestCriteria
      expect(request.criteria.size).to eq 1
    end
  end

  describe '#as_hash' do
    let(:response) { Everdeen::Response.new(status: 200) }
    let(:request) { Everdeen::RequestCriteria.new(type: 'Method', value: 'GET') }

    before do
      allow(subject).to receive(:request_criteria).and_return(request)
      allow(subject).to receive(:response).and_return(response)
    end

    it 'returns hash of the response defined attributes' do
      expect(subject.as_hash).to eq(
        max_matches: 3,
        pass_through: false,
        request_criteria: request.as_hash,
        respond_with: response.as_hash
      )
    end
  end
end

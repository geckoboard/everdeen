RSpec.describe 'Response', type: :unit do
  subject do
    Everdeen::Response.new(
      status: 200,
      headers: {'Host' => 'example.com'},
      body: 'SGVsbG8gV29ybGQK',
      body_encoding: 'base64'
    )
  end

  describe '#initialize' do
    it 'assigns variables passed in the hash' do
      expect(subject.status).to eq 200
      expect(subject.headers).to eq('Host' => 'example.com')
      expect(subject.body).to eq 'SGVsbG8gV29ybGQK'
      expect(subject.body_encoding).to eq 'base64'
    end
  end

  describe '#body_encoding' do
    it 'returns empty string when body_coding is nil' do
      subject.body_encoding = nil
      expect(subject.body_encoding).to eq ''
    end

    it 'returns user set body_encoding' do
      subject.body_encoding = 'base64'
      expect(subject.body_encoding).to eq 'base64'
    end
  end

  describe '#as_hash' do
    it 'returns hash of the response defined attributes' do
      expect(subject.as_hash).to eq(
        status: 200,
        headers: {'Host' => 'example.com'},
        body: 'SGVsbG8gV29ybGQK',
        body_encoding: 'base64'
      )
    end
  end
end

RSpec.describe 'RequestCriteria', type: :unit do
  subject { Everdeen::RequestCriteria.new([{type: 'Host', value: 'example.com'}]) }

  describe '#initialize' do
    subject do
      Everdeen::RequestCriteria.new([
        {type: 'Host', value: 'example.com'},
        {type: 'Method', value: 'GET'}
      ])
    end

    it 'creates multiple criterion' do
      expect(subject.criteria.size).to eq 2
    end
  end

  describe '#criteria' do
    it 'returns criteria' do
      expect(subject.criteria.size).to eq 1
      expect(subject.criteria.first).to be_an_instance_of(Everdeen::Criterion)
    end
  end

  describe '#add_criteria' do
    it 'creates a new criterion and adds to criteria' do
      expect(Everdeen::Criterion).to receive(:new).with(type: 'Host', value: 'example.com')
      expect(subject.criteria.size).to eq 1
    end
  end

  describe '#to_hash' do
    let(:criterion_a) {{ type: 'Method', value: 'POST' }}
    let(:criterion_b) {{ key: 'Host', type:'Header', value: 'example.com', match_type: :exact }}

    before do
      allow(subject).to receive(:criteria).and_return(
        [Everdeen::Criterion.new(criterion_a), Everdeen::Criterion.new(criterion_b)])
    end

    it 'returns an array of criterion hashes' do
      expect(subject.to_hash).to eq [criterion_a.merge(key: nil, match_type: 'exact'), criterion_b]
    end
  end
end

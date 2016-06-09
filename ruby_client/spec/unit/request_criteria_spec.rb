RSpec.describe 'RequestCriteria', type: :unit do
  subject { Everdeen::RequestCriteria.new }

  describe '#criteria' do
    it 'returns an empty array when nil' do
      subject.instance_variable_set('@criteria', nil)
      expect(subject.criteria).to eq []
    end
  end

  describe '#add_criteria' do
    let(:criterion_attr) {{ type: 'Method', value: 'POST' }}

    it 'creates a new criterion and adds to criteria' do
      subject.add_criterion(criterion_attr)
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

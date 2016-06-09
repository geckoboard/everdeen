RSpec.describe 'Criterion', type: :unit do
  subject { Everdeen::Criterion.new(type: 'Header', key: 'Host', match_type: 'fuzzy', value: 'ZZ') }

  describe '#initialize' do
    it 'assigns variables passed in the hash' do
      expect(subject.type).to eq 'Header'
      expect(subject.key).to eq 'Host'
      expect(subject.match_type).to eq 'fuzzy'
      expect(subject.value).to eq 'ZZ'
    end
  end

  describe '#match_type' do
    [{returns: 'exact', match_type: nil}, {returns: 'regex', match_type: 'regex'}].each do |tc|
      it "returns #{tc[:returns]} when match_type is #{tc[:match_type].inspect}" do
        subject = Everdeen::Criterion.new(match_type: tc[:match_type])
        expect(subject.match_type).to eq tc[:returns]
      end
    end
  end

  describe '#to_hash' do
    it 'returns hash of the criterion defined attributes' do
      expect(subject.to_hash).to eq({
        key: 'Host',
        match_type: 'fuzzy',
        type: 'Header',
        value: 'ZZ'
      })
    end
  end
end

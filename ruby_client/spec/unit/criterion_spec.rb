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
    it 'returns exact when match_type is nil' do
      subject.instance_variable_set('@match_type', nil)
      expect(subject.match_type).to eq 'exact'
    end

    it 'returns user set match type' do
      subject.instance_variable_set('@match_type', 'regex')
      expect(subject.match_type).to eq 'regex'
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

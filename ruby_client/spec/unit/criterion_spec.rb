require 'spec_helper'

module Everdeen
  RSpec.describe Criterion, type: :unit do
    subject { Criterion.new(type: 'header', key: 'Host', match_type: 'fuzzy', value: 'ZZ') }

    describe '#initialize' do
      it 'assigns variables passed in the hash' do
        expect(subject.type).to eq 'header'
        expect(subject.key).to eq 'Host'
        expect(subject.match_type).to eq 'fuzzy'
        expect(subject.value).to eq 'ZZ'
      end
    end

    describe '#match_type' do
      it 'defaults to exact' do
        criterion = Criterion.new
        expect(criterion.match_type).to eq 'exact'
      end

      it 'returns user set value' do
        criterion = Criterion.new(match_type: 'regex')
        expect(criterion.match_type).to eq 'regex'
      end
    end

    describe '#to_hash' do
      it 'returns hash of the criterion defined attributes' do
        expect(subject.to_hash).to eq({
          key: 'Host',
          match_type: 'fuzzy',
          type: 'header',
          value: 'ZZ'
        })
      end
    end
  end
end

module Everdeen
  class Criterion
    attr_reader :key, :type, :value

    def initialize(args = {})
      args.each do |key, value|
        instance_variable_set("@#{key}", value)
      end
    end

    def match_type
      @match_type || 'exact'
    end

    def to_hash
      {
        key: key,
        match_type: match_type,
        type: type,
        value: value
      }
    end
  end
end

module Everdeen
  class Criterion
    attr_reader :key, :match_type, :type, :value

    def initialize(args = {})
      args.each do |key, value|
        instance_variable_set("@#{key}", value)
      end
    end

    def match_type
      @match_type.nil? ? 'exact' : @match_type
    end

    def as_hash
      {
        key: key,
        match_type: match_type,
        type: type,
        value: value
      }
    end
  end
end

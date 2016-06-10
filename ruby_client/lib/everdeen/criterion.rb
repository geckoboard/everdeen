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
      base = { key: key, match_type: match_type, type: type }

      if query_param_type? && value.is_a?(Array)
        base.merge(values: value)
      else
        base.merge(value: value)
      end
    end

    private

    def query_param_type?
      type == 'query_param'
    end
  end
end

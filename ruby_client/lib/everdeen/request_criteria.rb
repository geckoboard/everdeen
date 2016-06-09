module Everdeen
  class RequestCriteria
    def initialize(*criterion)
      criterion.each do |criterion_attr|
        add_criterion(criterion_attr)
      end
    end

    def criteria
      @criteria ||= []
    end

    def add_criterion(criterion_attrs)
      criteria << Criterion.new(criterion_attrs)
    end

    def as_hash
      criteria.map(&:as_hash)
    end
  end
end

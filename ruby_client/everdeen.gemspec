# coding: utf-8
lib = File.expand_path('../lib', __FILE__)
$LOAD_PATH.unshift(lib) unless $LOAD_PATH.include?(lib)
require 'everdeen/version'

Gem::Specification.new do |spec|
  spec.name          = "everdeen"
  spec.version       = Everdeen::VERSION
  spec.authors       = ['Daniel Upton', 'Jon Normington']
  spec.email         = ['daniel.upton@geckoboard.com', 'jon@geckoboard.com']

  spec.summary       = %q{ Everdeen ruby client to setup expectation requests through the API }
  spec.description   = spec.summary
  spec.homepage      = "https://github.com/geckoboard/everdeen"
  spec.license       = "MIT"

  spec.files         = `git ls-files -z`.split("\x0").reject { |f| f.match(%r{^(test|spec|features)/}) } + Dir['binaries/*']
  spec.bindir        = "exe"
  spec.executables   = spec.files.grep(%r{^exe/}) { |f| File.basename(f) }
  spec.require_paths = ["lib"]

  spec.add_development_dependency "bundler", "~> 1.8"
  spec.add_development_dependency "rspec", "~> 3.4"
  spec.add_development_dependency "rake", "~> 11"
end

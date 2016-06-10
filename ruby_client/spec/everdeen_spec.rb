require 'spec_helper'

RSpec.describe Everdeen, type: :integration do
  example 'starting the server and creating expectations' do
    server = Everdeen::Server.start(
      proxy_port: 4321,
      control_port: 4322
    )

    expect {
      server.create_expectations([
        Everdeen::Expectation.new(
          request_criteria: [
            {
              type: 'method',
              value: 'GET'
            }
          ],

          response: {
            status: 200,
            body: 'Hello World'
          }
        )
      ])
    }.to change { server.registered_expectations.count }.from(0).to 1

    server.stop
  end
end

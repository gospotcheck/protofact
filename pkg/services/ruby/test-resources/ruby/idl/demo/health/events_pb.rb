# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: idl/events.proto

require 'google/protobuf'

require 'github.com/gogo/protobuf/gogoproto/gogo_pb'
require 'google/protobuf/any_pb'
require 'google/protobuf/timestamp_pb'
Google::Protobuf::DescriptorPool.generated_pool.build do
  add_message "demo.health.Event" do
    optional :id, :string, 1
    optional :time, :message, 2, "google.protobuf.Timestamp"
    optional :operation, :enum, 3, "demo.health.Event.Operation"
    optional :source, :message, 4, "demo.health.Event.Source"
    optional :payload, :message, 5, "google.protobuf.Any"
  end
  add_message "demo.health.Event.Source" do
    optional :application_name, :string, 1
    optional :gsc_correlation_id, :string, 2
  end
  add_enum "demo.health.Event.Operation" do
    value :EVENT_OPERATION_INVALID, 0
    value :EVENT_OPERATION_CREATED, 1
    value :EVENT_OPERATION_UPDATED, 2
    value :EVENT_OPERATION_DELETED, 3
  end
end

module Demo
  module Health
    Event = Google::Protobuf::DescriptorPool.generated_pool.lookup("demo.health.Event").msgclass
    Event::Source = Google::Protobuf::DescriptorPool.generated_pool.lookup("demo.health.Event.Source").msgclass
    Event::Operation = Google::Protobuf::DescriptorPool.generated_pool.lookup("demo.health.Event.Operation").enummodule
  end
end

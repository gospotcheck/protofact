// Generated by the protocol buffer compiler.  DO NOT EDIT!
// source: idl/demo/health/health.proto

package com.demo.health;

public interface HealthOrBuilder extends
    // @@protoc_insertion_point(interface_extends:demo.health.Health)
    com.google.protobuf.MessageOrBuilder {

  /**
   * <code>string name = 1;</code>
   */
  java.lang.String getName();
  /**
   * <code>string name = 1;</code>
   */
  com.google.protobuf.ByteString
      getNameBytes();

  /**
   * <code>.demo.health.HealthStatus status = 2;</code>
   */
  int getStatusValue();
  /**
   * <code>.demo.health.HealthStatus status = 2;</code>
   */
  com.demo.health.HealthStatus getStatus();

  /**
   * <code>string reason = 3;</code>
   */
  java.lang.String getReason();
  /**
   * <code>string reason = 3;</code>
   */
  com.google.protobuf.ByteString
      getReasonBytes();
}
// Generated by the Scala Plugin for the Protocol Buffer Compiler.
// Do not edit!
//
// Protofile syntax: PROTO3

package com.demo.bifrost.token.v1.token

object TokenProto extends _root_.scalapb.GeneratedFileObject {
  lazy val dependencies: Seq[_root_.scalapb.GeneratedFileObject] = Seq.empty
  lazy val messagesCompanions: Seq[_root_.scalapb.GeneratedMessageCompanion[_ <: _root_.scalapb.GeneratedMessage]] =
    Seq[_root_.scalapb.GeneratedMessageCompanion[_ <: _root_.scalapb.GeneratedMessage]](
      com.demo.bifrost.token.v1.token.APIKey
    )
  private lazy val ProtoBytes: Array[Byte] =
      scalapb.Encoding.fromBase64(scala.collection.immutable.Seq(
  """CiVpZGwvY29yZS9iaWZyb3N0L3Rva2VuL3YxL3Rva2VuLnByb3RvEhVjb3JlLmJpZnJvc3QudG9rZW4udjEiWQoGQVBJS2V5E
  iAKBXRva2VuGAEgASgJQgriPwcSBXRva2VuUgV0b2tlbhItCgpjb21wYW55X2lkGAIgASgDQg7iPwsSCWNvbXBhbnlJZFIJY29tc
  GFueUlkQjgKGWNvbS5jb3JlLmJpZnJvc3QudG9rZW4udjFCClRva2VuUHJvdG9QAVoHdG9rZW52MaICA0NCVGIGcHJvdG8z"""
      ).mkString)
  lazy val scalaDescriptor: _root_.scalapb.descriptors.FileDescriptor = {
    val scalaProto = com.google.protobuf.descriptor.FileDescriptorProto.parseFrom(ProtoBytes)
    _root_.scalapb.descriptors.FileDescriptor.buildFrom(scalaProto, dependencies.map(_.scalaDescriptor))
  }
  lazy val javaDescriptor: com.google.protobuf.Descriptors.FileDescriptor = {
    val javaProto = com.google.protobuf.DescriptorProtos.FileDescriptorProto.parseFrom(ProtoBytes)
    com.google.protobuf.Descriptors.FileDescriptor.buildFrom(javaProto, Array(
    ))
  }
  @deprecated("Use javaDescriptor instead. In a future version this will refer to scalaDescriptor.", "ScalaPB 0.5.47")
  def descriptor: com.google.protobuf.Descriptors.FileDescriptor = javaDescriptor
}
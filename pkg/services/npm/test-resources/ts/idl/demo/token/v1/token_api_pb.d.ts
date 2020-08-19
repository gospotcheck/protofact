// package: demo.token.v1
// file: idl/demo/token/v1/token_api.proto

import * as jspb from "google-protobuf";
import * as idl_demo_token_v1_token_pb from "../../../../../idl/demo/token/v1/token_pb";

export class FindTokenRequest extends jspb.Message {
  getCompanyId(): number;
  setCompanyId(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): FindTokenRequest.AsObject;
  static toObject(includeInstance: boolean, msg: FindTokenRequest): FindTokenRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: FindTokenRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): FindTokenRequest;
  static deserializeBinaryFromReader(message: FindTokenRequest, reader: jspb.BinaryReader): FindTokenRequest;
}

export namespace FindTokenRequest {
  export type AsObject = {
    companyId: number,
  }
}

export class FindTokenResponse extends jspb.Message {
  hasApiKey(): boolean;
  clearApiKey(): void;
  getApiKey(): idl_demo_token_v1_token_pb.APIKey | undefined;
  setApiKey(value?: idl_demo_token_v1_token_pb.APIKey): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): FindTokenResponse.AsObject;
  static toObject(includeInstance: boolean, msg: FindTokenResponse): FindTokenResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: FindTokenResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): FindTokenResponse;
  static deserializeBinaryFromReader(message: FindTokenResponse, reader: jspb.BinaryReader): FindTokenResponse;
}

export namespace FindTokenResponse {
  export type AsObject = {
    apiKey?: idl_demo_token_v1_token_pb.APIKey.AsObject,
  }
}

export class CreateTokenRequest extends jspb.Message {
  getCompanyId(): number;
  setCompanyId(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateTokenRequest.AsObject;
  static toObject(includeInstance: boolean, msg: CreateTokenRequest): CreateTokenRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: CreateTokenRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateTokenRequest;
  static deserializeBinaryFromReader(message: CreateTokenRequest, reader: jspb.BinaryReader): CreateTokenRequest;
}

export namespace CreateTokenRequest {
  export type AsObject = {
    companyId: number,
  }
}

export class CreateTokenResponse extends jspb.Message {
  hasApiKey(): boolean;
  clearApiKey(): void;
  getApiKey(): idl_demo_token_v1_token_pb.APIKey | undefined;
  setApiKey(value?: idl_demo_token_v1_token_pb.APIKey): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateTokenResponse.AsObject;
  static toObject(includeInstance: boolean, msg: CreateTokenResponse): CreateTokenResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: CreateTokenResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateTokenResponse;
  static deserializeBinaryFromReader(message: CreateTokenResponse, reader: jspb.BinaryReader): CreateTokenResponse;
}

export namespace CreateTokenResponse {
  export type AsObject = {
    apiKey?: idl_demo_token_v1_token_pb.APIKey.AsObject,
  }
}

export class RevokeTokenRequest extends jspb.Message {
  getCompanyId(): number;
  setCompanyId(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RevokeTokenRequest.AsObject;
  static toObject(includeInstance: boolean, msg: RevokeTokenRequest): RevokeTokenRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RevokeTokenRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RevokeTokenRequest;
  static deserializeBinaryFromReader(message: RevokeTokenRequest, reader: jspb.BinaryReader): RevokeTokenRequest;
}

export namespace RevokeTokenRequest {
  export type AsObject = {
    companyId: number,
  }
}

export class RevokeTokenResponse extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RevokeTokenResponse.AsObject;
  static toObject(includeInstance: boolean, msg: RevokeTokenResponse): RevokeTokenResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RevokeTokenResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RevokeTokenResponse;
  static deserializeBinaryFromReader(message: RevokeTokenResponse, reader: jspb.BinaryReader): RevokeTokenResponse;
}

export namespace RevokeTokenResponse {
  export type AsObject = {
  }
}

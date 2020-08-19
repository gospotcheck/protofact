// package: demo.token.v1
// file: idl/demo/token/v1/token_api.proto

import * as idl_demo_token_v1_token_api_pb from "../../../../../idl/demo/token/v1/token_api_pb";
import {grpc} from "@improbable-eng/grpc-web";

type TokenAPIFindToken = {
  readonly methodName: string;
  readonly service: typeof TokenAPI;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof idl_demo_token_v1_token_api_pb.FindTokenRequest;
  readonly responseType: typeof idl_demo_token_v1_token_api_pb.FindTokenResponse;
};

type TokenAPICreateToken = {
  readonly methodName: string;
  readonly service: typeof TokenAPI;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof idl_demo_token_v1_token_api_pb.CreateTokenRequest;
  readonly responseType: typeof idl_demo_token_v1_token_api_pb.CreateTokenResponse;
};

type TokenAPIRevokeToken = {
  readonly methodName: string;
  readonly service: typeof TokenAPI;
  readonly requestStream: false;
  readonly responseStream: false;
  readonly requestType: typeof idl_demo_token_v1_token_api_pb.RevokeTokenRequest;
  readonly responseType: typeof idl_demo_token_v1_token_api_pb.RevokeTokenResponse;
};

export class TokenAPI {
  static readonly serviceName: string;
  static readonly FindToken: TokenAPIFindToken;
  static readonly CreateToken: TokenAPICreateToken;
  static readonly RevokeToken: TokenAPIRevokeToken;
}

export type ServiceError = { message: string, code: number; metadata: grpc.Metadata }
export type Status = { details: string, code: number; metadata: grpc.Metadata }

interface UnaryResponse {
  cancel(): void;
}
interface ResponseStream<T> {
  cancel(): void;
  on(type: 'data', handler: (message: T) => void): ResponseStream<T>;
  on(type: 'end', handler: (status?: Status) => void): ResponseStream<T>;
  on(type: 'status', handler: (status: Status) => void): ResponseStream<T>;
}
interface RequestStream<T> {
  write(message: T): RequestStream<T>;
  end(): void;
  cancel(): void;
  on(type: 'end', handler: (status?: Status) => void): RequestStream<T>;
  on(type: 'status', handler: (status: Status) => void): RequestStream<T>;
}
interface BidirectionalStream<ReqT, ResT> {
  write(message: ReqT): BidirectionalStream<ReqT, ResT>;
  end(): void;
  cancel(): void;
  on(type: 'data', handler: (message: ResT) => void): BidirectionalStream<ReqT, ResT>;
  on(type: 'end', handler: (status?: Status) => void): BidirectionalStream<ReqT, ResT>;
  on(type: 'status', handler: (status: Status) => void): BidirectionalStream<ReqT, ResT>;
}

export class TokenAPIClient {
  readonly serviceHost: string;

  constructor(serviceHost: string, options?: grpc.RpcOptions);
  findToken(
    requestMessage: idl_demo_token_v1_token_api_pb.FindTokenRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: idl_demo_token_v1_token_api_pb.FindTokenResponse|null) => void
  ): UnaryResponse;
  findToken(
    requestMessage: idl_demo_token_v1_token_api_pb.FindTokenRequest,
    callback: (error: ServiceError|null, responseMessage: idl_demo_token_v1_token_api_pb.FindTokenResponse|null) => void
  ): UnaryResponse;
  createToken(
    requestMessage: idl_demo_token_v1_token_api_pb.CreateTokenRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: idl_demo_token_v1_token_api_pb.CreateTokenResponse|null) => void
  ): UnaryResponse;
  createToken(
    requestMessage: idl_demo_token_v1_token_api_pb.CreateTokenRequest,
    callback: (error: ServiceError|null, responseMessage: idl_demo_token_v1_token_api_pb.CreateTokenResponse|null) => void
  ): UnaryResponse;
  revokeToken(
    requestMessage: idl_demo_token_v1_token_api_pb.RevokeTokenRequest,
    metadata: grpc.Metadata,
    callback: (error: ServiceError|null, responseMessage: idl_demo_token_v1_token_api_pb.RevokeTokenResponse|null) => void
  ): UnaryResponse;
  revokeToken(
    requestMessage: idl_demo_token_v1_token_api_pb.RevokeTokenRequest,
    callback: (error: ServiceError|null, responseMessage: idl_demo_token_v1_token_api_pb.RevokeTokenResponse|null) => void
  ): UnaryResponse;
}

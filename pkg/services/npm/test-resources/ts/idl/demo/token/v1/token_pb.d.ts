// package: demo.token.v1
// file: idl/demo/token/v1/token.proto

import * as jspb from "google-protobuf";

export class APIKey extends jspb.Message {
  getToken(): string;
  setToken(value: string): void;

  getCompanyId(): number;
  setCompanyId(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): APIKey.AsObject;
  static toObject(includeInstance: boolean, msg: APIKey): APIKey.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: APIKey, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): APIKey;
  static deserializeBinaryFromReader(message: APIKey, reader: jspb.BinaryReader): APIKey;
}

export namespace APIKey {
  export type AsObject = {
    token: string,
    companyId: number,
  }
}

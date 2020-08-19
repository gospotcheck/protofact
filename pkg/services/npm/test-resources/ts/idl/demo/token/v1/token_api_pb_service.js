// package: demo.token.v1
// file: idl/demo/token/v1/token_api.proto

var idl_demo_token_v1_token_api_pb = require("../../../../../idl/demo/token/v1/token_api_pb");
var grpc = require("@improbable-eng/grpc-web").grpc;

var TokenAPI = (function () {
  function TokenAPI() {}
  TokenAPI.serviceName = "demo.token.v1.TokenAPI";
  return TokenAPI;
}());

TokenAPI.FindToken = {
  methodName: "FindToken",
  service: TokenAPI,
  requestStream: false,
  responseStream: false,
  requestType: idl_demo_token_v1_token_api_pb.FindTokenRequest,
  responseType: idl_demo_token_v1_token_api_pb.FindTokenResponse
};

TokenAPI.CreateToken = {
  methodName: "CreateToken",
  service: TokenAPI,
  requestStream: false,
  responseStream: false,
  requestType: idl_demo_token_v1_token_api_pb.CreateTokenRequest,
  responseType: idl_demo_token_v1_token_api_pb.CreateTokenResponse
};

TokenAPI.RevokeToken = {
  methodName: "RevokeToken",
  service: TokenAPI,
  requestStream: false,
  responseStream: false,
  requestType: idl_demo_token_v1_token_api_pb.RevokeTokenRequest,
  responseType: idl_demo_token_v1_token_api_pb.RevokeTokenResponse
};

exports.TokenAPI = TokenAPI;

function TokenAPIClient(serviceHost, options) {
  this.serviceHost = serviceHost;
  this.options = options || {};
}

TokenAPIClient.prototype.findToken = function findToken(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(TokenAPI.FindToken, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

TokenAPIClient.prototype.createToken = function createToken(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(TokenAPI.CreateToken, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

TokenAPIClient.prototype.revokeToken = function revokeToken(requestMessage, metadata, callback) {
  if (arguments.length === 2) {
    callback = arguments[1];
  }
  var client = grpc.unary(TokenAPI.RevokeToken, {
    request: requestMessage,
    host: this.serviceHost,
    metadata: metadata,
    transport: this.options.transport,
    debug: this.options.debug,
    onEnd: function (response) {
      if (callback) {
        if (response.status !== grpc.Code.OK) {
          var err = new Error(response.statusMessage);
          err.code = response.status;
          err.metadata = response.trailers;
          callback(err, null);
        } else {
          callback(null, response.message);
        }
      }
    }
  });
  return {
    cancel: function () {
      callback = null;
      client.close();
    }
  };
};

exports.TokenAPIClient = TokenAPIClient;

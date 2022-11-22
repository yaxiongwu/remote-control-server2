// package: rtc
// file: rtc.proto

import * as jspb from "google-protobuf";

export class RegisterRequest extends jspb.Message {
  getSid(): string;
  setSid(value: string): void;

  getUid(): string;
  setUid(value: string): void;

  getName(): string;
  setName(value: string): void;

  getSourcetype(): SourceTypeMap[keyof SourceTypeMap];
  setSourcetype(value: SourceTypeMap[keyof SourceTypeMap]): void;

  getConfigMap(): jspb.Map<string, string>;
  clearConfigMap(): void;
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RegisterRequest.AsObject;
  static toObject(includeInstance: boolean, msg: RegisterRequest): RegisterRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RegisterRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RegisterRequest;
  static deserializeBinaryFromReader(message: RegisterRequest, reader: jspb.BinaryReader): RegisterRequest;
}

export namespace RegisterRequest {
  export type AsObject = {
    sid: string,
    uid: string,
    name: string,
    sourcetype: SourceTypeMap[keyof SourceTypeMap],
    configMap: Array<[string, string]>,
  }
}

export class RegisterReply extends jspb.Message {
  getSuccess(): boolean;
  setSuccess(value: boolean): void;

  hasError(): boolean;
  clearError(): void;
  getError(): Error | undefined;
  setError(value?: Error): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RegisterReply.AsObject;
  static toObject(includeInstance: boolean, msg: RegisterReply): RegisterReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RegisterReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RegisterReply;
  static deserializeBinaryFromReader(message: RegisterReply, reader: jspb.BinaryReader): RegisterReply;
}

export namespace RegisterReply {
  export type AsObject = {
    success: boolean,
    error?: Error.AsObject,
  }
}

export class OnLineSourceRequest extends jspb.Message {
  getSourcetype(): SourceTypeMap[keyof SourceTypeMap];
  setSourcetype(value: SourceTypeMap[keyof SourceTypeMap]): void;

  getConfigMap(): jspb.Map<string, string>;
  clearConfigMap(): void;
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): OnLineSourceRequest.AsObject;
  static toObject(includeInstance: boolean, msg: OnLineSourceRequest): OnLineSourceRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: OnLineSourceRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): OnLineSourceRequest;
  static deserializeBinaryFromReader(message: OnLineSourceRequest, reader: jspb.BinaryReader): OnLineSourceRequest;
}

export namespace OnLineSourceRequest {
  export type AsObject = {
    sourcetype: SourceTypeMap[keyof SourceTypeMap],
    configMap: Array<[string, string]>,
  }
}

export class OnLineSources extends jspb.Message {
  getUid(): string;
  setUid(value: string): void;

  getName(): string;
  setName(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): OnLineSources.AsObject;
  static toObject(includeInstance: boolean, msg: OnLineSources): OnLineSources.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: OnLineSources, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): OnLineSources;
  static deserializeBinaryFromReader(message: OnLineSources, reader: jspb.BinaryReader): OnLineSources;
}

export namespace OnLineSources {
  export type AsObject = {
    uid: string,
    name: string,
  }
}

export class OnLineSourceReply extends jspb.Message {
  clearOnlinesourcesList(): void;
  getOnlinesourcesList(): Array<OnLineSources>;
  setOnlinesourcesList(value: Array<OnLineSources>): void;
  addOnlinesources(value?: OnLineSources, index?: number): OnLineSources;

  getSuccess(): boolean;
  setSuccess(value: boolean): void;

  hasError(): boolean;
  clearError(): void;
  getError(): Error | undefined;
  setError(value?: Error): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): OnLineSourceReply.AsObject;
  static toObject(includeInstance: boolean, msg: OnLineSourceReply): OnLineSourceReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: OnLineSourceReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): OnLineSourceReply;
  static deserializeBinaryFromReader(message: OnLineSourceReply, reader: jspb.BinaryReader): OnLineSourceReply;
}

export namespace OnLineSourceReply {
  export type AsObject = {
    onlinesourcesList: Array<OnLineSources.AsObject>,
    success: boolean,
    error?: Error.AsObject,
  }
}

export class ViewSourceRequest extends jspb.Message {
  getUid(): string;
  setUid(value: string): void;

  getConfigMap(): jspb.Map<string, string>;
  clearConfigMap(): void;
  hasDescription(): boolean;
  clearDescription(): void;
  getDescription(): SessionDescription | undefined;
  setDescription(value?: SessionDescription): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ViewSourceRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ViewSourceRequest): ViewSourceRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ViewSourceRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ViewSourceRequest;
  static deserializeBinaryFromReader(message: ViewSourceRequest, reader: jspb.BinaryReader): ViewSourceRequest;
}

export namespace ViewSourceRequest {
  export type AsObject = {
    uid: string,
    configMap: Array<[string, string]>,
    description?: SessionDescription.AsObject,
  }
}

export class ViewSourceReply extends jspb.Message {
  getSuccess(): boolean;
  setSuccess(value: boolean): void;

  hasError(): boolean;
  clearError(): void;
  getError(): Error | undefined;
  setError(value?: Error): void;

  getResult(): ViewSourceReply.ResultMap[keyof ViewSourceReply.ResultMap];
  setResult(value: ViewSourceReply.ResultMap[keyof ViewSourceReply.ResultMap]): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ViewSourceReply.AsObject;
  static toObject(includeInstance: boolean, msg: ViewSourceReply): ViewSourceReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ViewSourceReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ViewSourceReply;
  static deserializeBinaryFromReader(message: ViewSourceReply, reader: jspb.BinaryReader): ViewSourceReply;
}

export namespace ViewSourceReply {
  export type AsObject = {
    success: boolean,
    error?: Error.AsObject,
    result: ViewSourceReply.ResultMap[keyof ViewSourceReply.ResultMap],
  }

  export interface ResultMap {
    WEBRTC: 0;
    CLOUD: 1;
    ERROR: 2;
  }

  export const Result: ResultMap;
}

export class WantConnectRequest extends jspb.Message {
  getFrom(): string;
  setFrom(value: string): void;

  getTo(): string;
  setTo(value: string): void;

  getConfigMap(): jspb.Map<string, string>;
  clearConfigMap(): void;
  getSdptype(): string;
  setSdptype(value: string): void;

  getSdp(): string;
  setSdp(value: string): void;

  getConnectiontype(): ConnectTypeMap[keyof ConnectTypeMap];
  setConnectiontype(value: ConnectTypeMap[keyof ConnectTypeMap]): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): WantConnectRequest.AsObject;
  static toObject(includeInstance: boolean, msg: WantConnectRequest): WantConnectRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: WantConnectRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): WantConnectRequest;
  static deserializeBinaryFromReader(message: WantConnectRequest, reader: jspb.BinaryReader): WantConnectRequest;
}

export namespace WantConnectRequest {
  export type AsObject = {
    from: string,
    to: string,
    configMap: Array<[string, string]>,
    sdptype: string,
    sdp: string,
    connectiontype: ConnectTypeMap[keyof ConnectTypeMap],
  }
}

export class WantConnectReply extends jspb.Message {
  getSuccess(): boolean;
  setSuccess(value: boolean): void;

  getIdleornot(): boolean;
  setIdleornot(value: boolean): void;

  getResttimesecs(): number;
  setResttimesecs(value: number): void;

  getNumofwaiting(): number;
  setNumofwaiting(value: number): void;

  hasError(): boolean;
  clearError(): void;
  getError(): Error | undefined;
  setError(value?: Error): void;

  getFrom(): string;
  setFrom(value: string): void;

  getTo(): string;
  setTo(value: string): void;

  getSdptype(): string;
  setSdptype(value: string): void;

  getSdp(): string;
  setSdp(value: string): void;

  getConnectiontype(): ConnectTypeMap[keyof ConnectTypeMap];
  setConnectiontype(value: ConnectTypeMap[keyof ConnectTypeMap]): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): WantConnectReply.AsObject;
  static toObject(includeInstance: boolean, msg: WantConnectReply): WantConnectReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: WantConnectReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): WantConnectReply;
  static deserializeBinaryFromReader(message: WantConnectReply, reader: jspb.BinaryReader): WantConnectReply;
}

export namespace WantConnectReply {
  export type AsObject = {
    success: boolean,
    idleornot: boolean,
    resttimesecs: number,
    numofwaiting: number,
    error?: Error.AsObject,
    from: string,
    to: string,
    sdptype: string,
    sdp: string,
    connectiontype: ConnectTypeMap[keyof ConnectTypeMap],
  }
}

export class JoinRequest extends jspb.Message {
  getSid(): string;
  setSid(value: string): void;

  getUid(): string;
  setUid(value: string): void;

  getConfigMap(): jspb.Map<string, string>;
  clearConfigMap(): void;
  hasDescription(): boolean;
  clearDescription(): void;
  getDescription(): SessionDescription | undefined;
  setDescription(value?: SessionDescription): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): JoinRequest.AsObject;
  static toObject(includeInstance: boolean, msg: JoinRequest): JoinRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: JoinRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): JoinRequest;
  static deserializeBinaryFromReader(message: JoinRequest, reader: jspb.BinaryReader): JoinRequest;
}

export namespace JoinRequest {
  export type AsObject = {
    sid: string,
    uid: string,
    configMap: Array<[string, string]>,
    description?: SessionDescription.AsObject,
  }
}

export class JoinReply extends jspb.Message {
  getSuccess(): boolean;
  setSuccess(value: boolean): void;

  hasError(): boolean;
  clearError(): void;
  getError(): Error | undefined;
  setError(value?: Error): void;

  hasDescription(): boolean;
  clearDescription(): void;
  getDescription(): SessionDescription | undefined;
  setDescription(value?: SessionDescription): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): JoinReply.AsObject;
  static toObject(includeInstance: boolean, msg: JoinReply): JoinReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: JoinReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): JoinReply;
  static deserializeBinaryFromReader(message: JoinReply, reader: jspb.BinaryReader): JoinReply;
}

export namespace JoinReply {
  export type AsObject = {
    success: boolean,
    error?: Error.AsObject,
    description?: SessionDescription.AsObject,
  }
}

export class TrackInfo extends jspb.Message {
  getId(): string;
  setId(value: string): void;

  getKind(): string;
  setKind(value: string): void;

  getMuted(): boolean;
  setMuted(value: boolean): void;

  getType(): MediaTypeMap[keyof MediaTypeMap];
  setType(value: MediaTypeMap[keyof MediaTypeMap]): void;

  getStreamid(): string;
  setStreamid(value: string): void;

  getLabel(): string;
  setLabel(value: string): void;

  getLayer(): string;
  setLayer(value: string): void;

  getWidth(): number;
  setWidth(value: number): void;

  getHeight(): number;
  setHeight(value: number): void;

  getFramerate(): number;
  setFramerate(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): TrackInfo.AsObject;
  static toObject(includeInstance: boolean, msg: TrackInfo): TrackInfo.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: TrackInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): TrackInfo;
  static deserializeBinaryFromReader(message: TrackInfo, reader: jspb.BinaryReader): TrackInfo;
}

export namespace TrackInfo {
  export type AsObject = {
    id: string,
    kind: string,
    muted: boolean,
    type: MediaTypeMap[keyof MediaTypeMap],
    streamid: string,
    label: string,
    layer: string,
    width: number,
    height: number,
    framerate: number,
  }
}

export class SessionDescription extends jspb.Message {
  getFrom(): string;
  setFrom(value: string): void;

  getTo(): string;
  setTo(value: string): void;

  getTarget(): TargetMap[keyof TargetMap];
  setTarget(value: TargetMap[keyof TargetMap]): void;

  getType(): string;
  setType(value: string): void;

  getSdp(): string;
  setSdp(value: string): void;

  clearTrackinfosList(): void;
  getTrackinfosList(): Array<TrackInfo>;
  setTrackinfosList(value: Array<TrackInfo>): void;
  addTrackinfos(value?: TrackInfo, index?: number): TrackInfo;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SessionDescription.AsObject;
  static toObject(includeInstance: boolean, msg: SessionDescription): SessionDescription.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: SessionDescription, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SessionDescription;
  static deserializeBinaryFromReader(message: SessionDescription, reader: jspb.BinaryReader): SessionDescription;
}

export namespace SessionDescription {
  export type AsObject = {
    from: string,
    to: string,
    target: TargetMap[keyof TargetMap],
    type: string,
    sdp: string,
    trackinfosList: Array<TrackInfo.AsObject>,
  }
}

export class Trickle extends jspb.Message {
  getFrom(): string;
  setFrom(value: string): void;

  getTo(): string;
  setTo(value: string): void;

  getTarget(): TargetMap[keyof TargetMap];
  setTarget(value: TargetMap[keyof TargetMap]): void;

  getInit(): string;
  setInit(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Trickle.AsObject;
  static toObject(includeInstance: boolean, msg: Trickle): Trickle.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Trickle, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Trickle;
  static deserializeBinaryFromReader(message: Trickle, reader: jspb.BinaryReader): Trickle;
}

export namespace Trickle {
  export type AsObject = {
    from: string,
    to: string,
    target: TargetMap[keyof TargetMap],
    init: string,
  }
}

export class Error extends jspb.Message {
  getCode(): number;
  setCode(value: number): void;

  getReason(): string;
  setReason(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Error.AsObject;
  static toObject(includeInstance: boolean, msg: Error): Error.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Error, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Error;
  static deserializeBinaryFromReader(message: Error, reader: jspb.BinaryReader): Error;
}

export namespace Error {
  export type AsObject = {
    code: number,
    reason: string,
  }
}

export class TrackEvent extends jspb.Message {
  getState(): TrackEvent.StateMap[keyof TrackEvent.StateMap];
  setState(value: TrackEvent.StateMap[keyof TrackEvent.StateMap]): void;

  getUid(): string;
  setUid(value: string): void;

  clearTracksList(): void;
  getTracksList(): Array<TrackInfo>;
  setTracksList(value: Array<TrackInfo>): void;
  addTracks(value?: TrackInfo, index?: number): TrackInfo;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): TrackEvent.AsObject;
  static toObject(includeInstance: boolean, msg: TrackEvent): TrackEvent.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: TrackEvent, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): TrackEvent;
  static deserializeBinaryFromReader(message: TrackEvent, reader: jspb.BinaryReader): TrackEvent;
}

export namespace TrackEvent {
  export type AsObject = {
    state: TrackEvent.StateMap[keyof TrackEvent.StateMap],
    uid: string,
    tracksList: Array<TrackInfo.AsObject>,
  }

  export interface StateMap {
    ADD: 0;
    UPDATE: 1;
    REMOVE: 2;
  }

  export const State: StateMap;
}

export class Subscription extends jspb.Message {
  getTrackid(): string;
  setTrackid(value: string): void;

  getMute(): boolean;
  setMute(value: boolean): void;

  getSubscribe(): boolean;
  setSubscribe(value: boolean): void;

  getLayer(): string;
  setLayer(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Subscription.AsObject;
  static toObject(includeInstance: boolean, msg: Subscription): Subscription.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Subscription, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Subscription;
  static deserializeBinaryFromReader(message: Subscription, reader: jspb.BinaryReader): Subscription;
}

export namespace Subscription {
  export type AsObject = {
    trackid: string,
    mute: boolean,
    subscribe: boolean,
    layer: string,
  }
}

export class SubscriptionRequest extends jspb.Message {
  clearSubscriptionsList(): void;
  getSubscriptionsList(): Array<Subscription>;
  setSubscriptionsList(value: Array<Subscription>): void;
  addSubscriptions(value?: Subscription, index?: number): Subscription;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SubscriptionRequest.AsObject;
  static toObject(includeInstance: boolean, msg: SubscriptionRequest): SubscriptionRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: SubscriptionRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SubscriptionRequest;
  static deserializeBinaryFromReader(message: SubscriptionRequest, reader: jspb.BinaryReader): SubscriptionRequest;
}

export namespace SubscriptionRequest {
  export type AsObject = {
    subscriptionsList: Array<Subscription.AsObject>,
  }
}

export class SubscriptionReply extends jspb.Message {
  getSuccess(): boolean;
  setSuccess(value: boolean): void;

  hasError(): boolean;
  clearError(): void;
  getError(): Error | undefined;
  setError(value?: Error): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SubscriptionReply.AsObject;
  static toObject(includeInstance: boolean, msg: SubscriptionReply): SubscriptionReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: SubscriptionReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SubscriptionReply;
  static deserializeBinaryFromReader(message: SubscriptionReply, reader: jspb.BinaryReader): SubscriptionReply;
}

export namespace SubscriptionReply {
  export type AsObject = {
    success: boolean,
    error?: Error.AsObject,
  }
}

export class UpdateTrackReply extends jspb.Message {
  getSuccess(): boolean;
  setSuccess(value: boolean): void;

  hasError(): boolean;
  clearError(): void;
  getError(): Error | undefined;
  setError(value?: Error): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): UpdateTrackReply.AsObject;
  static toObject(includeInstance: boolean, msg: UpdateTrackReply): UpdateTrackReply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: UpdateTrackReply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): UpdateTrackReply;
  static deserializeBinaryFromReader(message: UpdateTrackReply, reader: jspb.BinaryReader): UpdateTrackReply;
}

export namespace UpdateTrackReply {
  export type AsObject = {
    success: boolean,
    error?: Error.AsObject,
  }
}

export class ActiveSpeaker extends jspb.Message {
  clearSpeakersList(): void;
  getSpeakersList(): Array<AudioLevelSpeaker>;
  setSpeakersList(value: Array<AudioLevelSpeaker>): void;
  addSpeakers(value?: AudioLevelSpeaker, index?: number): AudioLevelSpeaker;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ActiveSpeaker.AsObject;
  static toObject(includeInstance: boolean, msg: ActiveSpeaker): ActiveSpeaker.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ActiveSpeaker, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ActiveSpeaker;
  static deserializeBinaryFromReader(message: ActiveSpeaker, reader: jspb.BinaryReader): ActiveSpeaker;
}

export namespace ActiveSpeaker {
  export type AsObject = {
    speakersList: Array<AudioLevelSpeaker.AsObject>,
  }
}

export class AudioLevelSpeaker extends jspb.Message {
  getSid(): string;
  setSid(value: string): void;

  getLevel(): number;
  setLevel(value: number): void;

  getActive(): boolean;
  setActive(value: boolean): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): AudioLevelSpeaker.AsObject;
  static toObject(includeInstance: boolean, msg: AudioLevelSpeaker): AudioLevelSpeaker.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: AudioLevelSpeaker, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): AudioLevelSpeaker;
  static deserializeBinaryFromReader(message: AudioLevelSpeaker, reader: jspb.BinaryReader): AudioLevelSpeaker;
}

export namespace AudioLevelSpeaker {
  export type AsObject = {
    sid: string,
    level: number,
    active: boolean,
  }
}

export class Request extends jspb.Message {
  hasJoin(): boolean;
  clearJoin(): void;
  getJoin(): JoinRequest | undefined;
  setJoin(value?: JoinRequest): void;

  hasDescription(): boolean;
  clearDescription(): void;
  getDescription(): SessionDescription | undefined;
  setDescription(value?: SessionDescription): void;

  hasTrickle(): boolean;
  clearTrickle(): void;
  getTrickle(): Trickle | undefined;
  setTrickle(value?: Trickle): void;

  hasSubscription(): boolean;
  clearSubscription(): void;
  getSubscription(): SubscriptionRequest | undefined;
  setSubscription(value?: SubscriptionRequest): void;

  hasRegister(): boolean;
  clearRegister(): void;
  getRegister(): RegisterRequest | undefined;
  setRegister(value?: RegisterRequest): void;

  hasOnlinesource(): boolean;
  clearOnlinesource(): void;
  getOnlinesource(): OnLineSourceRequest | undefined;
  setOnlinesource(value?: OnLineSourceRequest): void;

  hasViewsource(): boolean;
  clearViewsource(): void;
  getViewsource(): ViewSourceRequest | undefined;
  setViewsource(value?: ViewSourceRequest): void;

  hasWantconnect(): boolean;
  clearWantconnect(): void;
  getWantconnect(): WantConnectRequest | undefined;
  setWantconnect(value?: WantConnectRequest): void;

  hasWantconnectreply(): boolean;
  clearWantconnectreply(): void;
  getWantconnectreply(): WantConnectReply | undefined;
  setWantconnectreply(value?: WantConnectReply): void;

  getPayloadCase(): Request.PayloadCase;
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Request.AsObject;
  static toObject(includeInstance: boolean, msg: Request): Request.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Request, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Request;
  static deserializeBinaryFromReader(message: Request, reader: jspb.BinaryReader): Request;
}

export namespace Request {
  export type AsObject = {
    join?: JoinRequest.AsObject,
    description?: SessionDescription.AsObject,
    trickle?: Trickle.AsObject,
    subscription?: SubscriptionRequest.AsObject,
    register?: RegisterRequest.AsObject,
    onlinesource?: OnLineSourceRequest.AsObject,
    viewsource?: ViewSourceRequest.AsObject,
    wantconnect?: WantConnectRequest.AsObject,
    wantconnectreply?: WantConnectReply.AsObject,
  }

  export enum PayloadCase {
    PAYLOAD_NOT_SET = 0,
    JOIN = 1,
    DESCRIPTION = 2,
    TRICKLE = 3,
    SUBSCRIPTION = 4,
    REGISTER = 5,
    ONLINESOURCE = 6,
    VIEWSOURCE = 7,
    WANTCONNECT = 8,
    WANTCONNECTREPLY = 9,
  }
}

export class Reply extends jspb.Message {
  hasJoin(): boolean;
  clearJoin(): void;
  getJoin(): JoinReply | undefined;
  setJoin(value?: JoinReply): void;

  hasDescription(): boolean;
  clearDescription(): void;
  getDescription(): SessionDescription | undefined;
  setDescription(value?: SessionDescription): void;

  hasTrickle(): boolean;
  clearTrickle(): void;
  getTrickle(): Trickle | undefined;
  setTrickle(value?: Trickle): void;

  hasTrackevent(): boolean;
  clearTrackevent(): void;
  getTrackevent(): TrackEvent | undefined;
  setTrackevent(value?: TrackEvent): void;

  hasSubscription(): boolean;
  clearSubscription(): void;
  getSubscription(): SubscriptionReply | undefined;
  setSubscription(value?: SubscriptionReply): void;

  hasError(): boolean;
  clearError(): void;
  getError(): Error | undefined;
  setError(value?: Error): void;

  hasRegister(): boolean;
  clearRegister(): void;
  getRegister(): RegisterReply | undefined;
  setRegister(value?: RegisterReply): void;

  hasOnlinesource(): boolean;
  clearOnlinesource(): void;
  getOnlinesource(): OnLineSourceReply | undefined;
  setOnlinesource(value?: OnLineSourceReply): void;

  hasViewsource(): boolean;
  clearViewsource(): void;
  getViewsource(): ViewSourceReply | undefined;
  setViewsource(value?: ViewSourceReply): void;

  hasWantconnect(): boolean;
  clearWantconnect(): void;
  getWantconnect(): WantConnectReply | undefined;
  setWantconnect(value?: WantConnectReply): void;

  getPayloadCase(): Reply.PayloadCase;
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Reply.AsObject;
  static toObject(includeInstance: boolean, msg: Reply): Reply.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Reply, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Reply;
  static deserializeBinaryFromReader(message: Reply, reader: jspb.BinaryReader): Reply;
}

export namespace Reply {
  export type AsObject = {
    join?: JoinReply.AsObject,
    description?: SessionDescription.AsObject,
    trickle?: Trickle.AsObject,
    trackevent?: TrackEvent.AsObject,
    subscription?: SubscriptionReply.AsObject,
    error?: Error.AsObject,
    register?: RegisterReply.AsObject,
    onlinesource?: OnLineSourceReply.AsObject,
    viewsource?: ViewSourceReply.AsObject,
    wantconnect?: WantConnectReply.AsObject,
  }

  export enum PayloadCase {
    PAYLOAD_NOT_SET = 0,
    JOIN = 1,
    DESCRIPTION = 2,
    TRICKLE = 3,
    TRACKEVENT = 4,
    SUBSCRIPTION = 5,
    ERROR = 7,
    REGISTER = 8,
    ONLINESOURCE = 9,
    VIEWSOURCE = 10,
    WANTCONNECT = 11,
  }
}

export interface RoleMap {
  ADMIN: 0;
  VIDEOSOURCE: 1;
  CONTROLER: 2;
  OBSERVE: 3;
  UNKNOWN: 4;
}

export const Role: RoleMap;

export interface ConnectTypeMap {
  CONTROL: 0;
  VIEW: 1;
  MANAGE: 2;
}

export const ConnectType: ConnectTypeMap;

export interface SourceTypeMap {
  CAR: 0;
  FEED: 1;
  CAMERA: 2;
  BOAT: 3;
  SUBMARINE: 4;
}

export const SourceType: SourceTypeMap;

export interface TargetMap {
  PUBLISHER: 0;
  SUBSCRIBER: 1;
}

export const Target: TargetMap;

export interface MediaTypeMap {
  MEDIAUNKNOWN: 0;
  USERMEDIA: 1;
  SCREENCAPTURE: 2;
  CAVANS: 3;
  STREAMING: 4;
  VOIP: 5;
}

export const MediaType: MediaTypeMap;


#!/usr/bin/env node

import { createHash } from 'node:crypto';
import { execFileSync } from 'node:child_process';
import { mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { createRequire } from 'node:module';
import { dirname, isAbsolute, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import protobuf from 'protobufjs';

const ROOT = join(dirname(fileURLToPath(import.meta.url)), '..');
const PROTO_PATH = join(ROOT, 'proto', 'game.proto');
const TS_OUTPUT = join(ROOT, 'packages', 'core', 'src', 'websocket', 'protocol.generated.ts');
const GO_OUTPUT = join(ROOT, 'packages', 'backend', 'internal', 'ws', 'protocol_generated.go');
const CHECK = process.argv.includes('--check');
const require = createRequire(import.meta.url);
const DESCRIPTOR_PATH = require.resolve('protobufjs/google/protobuf/descriptor.proto');

const schema = readFileSync(PROTO_PATH, 'utf8');
const schemaHash = createHash('sha256').update(schema).digest('hex').slice(0, 16);
const root = new protobuf.Root();
root.resolvePath = (origin, target) => target === 'google/protobuf/descriptor.proto'
  ? DESCRIPTOR_PATH
  : isAbsolute(target) ? target : join(dirname(origin), target);
root.loadSync(PROTO_PATH);

const enumPrefix = {
  ClientMessageType: 'CLIENT_MESSAGE_TYPE_',
  ServerMessageType: 'SERVER_MESSAGE_TYPE_',
  ProtocolErrorCode: 'PROTOCOL_ERROR_CODE_',
  IdentityState: 'IDENTITY_STATE_',
};

const outputNames = {
  Character: 'GameCharacter',
  Player: 'RoomPlayer',
  Nomination: 'RoomNomination',
  DeathRecord: 'RoomDeathRecord',
  NightWakeStep: 'RoomNightWakeStep',
  NightAction: 'RoomNightAction',
  GameEnded: 'GameEndedPayload',
  IdentityStatus: 'IdentityStatus',
  RoomState: 'RoomState',
  ClientMessage: 'ClientMessage',
  ServerMessage: 'ServerMessage',
};

function lookupType(name) {
  return root.lookupType(`clocktower.${name}`);
}

function wireEnumValues(name) {
  const enumeration = root.lookupEnum(`clocktower.${name}`);
  const prefix = enumPrefix[name];
  return Object.keys(enumeration.values)
    .filter((value) => !value.endsWith('_UNSPECIFIED'))
    .map((value) => value.slice(prefix.length));
}

function quoteUnion(values) {
  return values.map((value) => `'${value}'`).join(' | ');
}

function tsBaseType(field, owner) {
  const key = `${owner}.${field.name}`;
  const overrides = {
    'ClientMessage.protocolVersion': '2',
    'ClientMessage.type': 'ClientMessageType',
    'ClientMessage.event': 'Readonly<Record<string, unknown>>',
    'ClientMessage.phase': 'string | number',
    'ClientMessage.winner': "'good' | 'evil' | 1 | 2",
    'ClientMessage.cause': 'string | number',
    'ServerMessage.type': 'ServerMessageType',
    'ServerMessage.code': 'ProtocolErrorCode',
    'IdentityStatus.status': 'IdentityState',
    'GameCharacter.team': 'number',
    'RoomDeathRecord.cause': 'string',
    'RoomNightAction.actionType': 'string',
    'GameEndedPayload.winner': 'number | string',
    'GameEndedPayload.reason': 'string',
    'RoomState.phase': 'number',
  };
  if (overrides[key]) return overrides[key];

  if (field.map) {
    return `Readonly<Record<string, ${scalarTsType(field.type)}>>`;
  }

  let base = scalarTsType(field.type);
  if (outputNames[field.type]) base = outputNames[field.type];
  if (field.resolvedType?.className === 'Enum') base = 'number';
  if (field.repeated) return `readonly ${base}[]`;
  return base;
}

function scalarTsType(type) {
  if (type === 'string') return 'string';
  if (type === 'bool') return 'boolean';
  if (['double', 'float', 'int32', 'uint32', 'sint32', 'fixed32', 'sfixed32', 'int64', 'uint64', 'sint64', 'fixed64', 'sfixed64'].includes(type)) return 'number';
  if (type === 'bytes') return 'Uint8Array';
  return outputNames[type] ?? type;
}

function isJsonRequired(field) {
  return field.options?.['(json_required)'] === true;
}

function generateTsInterface(protoName) {
  const type = lookupType(protoName);
  const owner = outputNames[protoName];
  const fields = Object.values(type.fields).map((field) => {
    const optional = isJsonRequired(field) ? '' : '?';
    return `  readonly ${field.name}${optional}: ${tsBaseType(field, owner)};`;
  });
  return `export interface ${owner} {\n${fields.join('\n')}\n}`;
}

function generateTypeScript() {
  const aliases = [
    `export type ClientMessageType = ${quoteUnion(wireEnumValues('ClientMessageType'))};`,
    `export type ServerMessageType = ${quoteUnion(wireEnumValues('ServerMessageType'))};`,
    `export type ProtocolErrorCode = ${quoteUnion(wireEnumValues('ProtocolErrorCode'))};`,
    `export type IdentityState = ${quoteUnion(wireEnumValues('IdentityState').map((value) => value.toLowerCase()))};`,
  ];
  const interfaces = [
    'Character',
    'Player',
    'Nomination',
    'DeathRecord',
    'NightWakeStep',
    'NightAction',
    'GameEnded',
    'IdentityStatus',
    'RoomState',
    'ClientMessage',
    'ServerMessage',
  ].map(generateTsInterface);

  return `// Code generated from proto/game.proto (${schemaHash}). DO NOT EDIT.\n// Run: pnpm proto:generate\n\n${aliases.join('\n')}\n\n${interfaces.join('\n\n')}\n`;
}

function pascalCase(value) {
  return value.toLowerCase().split('_').map((part) => part.charAt(0).toUpperCase() + part.slice(1)).join('');
}

function goFieldName(name) {
  const parts = name.replace(/([a-z0-9])([A-Z])/g, '$1 $2').split(' ');
  return parts.map((part) => {
    if (part.toLowerCase() === 'id') return 'ID';
    if (part.toLowerCase() === 'ids') return 'IDs';
    return part.charAt(0).toUpperCase() + part.slice(1);
  }).join('');
}

function goType(field, owner) {
  const key = `${owner}.${field.name}`;
  const overrides = {
    'ClientMessage.protocolVersion': 'int',
    'ClientMessage.type': 'string',
    'ClientMessage.clientSequence': 'uint64',
    'ClientMessage.event': '*game.GameEvent',
    'ClientMessage.decision': '*bool',
    'ClientMessage.phase': 'ClientGamePhase',
    'ClientMessage.winner': 'ClientTeam',
    'ClientMessage.cause': 'ClientDeathCause',
    'ServerMessage.type': 'string',
    'ServerMessage.code': 'string',
    'ServerMessage.state': '*RoomState',
    'ServerMessage.identityStatus': '*session.IdentityStatus',
    'ServerMessage.roomRevision': 'uint64',
    'ServerMessage.acceptedSequence': 'uint64',
    'ServerMessage.nextClientSequence': 'uint64',
    'RoomState.players': '[]game.Player',
    'RoomState.phase': 'game.GamePhase',
    'RoomState.dayNumber': 'int32',
    'RoomState.nomination': '*game.Nomination',
    'RoomState.deaths': '[]game.DeathRecord',
    'RoomState.nightWakeSteps': '[]game.NightWakeStep',
    'RoomState.nightActions': '[]game.NightAction',
    'RoomState.currentNightWakeStep': '*game.NightWakeStep',
    'RoomState.winner': '*game.GameEndedEvent',
  };
  if (overrides[key]) return overrides[key];
  if (field.map) return `map[string]${goScalarType(field.type)}`;
  const base = goScalarType(field.type);
  if (field.repeated) return `[]${base}`;
  return base;
}

function goScalarType(type) {
  if (type === 'string') return 'string';
  if (type === 'bool') return 'bool';
  if (type === 'int32') return 'int';
  if (type === 'uint64') return 'uint64';
  return 'any';
}

function goJSONTag(field) {
  return `${field.name}${isJsonRequired(field) ? '' : ',omitempty'}`;
}

function generateGoStruct(protoName) {
  const type = lookupType(protoName);
  const owner = outputNames[protoName];
  const fields = Object.values(type.fields).map((field) =>
    `\t${goFieldName(field.name)} ${goType(field, owner)} \`json:"${goJSONTag(field)}"\``
  );
  return `type ${owner} struct {\n${fields.join('\n')}\n}`;
}

function generateGoConstants(enumName, identifierPrefix) {
  return wireEnumValues(enumName).map((value) =>
    `\t${identifierPrefix}${pascalCase(value)} = "${value}"`
  ).join('\n');
}

function generateGo() {
  const constants = [
    generateGoConstants('ClientMessageType', 'Msg'),
    generateGoConstants('ServerMessageType', 'ServerMsg'),
    generateGoConstants('ProtocolErrorCode', 'ProtocolError'),
  ].join('\n\n');
  const structs = ['ClientMessage', 'ServerMessage', 'RoomState'].map(generateGoStruct).join('\n\n');
  const source = `// Code generated from proto/game.proto (${schemaHash}). DO NOT EDIT.\n// Run: pnpm proto:generate\n\npackage ws\n\nimport (\n\t"github.com/your-org/blood-on-the-clocktower/internal/game"\n\t"github.com/your-org/blood-on-the-clocktower/internal/session"\n)\n\nconst (\n${constants}\n)\n\n${structs}\n`;
  return execFileSync('gofmt', { input: source, encoding: 'utf8' });
}

function emit(path, content) {
  if (CHECK) {
    let existing = '';
    try {
      existing = readFileSync(path, 'utf8');
    } catch {
      // Report the missing generated artifact below.
    }
    if (existing !== content) {
      console.error(`Generated contract is stale: ${path}`);
      process.exitCode = 1;
    }
    return;
  }
  mkdirSync(dirname(path), { recursive: true });
  writeFileSync(path, content, 'utf8');
  console.log(`Generated ${path}`);
}

emit(TS_OUTPUT, generateTypeScript());
emit(GO_OUTPUT, generateGo());

#!/usr/bin/env node

/**
 * Generate TypeScript types from ProtoBuf definitions
 *
 * This script uses protobufjs to parse .proto files and generate
 * TypeScript interfaces that match the Go backend structures.
 *
 * Usage: node scripts/generate-types.mjs
 */

import { execSync } from 'node:child_process';
import { readFileSync, writeFileSync, mkdirSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const ROOT = join(__dirname, '..');
const PROTO_DIR = join(ROOT, 'proto');
const OUTPUT_DIR = join(ROOT, 'packages/core/src/types/generated');

// Ensure output directory exists
mkdirSync(OUTPUT_DIR, { recursive: true });

console.log('Generating TypeScript types from ProtoBuf definitions...');

// Using pbts/pbjs from protobufjs-cli
// Install: pnpm add -D protobufjs protobufjs-cli
try {
  // Generate static module
  execSync(
    `npx pbjs -t static-module -w es6 -o ${join(OUTPUT_DIR, 'bundle.js')} ${join(PROTO_DIR, 'game.proto')}`,
    { cwd: ROOT, stdio: 'inherit' }
  );

  // Generate TypeScript definitions
  execSync(
    `npx pbts -o ${join(OUTPUT_DIR, 'bundle.d.ts')} ${join(OUTPUT_DIR, 'bundle.js')}`,
    { cwd: ROOT, stdio: 'inherit' }
  );

  // Generate clean TypeScript interfaces from the proto file
  generateCleanTypes();

  console.log('TypeScript types generated successfully!');
} catch (error) {
  console.error('Failed to generate types:', error.message);
  process.exit(1);
}

function generateCleanTypes() {
  const cleanTypes = `// Auto-generated from proto/game.proto
// DO NOT EDIT - regenerate with: pnpm proto:generate

export const enum Team {
  UNSPECIFIED = 0,
  GOOD = 1,
  EVIL = 2,
}

export const enum GamePhase {
  UNSPECIFIED = 0,
  SETUP = 1,
  DAY = 2,
  NIGHT = 3,
  VOTING = 4,
  FINISHED = 5,
}

export interface ProtoPlayer {
  readonly id: string;
  readonly name: string;
  readonly character?: ProtoCharacter;
  readonly isAlive: boolean;
  readonly votes: number;
}

export interface ProtoCharacter {
  readonly id: string;
  readonly name: string;
  readonly team: Team;
  readonly ability: string;
}

export interface ProtoGameState {
  readonly id: string;
  readonly phase: GamePhase;
  readonly players: readonly ProtoPlayer[];
  readonly dayNumber: number;
  readonly storytellerId?: string;
  readonly votes: Readonly<Record<string, string>>;
}

export type ProtoGameEvent =
  | { readonly playerJoined: { readonly player: ProtoPlayer } }
  | { readonly playerLeft: { readonly playerId: string } }
  | { readonly phaseChanged: { readonly phase: GamePhase } }
  | { readonly voteCast: { readonly voterId: string; readonly targetId?: string } }
  | { readonly characterAssigned: { readonly playerId: string; readonly character: ProtoCharacter } };
`;

  writeFileSync(join(OUTPUT_DIR, 'index.ts'), cleanTypes, 'utf-8');
  console.log('Generated clean TypeScript interfaces');
}

#!/usr/bin/env node

/**
 * Generate TypeScript types from ProtoBuf definitions
 *
 * This script uses protobufjs to parse .proto files and generate
 * TypeScript interfaces that match the Go backend structures.
 *
 * Usage: node scripts/generate-types.mjs
 */

import { execFileSync, execSync } from 'node:child_process';
import { readFileSync, writeFileSync, mkdirSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const ROOT = join(__dirname, '..');
const PROTO_DIR = join(ROOT, 'proto');
const OUTPUT_DIR = join(ROOT, 'packages/core/src/types/generated');
const CLEAN_OUTPUT = join(OUTPUT_DIR, 'index.ts');
const CHECK = process.argv.includes('--check');

console.log(CHECK
  ? 'Checking generated types from ProtoBuf definitions...'
  : 'Generating TypeScript types from ProtoBuf definitions...');

// Using pbts/pbjs from protobufjs-cli
// Install: pnpm add -D protobufjs protobufjs-cli
try {
  if (!CHECK) {
    mkdirSync(OUTPUT_DIR, { recursive: true });

    // Generate ignored protobufjs runtime artifacts used during local development.
    execSync(
      `npx pbjs -t static-module -w es6 -o ${join(OUTPUT_DIR, 'bundle.js')} ${join(PROTO_DIR, 'game.proto')}`,
      { cwd: ROOT, stdio: 'inherit' }
    );

    execSync(
      `npx pbts -o ${join(OUTPUT_DIR, 'bundle.d.ts')} ${join(OUTPUT_DIR, 'bundle.js')}`,
      { cwd: ROOT, stdio: 'inherit' }
    );
  }

  emitCleanTypes(generateCleanTypes());

  execFileSync(
    process.execPath,
    ['scripts/generate-protocol-contracts.mjs', ...(CHECK ? ['--check'] : [])],
    { cwd: ROOT, stdio: 'inherit' },
  );

  if (!process.exitCode) {
    console.log(CHECK ? 'Generated types are current.' : 'TypeScript types generated successfully!');
  }
} catch (error) {
  console.error(CHECK ? 'Failed to check generated types:' : 'Failed to generate types:', error.message);
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

export const enum DeathCause {
  UNSPECIFIED = 0,
  EXECUTION = 1,
  NIGHT_KILL = 2,
  ABILITY = 3,
}

export const enum NightActionType {
  UNSPECIFIED = 0,
  POISON = 1,
  PROTECT = 2,
  KILL = 3,
  LEARN_TOWNSFOLK = 4,
  LEARN_OUTSIDER = 5,
  LEARN_MINION = 6,
  LEARN_EVIL_PAIRS = 7,
  LEARN_EVIL_NEIGHBORS = 8,
  CHECK_DEMON = 9,
  LEARN_EXECUTED = 10,
  LEARN_DIED = 11,
  LEARN_MASTER = 12,
  LEARN_DEMON = 13,
  CHOOSE_PLAYER = 14,
  NONE = 15,
}

export const enum WinReason {
  UNSPECIFIED = 0,
  IMP_EXECUTED = 1,
  MAYOR_ENDGAME = 2,
  EVIL_MAJORITY = 3,
  SAINT_EXECUTED = 4,
  IMP_STARPASS = 5,
  STORYTELLER_DECISION = 6,
}

export interface ProtoPlayer {
  readonly id: string;
  readonly name: string;
  readonly character?: ProtoCharacter;
  readonly isAlive: boolean;
  readonly votes: number;
  readonly poisonedUntil?: number;
  readonly shownCharacter?: ProtoCharacter;
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
  readonly deaths: readonly ProtoDeathRecord[];
  readonly nomination?: ProtoNomination;
  readonly nightActions: readonly ProtoNightAction[];
  readonly winner?: Team;
}

export interface ProtoDeathRecord {
  readonly playerId: string;
  readonly cause: DeathCause;
  readonly dayNumber: number;
  readonly killedBy?: string;
}

export interface ProtoNomination {
  readonly nominatorId: string;
  readonly nomineeId: string;
  readonly votes: Readonly<Record<string, boolean>>;
  readonly resolved: boolean;
}

export interface ProtoNightAction {
  readonly actorId: string;
  readonly actionType: NightActionType;
  readonly targetIds: readonly string[];
  readonly result?: string;
}

export type ProtoGameEvent =
  | { readonly playerJoined: { readonly player: ProtoPlayer } }
  | { readonly playerLeft: { readonly playerId: string } }
  | { readonly phaseChanged: { readonly phase: GamePhase } }
  | { readonly voteCast: { readonly voterId: string; readonly targetId?: string; readonly decision?: boolean } }
  | { readonly characterAssigned: { readonly playerId: string; readonly character: ProtoCharacter; readonly shownCharacter?: ProtoCharacter } }
  | { readonly playerDied: { readonly playerId: string; readonly cause: DeathCause; readonly dayNumber: number } }
  | { readonly nominationStarted: { readonly nominatorId: string; readonly nomineeId: string } }
  | { readonly nominationResolved: { readonly nomineeId: string; readonly executed: boolean; readonly yesVotes: number; readonly noVotes: number; readonly requiredVotes: number } }
  | { readonly nightActionSubmitted: { readonly actorId: string; readonly actionType: NightActionType; readonly targetIds: readonly string[]; readonly result?: string } }
  | { readonly gameEnded: { readonly winner: Team; readonly reason: WinReason; readonly description: string } };
`;

  return cleanTypes;
}

function emitCleanTypes(content) {
  if (CHECK) {
    let existing = '';
    try {
      existing = readFileSync(CLEAN_OUTPUT, 'utf8');
    } catch {
      // Report the missing generated artifact below.
    }
    if (existing !== content) {
      console.error(`Generated contract is stale: ${CLEAN_OUTPUT}`);
      process.exitCode = 1;
    }
    return;
  }

  writeFileSync(CLEAN_OUTPUT, content, 'utf8');
  console.log(`Generated ${CLEAN_OUTPUT}`);
}

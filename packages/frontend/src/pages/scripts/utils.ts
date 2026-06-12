import {
  getBaseRoleCount,
  getScriptWakeOrder,
  TROUBLE_BREWING_SCRIPT,
} from '@clocktower/core';
import type { CharacterType, ScriptCharacterDefinition } from '@clocktower/core';

export const DEFAULT_SCRIPT_ID = TROUBLE_BREWING_SCRIPT.id;

export const CHARACTER_TYPE_LABELS: Record<CharacterType, string> = {
  townsfolk: '镇民',
  outsider: '外来者',
  minion: '爪牙',
  demon: '恶魔',
};

export const CHARACTER_TYPE_ORDER: readonly CharacterType[] = ['townsfolk', 'outsider', 'minion', 'demon'];

export function charactersByType(type: CharacterType): readonly ScriptCharacterDefinition[] {
  return TROUBLE_BREWING_SCRIPT.characters.filter((character) => character.type === type);
}

export function roleCountText(playerCount: number): string {
  const count = getBaseRoleCount(playerCount);
  if (!count) return '';
  return `${playerCount}人：镇民${count.townsfolk} / 外来者${count.outsiders} / 爪牙${count.minions} / 恶魔${count.demons}`;
}

export function getFirstNightOrder() {
  return getScriptWakeOrder(DEFAULT_SCRIPT_ID, 1);
}

export function getLaterNightOrder() {
  return getScriptWakeOrder(DEFAULT_SCRIPT_ID, 2);
}

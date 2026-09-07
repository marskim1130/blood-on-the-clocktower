import { TROUBLE_BREWING_SCRIPT, type GameCharacter } from '@clocktower/core';

const CHARACTER_BY_ID = new Map(
  TROUBLE_BREWING_SCRIPT.characters.map((character) => [character.id, character]),
);

const ENGLISH_CHARACTER_NAMES: ReadonlyArray<readonly [string, string]> = [
  ['Fortune Teller', '占卜师'],
  ['Scarlet Woman', '红唇女郎'],
  ['Washerwoman', '洗衣妇'],
  ['Investigator', '调查员'],
  ['Ravenkeeper', '守鸦人'],
  ['Undertaker', '入殓师'],
  ['Librarian', '图书管理员'],
  ['Poisoner', '投毒者'],
  ['Recluse', '隐士'],
  ['Soldier', '士兵'],
  ['Empath', '共情者'],
  ['Butler', '管家'],
  ['Drunk', '酒鬼'],
  ['Baron', '男爵'],
  ['Slayer', '杀手'],
  ['Virgin', '圣女'],
  ['Mayor', '镇长'],
  ['Saint', '圣徒'],
  ['Monk', '僧侣'],
  ['Chef', '厨师'],
  ['Spy', '间谍'],
  ['Imp', '小恶魔'],
];

type CharacterSummary = Pick<GameCharacter, 'id' | 'name' | 'ability'>;

export function characterDisplayName(character?: CharacterSummary): string | undefined {
  if (!character) return undefined;
  return CHARACTER_BY_ID.get(character.id)?.name ?? character.name;
}

export function characterDisplayAbility(character?: CharacterSummary): string | undefined {
  if (!character) return undefined;
  return CHARACTER_BY_ID.get(character.id)?.ability ?? character.ability;
}

export function scriptDisplayName(scriptId: string, fallback: string): string {
  return scriptId === TROUBLE_BREWING_SCRIPT.id ? TROUBLE_BREWING_SCRIPT.name : fallback;
}

export function nightResultDisplay(result: string): string {
  const exactLabels: Readonly<Record<string, string>> = {
    yes: '是',
    no: '否',
    none: '无',
  };
  let display = exactLabels[result] ?? result;
  display = display
    .replace(/^Demon:\s*/, '恶魔：')
    .replace(/^Minions:\s*/, '爪牙：')
    .replace(/\|\s*Minions:\s*/, '｜爪牙：')
    .replace(/\|\s*Bluffs:\s*/, '｜伪装身份：');
  for (const [english, chinese] of ENGLISH_CHARACTER_NAMES) {
    display = display.replace(new RegExp(`\\b${english}\\b`, 'g'), chinese);
  }
  return display;
}

import { describe, expect, it } from 'vitest';
import {
  characterDisplayAbility,
  characterDisplayName,
  nightResultDisplay,
  scriptDisplayName,
} from './character-display';

describe('localized character display', () => {
  it('uses the local script copy for protocol characters', () => {
    const imp = { id: 'imp', name: 'Imp', ability: 'English ability' };

    expect(characterDisplayName(imp)).toBe('小恶魔');
    expect(characterDisplayAbility(imp)).not.toBe('English ability');
    expect(scriptDisplayName('trouble_brewing', 'Trouble Brewing')).toBe('暗流涌动');
  });

  it('keeps unknown future characters readable', () => {
    const future = { id: 'future', name: '未来角色', ability: '未来能力' };

    expect(characterDisplayName(future)).toBe('未来角色');
    expect(characterDisplayAbility(future)).toBe('未来能力');
  });

  it('localizes structured night information strings', () => {
    expect(nightResultDisplay('Demon: 测试玩家4 (Imp)')).toBe('恶魔：测试玩家4 (小恶魔)');
    expect(nightResultDisplay('Minions: 测试玩家5 (Poisoner)')).toBe('爪牙：测试玩家5 (投毒者)');
    expect(nightResultDisplay('yes')).toBe('是');
  });
});

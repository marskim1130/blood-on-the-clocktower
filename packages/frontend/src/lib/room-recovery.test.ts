import { describe, expect, it } from 'vitest';
import { encodeRecoveryCode, parseRecoveryCode } from './room-recovery';

describe('房间恢复码', () => {
  it('可在另一台设备还原完整房间身份', () => {
    const identity = { roomId: 'room-1', playerId: 'player-2', recoveryCredential: 'recovery-token' };
    expect(parseRecoveryCode(encodeRecoveryCode(identity))).toEqual({ ok: true, identity });
    expect(encodeRecoveryCode(identity).startsWith('CT3:')).toBe(true);
  });
  it('错误码返回固定错误，不抛出或回显包含凭据的输入', () => {
    const secret = 'TOP-SECRET';
    for (const code of ['', secret, `CT2:${secret}`, 'CT2:%7B%7D', 'CT3:%7B%7D', `CT2:${'a'.repeat(9000)}`]) {
      expect(parseRecoveryCode(code)).toEqual({ ok: false, error: '恢复码无效，请从原设备重新复制完整恢复码。' });
    }
  });
  it('导出排除同设备恢复凭据并拒绝旧 CT2 完整凭据', () => {
    const identity = { roomId: 'room', playerId: 'player', recoveryCredential: 'recovery', resumeCredential: 'DO-NOT-EXPORT' };
    expect(decodeURIComponent(encodeRecoveryCode(identity))).not.toContain('DO-NOT-EXPORT');
    expect(parseRecoveryCode(`CT2:${encodeURIComponent(JSON.stringify({ version: 2, ...identity }))}`).ok).toBe(false);
  });
});

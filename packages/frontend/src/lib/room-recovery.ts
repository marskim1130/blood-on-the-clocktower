export interface RecoveryIdentity {
  readonly roomId: string;
  readonly playerId: string;
  readonly recoveryCredential: string;
}

export function encodeRecoveryCode(identity: RecoveryIdentity): string {
  const payload = { roomId: identity.roomId, playerId: identity.playerId, recoveryCredential: identity.recoveryCredential };
  return `CT3:${encodeURIComponent(JSON.stringify(payload))}`;
}

export function parseRecoveryCode(code: string): { ok: true; identity: RecoveryIdentity } | { ok: false; error: string } {
  const invalid = { ok: false as const, error: '恢复码无效，请从原设备重新复制完整恢复码。' };
  const value = code.trim();
  if (!value.startsWith('CT3:') || value.length > 8192) return invalid;
  try {
    const identity = JSON.parse(decodeURIComponent(value.slice(4))) as Partial<RecoveryIdentity> | null;
    if (!identity ||
      typeof identity.roomId !== 'string' || !identity.roomId.trim() || identity.roomId.length > 128 ||
      typeof identity.playerId !== 'string' || !identity.playerId.trim() || identity.playerId.length > 128 ||
      typeof identity.recoveryCredential !== 'string' || !identity.recoveryCredential.trim() || identity.recoveryCredential.length > 4096) return invalid;
    return { ok: true, identity: { roomId: identity.roomId, playerId: identity.playerId, recoveryCredential: identity.recoveryCredential } };
  } catch {
    // Parsing errors may contain credential text; never propagate or log them.
    return invalid;
  }
}

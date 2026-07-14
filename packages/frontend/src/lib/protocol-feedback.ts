const ERROR_CODE_LABELS: Readonly<Record<string, string>> = {
  INVALID_MESSAGE: '请求格式无效，请重试',
  UNSUPPORTED_PROTOCOL: '客户端版本过旧，请更新应用',
  ROOM_NOT_FOUND: '房间不存在或已关闭',
  INVALID_CREDENTIAL: '房间身份已失效，请重新加入',
  STALE_CONNECTION: '当前连接已被新连接接管',
  FORBIDDEN: '没有权限执行该操作',
  ROOM_FULL: '房间人数已满，请稍后重试或更换房间',
  INVALID_COMMAND: '当前操作不符合游戏规则',
  PARTICIPANT_SET_FROZEN: '身份发放后不能更改参与者',
  UNEXPECTED_SEQUENCE: '操作序号不同步，正在获取最新状态',
  SEQUENCE_CONFLICT: '操作序号冲突，正在获取最新状态',
  IDEMPOTENCY_CONFLICT: '重复请求的参数不一致',
  PERSISTENCE_UNAVAILABLE: '服务器暂时无法保存状态，请稍后重试',
  PERSISTENCE_CONFLICT: '房间状态发生冲突，请重新进入',
  INTERNAL: '服务器内部错误，请稍后重试',
};

const ERROR_LABELS: Readonly<Record<string, string>> = {
  'room not found': '房间不存在或已关闭',
  'room is full': '房间人数已满，请稍后重试或更换房间',
  'storyteller already set': '说书人已经设置',
  'target player not found': '目标玩家不存在',
  'only storyteller can assign characters': '只有说书人可以发放身份',
  'invalid character assignment for player count': '角色配置与当前玩家人数不匹配',
  'characters have already been assigned': '身份已经发放，不能再次修改',
  'storyteller must be set before starting the game': '开始游戏前必须设置说书人',
  'only the storyteller can start the game': '只有说书人可以开始游戏',
  'nominations can only happen during the day phase': '只能在白天发起提名',
  'cannot nominate yourself': '不能提名自己',
  'dead players cannot nominate': '死亡玩家不能提名',
  'cannot nominate a dead player': '不能提名死亡玩家',
  'no active nomination to vote on': '当前没有进行中的投票',
  'ghost vote already used': '你的幽灵票已经使用',
  'only the storyteller can resolve a nomination': '只有说书人可以结算投票',
  'night actions can only be submitted during the night phase': '只能在夜晚提交夜间行动',
  'no remaining night wake steps': '今夜没有剩余唤醒步骤',
  'only the storyteller can resolve the night': '只有说书人可以结束夜晚',
  'game is already finished': '游戏已经结束',
};

export function protocolErrorMessage(error?: string, code?: string): string {
  if (error && ERROR_LABELS[error]) return ERROR_LABELS[error];
  if (code && ERROR_CODE_LABELS[code]) return ERROR_CODE_LABELS[code];
  if (!error) return '操作失败，请稍后重试';

  if (/^player .+ not found(?: in game)?$/.test(error)) return '目标玩家不存在';
  if (/^player .+ has already voted$/.test(error)) return '你已经完成本轮投票';
  if (/^player .+ is already dead$/.test(error)) return '该玩家已经死亡';
  if (/^night action .+ requires at least .+ target\(s\)$/.test(error)) return '选择的目标数量不足';
  if (/^night action .+ allows at most .+ target\(s\)$/.test(error)) return '选择的目标数量超过限制';
  if (/^expected night action .+, got .+$/.test(error)) return '夜间步骤已变化，请按当前步骤重新提交';
  if (/^cannot resolve night with .+ wake step\(s\) remaining$/.test(error)) return '仍有唤醒步骤未完成';
  return '操作失败，请稍后重试';
}

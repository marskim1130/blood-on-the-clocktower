import React from 'react';
import { Button, ScrollView, Text, View } from '@tarojs/components';
import Taro from '@tarojs/taro';
import type { RoomState } from '@clocktower/core';

const ACTIONS: Readonly<Record<string, string>> = {
  SetReady: '变更准备状态', ConfirmCharacter: '确认身份', AssignCharacters: '发放身份', StartGame: '开始游戏',
  SubmitNightAction: '提交夜间行动', ConfirmNightAction: '确认夜间信息', AcknowledgeNightAction: '夜间已阅',
  SkipNightAction: '跳过夜间步骤', PrepareDawn: '生成黎明建议', ConfirmDawn: '确认黎明死亡',
  Nominate: '发起提名', CastVote: '玩家投票', RecordVote: '说书人代投', ResolveNomination: '结算投票',
  AdvanceNominationStage: '推进控辩阶段', ControlNominationTimer: '调整投票计时', ExpireNominationTimer: '计时到期',
  FinalizeDay: '确认日终处决', ChangePhase: '变更阶段', KillPlayer: '记录死亡', ExecutePlayer: '执行处决',
  UseSlayerAbility: '猎手行动', EndGame: '判定胜负', PublishGrimoire: '公开魔典', UndoGame: '撤销操作', RedoGame: '重做操作',
};
const PHASES: Readonly<Record<number, string>> = { 1: '准备', 2: '白天', 3: '夜晚', 4: '投票', 5: '结束' };

interface Props {
  readonly room: RoomState;
  readonly actorId: string;
  readonly busy: boolean;
  readonly onUndo: (confirmPhaseChange: boolean) => void;
  readonly onRedo: (confirmPhaseChange: boolean) => void;
}

export function GameHistoryPanel({ room, actorId, busy, onUndo, onRedo }: Props): React.ReactElement | null {
  if (!room.storytellerId || actorId !== room.storytellerId) return null;
  async function confirm(redo: boolean): Promise<void> {
    const crossesPhase = Boolean(redo ? room.redoCrossesPhase : room.undoCrossesPhase);
    const answer = await Taro.showModal({
      title: crossesPhase ? '确认跨阶段回退？' : redo ? '重做上一步？' : '撤销上一步？',
      content: `${crossesPhase ? '此操作会切换游戏阶段。' : ''}已经展示的身份、夜间信息和投票结果无法从玩家记忆中撤回。恢复投票会暂停计时，房间成员和身份凭据不会回滚。`,
      confirmText: redo ? '确认重做' : '确认撤销',
    });
    if (answer.confirm) (redo ? onRedo : onUndo)(crossesPhase);
  }
  return (
    <View className='sectionBand'>
      <Text className='sectionHeading'>说书人操作记录</Text>
      <Text className='sectionDescription'>仅你可见。保留最近 100 条记录，最多撤回 30 步；成员和座位变更会建立新的历史边界。</Text>
      <View className='commandRow'>
        <Button className='secondaryButton' disabled={busy || !room.canUndo} onClick={() => { void confirm(false); }}>撤销上一步</Button>
        <Button className='secondaryButton' disabled={busy || !room.canRedo} onClick={() => { void confirm(true); }}>重做上一步</Button>
      </View>
      <ScrollView scrollY style={{ maxHeight: '300px' }}>
        {(room.operationLog ?? []).slice().reverse().map((entry) => (
          <View className='playerRow' key={entry.id}>
            <Text className='seatNumber'>{entry.id}</Text>
            <View className='playerBody'>
              <Text className='playerName'>{ACTIONS[entry.action] ?? entry.action}</Text>
              <Text className='playerMeta'>{entry.actorId === room.storytellerId ? '说书人' : room.players.find(player => player.id === entry.actorId)?.name ?? '玩家'} · {PHASES[entry.fromPhase] ?? '准备'} → {PHASES[entry.toPhase] ?? '准备'}</Text>
            </View>
          </View>
        ))}
      </ScrollView>
      {!(room.operationLog?.length) && <Text className='sectionDescription'>本局尚无操作记录。</Text>}
    </View>
  );
}

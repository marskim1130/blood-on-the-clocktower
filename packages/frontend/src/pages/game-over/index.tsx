import { Button, Text, View } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { useMemo } from 'react';
import { LoadingState, SessionShell } from '../../components/session-shell';
import { characterDisplayAbility, characterDisplayName } from '../../lib/character-display';
import { normalizeWinner, selectGameOver } from '../../lib/room-experience';
import { useRoomSession } from '../../lib/room-session-store';
import { SESSION_ROUTES } from '../../lib/session-routing';
import { useSessionRoute } from '../../lib/use-session-route';
import './index.css';

const REASON_LABELS: Readonly<Record<string, string>> = {
  imp_executed: '小恶魔被处决',
  mayor_endgame: '镇长存活到最终三人',
  evil_majority: '邪恶阵营控制了存活人数',
  saint_executed: '圣徒被处决',
  imp_starpass: '小恶魔自杀后无人继承',
  storyteller_decision: '说书人裁定',
};

const DEATH_CAUSE_LABELS: Readonly<Record<string, string>> = {
  execution: '处决',
  night_kill: '夜间死亡',
  ability: '能力致死',
};

const DESCRIPTION_LABELS: Readonly<Record<string, string>> = {
  'Storyteller ended the game.': '说书人结束了本局游戏。',
  'The Saint was executed — evil wins!': '圣徒被处决，邪恶阵营获胜。',
  'The Demon is dead — good wins!': '恶魔死亡，善良阵营获胜。',
  'Only two players remain alive — evil wins!': '仅剩两名存活玩家，邪恶阵营获胜。',
  'Only 3 players remain with no execution — Mayor wins for good!': '最终三人无人被处决，镇长令善良阵营获胜。',
};

function teamLabel(team: number): string {
  if (team === 1) return '善良';
  if (team === 2) return '邪恶';
  return '未知阵营';
}

export default function GameOverPage() {
  const experience = useRoomSession((state) => state.experience);
  const room = experience.roomState;
  const identity = useRoomSession((state) => state.identity);
  const playerId = useRoomSession((state) => state.playerId);
  const pendingCommand = useRoomSession((state) => state.pendingCommand);
  const leaveRoom = useRoomSession((state) => state.leaveRoom);
  const closeRoom = useRoomSession((state) => state.closeRoom);
  const gameOver = selectGameOver(experience);

  useSessionRoute(SESSION_ROUTES.over);

  const deaths = useMemo(
    () => new Map(room?.deaths?.map((death) => [death.playerId, death]) ?? []),
    [room],
  );

  if (!room) {
    return (
      <SessionShell eyebrow='游戏结果' title='恢复最终结果'>
        {identity ? <LoadingState label='正在恢复最终投影...' /> : <LoadingState label='正在返回大厅...' />}
      </SessionShell>
    );
  }

  const isCreator = room.creatorId === playerId;
  const busy = pendingCommand !== null;
  const winner = gameOver ? normalizeWinner(gameOver.winner) : null;
  const winnerLabel = winner === 'good' ? '善良阵营获胜' : winner === 'evil' ? '邪恶阵营获胜' : '结果不可用';
  const reasonLabel = gameOver ? REASON_LABELS[gameOver.reason] ?? `未知原因：${gameOver.reason}` : '服务器未提供胜负结果';
  const resultDescription = gameOver?.description
    ? DESCRIPTION_LABELS[gameOver.description] ?? gameOver.description
    : '';

  function confirmExit(): void {
    void Taro.showModal({
      title: isCreator ? '关闭房间？' : '离开房间？',
      content: isCreator
        ? '关闭后全员将失去该房间的恢复入口。'
        : '离开会撤销此设备的房间凭证，其他成员仍可查看最终结果。',
      confirmText: isCreator ? '关闭房间' : '确认离开',
      confirmColor: '#7b2028',
    }).then((result) => {
      if (!result.confirm) return;
      if (isCreator) closeRoom(); else leaveRoom();
    });
  }

  return (
    <SessionShell
      eyebrow='游戏结果'
      title='真相揭晓'
      actions={
        <Button className={isCreator ? 'dangerButton' : 'commandButton'} disabled={busy} onClick={confirmExit}>
          {isCreator ? '关闭房间' : '离开房间'}
        </Button>
      }
    >
      <View className={`resultBanner ${gameOver ? '' : 'resultMissing'}`}>
        <Text className='resultTeam'>{winnerLabel}</Text>
        <Text className='resultReason'>{reasonLabel}</Text>
        {resultDescription && <Text className='resultDescription'>{resultDescription}</Text>}
      </View>

      <View className='sectionBand'>
        <Text className='sectionHeading'>身份揭示</Text>
        <View className='revealList'>
          {room.players.map((player, index) => {
            const death = deaths.get(player.id);
            return (
              <View className='revealItem' key={player.id}>
                <View className='revealHeader'>
                  <Text className='playerName'>{String(index + 1).padStart(2, '0')} · {player.name || `座位 ${index + 1}`}</Text>
                  <Text className={`tag ${player.character?.team === 2 ? 'tagEvil' : 'tagGood'}`}>{teamLabel(player.character?.team ?? 0)}</Text>
                </View>
                <Text className='revealRole'>{characterDisplayName(player.character) ?? '身份缺失'}</Text>
                {player.shownCharacter && <Text className='playerMeta'>曾展示为：{characterDisplayName(player.shownCharacter)}</Text>}
                {player.character?.ability && <Text className='revealAbility'>{characterDisplayAbility(player.character)}</Text>}
                <Text className='deathCause'>
                  {death ? `第 ${death.dayNumber} 天 · ${DEATH_CAUSE_LABELS[death.cause] ?? death.cause}` : '存活至游戏结束'}
                </Text>
              </View>
            );
          })}
        </View>
      </View>

      <View className='sectionBand'>
        <Text className='sectionHeading'>死亡时间线</Text>
        {room.deaths && room.deaths.length > 0 ? room.deaths.map((death) => {
          const player = room.players.find((candidate) => candidate.id === death.playerId);
          return (
            <View className='playerRow' key={`${death.dayNumber}-${death.playerId}`}>
              <Text className='seatNumber'>{death.dayNumber}</Text>
              <View className='playerBody'>
                <Text className='playerName'>{player?.name ?? '未知玩家'}</Text>
                <Text className='playerMeta'>{DEATH_CAUSE_LABELS[death.cause] ?? death.cause}</Text>
              </View>
            </View>
          );
        }) : (
          <Text className='sectionDescription'>本局没有死亡记录。</Text>
        )}
      </View>
    </SessionShell>
  );
}

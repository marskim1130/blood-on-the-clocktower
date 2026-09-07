import { Button, Text, View } from '@tarojs/components';
import { useMemo, useState } from 'react';
import type { RoomState } from '@clocktower/core';
import { moveSeatClockwise } from './utils';

type SeatPlayer = RoomState['players'][number];

interface SeatOrderEditorProps {
  readonly players: readonly SeatPlayer[];
  readonly currentPlayerId: string;
  readonly busy: boolean;
  readonly onSubmit: (seatOrder: readonly string[]) => void;
}

export function SeatOrderEditor({
  players,
  currentPlayerId,
  busy,
  onSubmit,
}: SeatOrderEditorProps) {
  const authoritativeSeatOrder = players.map((player) => player.id);
  const [draftSeatOrder, setDraftSeatOrder] = useState<string[]>(() => authoritativeSeatOrder);
  const playersById = useMemo(
    () => new Map(players.map((player) => [player.id, player])),
    [players],
  );
  const hasChanges = draftSeatOrder.some(
    (playerId, index) => playerId !== authoritativeSeatOrder[index],
  );

  function resetDraft(): void {
    setDraftSeatOrder([...authoritativeSeatOrder]);
  }

  function submitDraft(): void {
    if (!hasChanges || busy) return;
    onSubmit([...draftSeatOrder]);
  }

  return (
    <View className='seatOrderEditor'>
      <Text className='sectionDescription'>
        列表从 01 号位开始按顺时针排列。逐次移动玩家，确认无误后一次提交完整座次。
      </Text>
      <View className='playerList seatOrderDraftList'>
        {draftSeatOrder.map((draftPlayerId, index) => {
          const player = playersById.get(draftPlayerId);
          const playerName = player?.name || `座位 ${index + 1}`;
          return (
            <View className='playerRow seatOrderDraftRow' key={draftPlayerId}>
              <Text className='seatNumber'>{String(index + 1).padStart(2, '0')}</Text>
              <View className='playerBody'>
                <Text className='playerName'>
                  {playerName}{draftPlayerId === currentPlayerId ? '（你）' : ''}
                </Text>
                <Text className='playerMeta'>本机座次草稿</Text>
              </View>
              <Button
                className='quietButton seatMoveButton'
                aria-label={`${playerName}顺时针移动一席`}
                disabled={busy}
                onClick={() => setDraftSeatOrder((current) => moveSeatClockwise(current, draftPlayerId))}
              >
                <Text className='buttonLabelDark'>顺时针一席</Text>
              </Button>
            </View>
          );
        })}
      </View>
      <Text className={`seatOrderStatus ${hasChanges ? 'seatOrderStatusPending' : ''}`}>
        {hasChanges
          ? (busy ? '正在等待服务器确认座次…' : '当前调整尚未提交，仅你可见。')
          : '当前草稿与服务器座次一致。'}
      </Text>
      <View className='seatOrderActions'>
        <Button className='quietButton' disabled={busy || !hasChanges} onClick={resetDraft}>
          <Text className='buttonLabelDark'>还原</Text>
        </Button>
        <Button className='commandButton' disabled={busy || !hasChanges} onClick={submitDraft}>
          确认座次
        </Button>
      </View>
    </View>
  );
}

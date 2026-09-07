import { Button, Text, View } from '@tarojs/components';
import { useDidHide } from '@tarojs/taro';
import { useCallback, useEffect, useRef, useState } from 'react';
import type { GameCharacter } from '@clocktower/core';
import { characterDisplayAbility, characterDisplayName } from '../lib/character-display';
import {
  PRIVATE_CHARACTER_REVEAL_MS,
  concealPrivateCharacter,
  createPrivateCharacterVisibility,
  revealPrivateCharacter,
} from './private-character-visibility';
import './private-character-card.css';

export interface PrivateCharacterCardProps {
  readonly character: GameCharacter;
  readonly roomId: string;
  readonly playerName: string;
  readonly confirmed?: boolean;
  readonly confirmPending?: boolean;
  readonly onConfirm?: () => void;
}

export function PrivateCharacterCard({
  character,
  roomId,
  playerName,
  confirmed = false,
  confirmPending = false,
  onConfirm,
}: PrivateCharacterCardProps) {
  const [visibility, setVisibility] = useState(createPrivateCharacterVisibility);
  const concealTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const clearConcealTimer = useCallback(() => {
    if (concealTimer.current === null) return;
    clearTimeout(concealTimer.current);
    concealTimer.current = null;
  }, []);

  const conceal = useCallback(() => {
    clearConcealTimer();
    setVisibility((current) => concealPrivateCharacter(current));
  }, [clearConcealTimer]);

  const reveal = useCallback(() => {
    clearConcealTimer();
    setVisibility((current) => revealPrivateCharacter(current, Date.now()));
    concealTimer.current = setTimeout(conceal, PRIVATE_CHARACTER_REVEAL_MS);
  }, [clearConcealTimer, conceal]);

  useDidHide(conceal);

  useEffect(() => {
    clearConcealTimer();
    setVisibility(createPrivateCharacterVisibility());
    return clearConcealTimer;
  }, [character.id, clearConcealTimer]);

  return (
    <View className='privateCharacterCard'>
      <Text className='privateCharacterNotice'>身份属于私密信息，请避开其他玩家屏幕。</Text>
      <View
        className={`privateCharacterReveal ${visibility.revealed ? 'privateCharacterRevealOpen' : ''}`}
        hoverClass='privateCharacterRevealPressed'
        ariaLabel='按住查看你的身份，松手立即隐藏'
        onTouchStart={reveal}
        onTouchEnd={conceal}
        onTouchCancel={conceal}
      >
        {visibility.revealed ? (
          <View className='privateCharacterSecret'>
            <View className='privateCharacterWatermarks'>
              {[0, 1, 2].map((item) => (
                <Text className='privateCharacterWatermark' key={item}>{roomId} · {playerName}</Text>
              ))}
            </View>
            <Text className='privateCharacterRole'>{characterDisplayName(character)}</Text>
            <Text className='privateCharacterAbility'>{characterDisplayAbility(character)}</Text>
          </View>
        ) : (
          <View className='privateCharacterMask'>
            <Text className='privateCharacterMaskTitle'>按住查看身份</Text>
            <Text className='privateCharacterMaskHint'>松手、切到后台或 30 秒后立即隐藏</Text>
          </View>
        )}
      </View>

      {onConfirm && (
        <View className='privateCharacterConfirm'>
          <Button
            className={confirmed ? 'secondaryButton fullWidthButton' : 'commandButton fullWidthButton'}
            disabled={confirmed || confirmPending || !visibility.hasRevealed}
            onClick={onConfirm}
          >
            {confirmed ? '身份已确认' : confirmPending ? '等待服务器确认…' : '我已记住身份'}
          </Button>
          {!visibility.hasRevealed && !confirmed && (
            <Text className='privateCharacterConfirmHint'>至少按住查看一次后才能确认。</Text>
          )}
        </View>
      )}
    </View>
  );
}

import { Button, Text, View } from '@tarojs/components';
import { useDidHide } from '@tarojs/taro';
import React, { useCallback, useEffect, useRef } from 'react';
import { PRIVATE_CHARACTER_REVEAL_MS, concealPrivateCharacter, createPrivateCharacterVisibility, revealPrivateCharacter } from './private-character-visibility';
import './private-character-card.css';

interface PrivateNightResultProps {
  readonly resultKey: string;
  readonly roomId: string;
  readonly playerName: string;
  readonly targets?: string;
  readonly result: string;
  readonly busy?: boolean;
  readonly onAcknowledge: () => void;
}

export function PrivateNightResult({ resultKey, roomId, playerName, targets, result, busy, onAcknowledge }: PrivateNightResultProps) {
  const [visibility, setVisibility] = React.useState(createPrivateCharacterVisibility);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const clearTimer = useCallback(() => {
    if (timer.current !== null) clearTimeout(timer.current);
    timer.current = null;
  }, []);
  const conceal = useCallback(() => {
    clearTimer();
    setVisibility(concealPrivateCharacter);
  }, [clearTimer]);
  const reveal = useCallback(() => {
    clearTimer();
    setVisibility((state) => revealPrivateCharacter(state, Date.now()));
    timer.current = setTimeout(conceal, PRIVATE_CHARACTER_REVEAL_MS);
  }, [clearTimer, conceal]);
  useDidHide(conceal);
  useEffect(() => {
    clearTimer();
    setVisibility(createPrivateCharacterVisibility());
    return clearTimer;
  }, [resultKey, result, targets, clearTimer]);

  return (
    <View className='privateCharacterCard'>
      <Text className='privateCharacterNotice'>说书人已确认，请避开其他玩家屏幕。</Text>
      <View className={`privateCharacterReveal ${visibility.revealed ? 'privateCharacterRevealOpen' : ''}`}
        ariaLabel='按住查看夜间信息，松手立即隐藏'
        onTouchStart={reveal} onTouchEnd={conceal} onTouchCancel={conceal}>
        {visibility.revealed ? (
          <View className='privateCharacterSecret'>
            <View className='privateCharacterWatermarks'>
              {[0, 1, 2].map((item) => <Text className='privateCharacterWatermark' key={item}>{roomId} · {playerName}</Text>)}
            </View>
            {targets && <Text className='privateCharacterAbility'>目标：{targets}</Text>}
            <Text className='privateCharacterAbility'>{result}</Text>
          </View>
        ) : (
          <View className='privateCharacterMask'>
            <Text className='privateCharacterMaskTitle'>按住查看夜间信息</Text>
            <Text className='privateCharacterMaskHint'>松手、切到后台或 30 秒后立即隐藏</Text>
          </View>
        )}
      </View>
      <Button className='commandButton fullWidthButton' disabled={busy || !visibility.hasRevealed}
        onClick={() => { conceal(); onAcknowledge(); }}>我已阅，闭眼</Button>
    </View>
  );
}

import { Button, ScrollView, Text, View } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { TROUBLE_BREWING_SCRIPT } from '@clocktower/core';
import type { NightWakeStep } from '@clocktower/core';
import {
  CHARACTER_TYPE_LABELS,
  CHARACTER_TYPE_ORDER,
  charactersByType,
  roleCountText,
  getFirstNightOrder,
  getLaterNightOrder,
} from './utils';
import './index.css';

function closePage(): void {
  void Taro.navigateBack();
}

function wakeStepName(step: NightWakeStep): string {
  const character = TROUBLE_BREWING_SCRIPT.characters.find((item) => item.id === step.characterId);
  if (character) return character.name;
  if (step.characterType) return CHARACTER_TYPE_LABELS[step.characterType] ?? step.characterType;
  return step.characterId;
}

function wakeStepKey(prefix: string, step: NightWakeStep): string {
  return `${prefix}-${step.order}-${step.characterId || step.characterType || step.actionType}`;
}

export default function ScriptsPage() {
  const firstNight = getFirstNightOrder();
  const laterNight = getLaterNightOrder();

  return (
    <ScrollView className='scriptPage' scrollY>
      <View className='scriptHero'>
        <Text className='scriptTitle'>{TROUBLE_BREWING_SCRIPT.name}</Text>
        <Text className='scriptSubtitle'>角色目录、合法分布与夜晚唤醒顺序</Text>
      </View>

      <View className='scriptSection'>
        <Text className='sectionTitle'>玩家人数</Text>
        {[5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15].map((playerCount) => (
          <Text className='roleCountLine' key={playerCount}>{roleCountText(playerCount)}</Text>
        ))}
        <Text className='scriptHint'>男爵在场时，减少 2 个镇民并增加 2 个外来者。</Text>
      </View>

      <View className='scriptSection'>
        <Text className='sectionTitle'>角色</Text>
        {CHARACTER_TYPE_ORDER.map((type) => (
          <View className='characterGroup' key={type}>
            <Text className='groupTitle'>{CHARACTER_TYPE_LABELS[type]}</Text>
            {charactersByType(type).map((character) => (
              <View className='characterRow' key={character.id}>
                <Text className='characterName'>{character.name}</Text>
                <Text className='characterAbility'>{character.ability}</Text>
              </View>
            ))}
          </View>
        ))}
      </View>

      <View className='scriptSection'>
        <Text className='sectionTitle'>首夜顺序</Text>
        {firstNight.map((step) => (
          <View className='wakeRow' key={wakeStepKey('first', step)}>
            <Text className='wakeOrder'>{step.order}</Text>
            <View className='wakeInfo'>
              <Text className='wakeName'>{wakeStepName(step)}</Text>
              <Text className='scriptHint'>{step.prompt}</Text>
            </View>
          </View>
        ))}
      </View>

      <View className='scriptSection'>
        <Text className='sectionTitle'>后续夜晚</Text>
        {laterNight.map((step) => (
          <View className='wakeRow' key={wakeStepKey('later', step)}>
            <Text className='wakeOrder'>{step.order}</Text>
            <View className='wakeInfo'>
              <Text className='wakeName'>{wakeStepName(step)}</Text>
              <Text className='scriptHint'>{step.prompt}</Text>
            </View>
          </View>
        ))}
      </View>

      <Button className='backButton' onClick={closePage}>返回房间</Button>
    </ScrollView>
  );
}

import { Button, ScrollView, Text, View } from '@tarojs/components';
import Taro from '@tarojs/taro';
import {
  getBaseRoleCount,
  getScriptWakeOrder,
  TROUBLE_BREWING_SCRIPT,
} from '@clocktower/core';
import type { CharacterType, ScriptCharacterDefinition } from '@clocktower/core';
import './index.css';

const DEFAULT_SCRIPT_ID = TROUBLE_BREWING_SCRIPT.id;

const CHARACTER_TYPE_LABELS: Record<CharacterType, string> = {
  townsfolk: '镇民 Townsfolk',
  outsider: '外来者 Outsider',
  minion: '爪牙 Minion',
  demon: '恶魔 Demon',
};

const CHARACTER_TYPE_ORDER: readonly CharacterType[] = ['townsfolk', 'outsider', 'minion', 'demon'];

function charactersByType(type: CharacterType): readonly ScriptCharacterDefinition[] {
  return TROUBLE_BREWING_SCRIPT.characters.filter((character) => character.type === type);
}

function roleCountText(playerCount: number): string {
  const count = getBaseRoleCount(playerCount);
  if (!count) return '';
  return `${playerCount}人：镇民${count.townsfolk} / 外来者${count.outsiders} / 爪牙${count.minions} / 恶魔${count.demons}`;
}

function closePage(): void {
  void Taro.navigateBack();
}

export default function ScriptsPage() {
  const firstNight = getScriptWakeOrder(DEFAULT_SCRIPT_ID, 1);
  const laterNight = getScriptWakeOrder(DEFAULT_SCRIPT_ID, 2);

  return (
    <ScrollView className='scriptPage' scrollY>
      <View className='scriptHero'>
        <Text className='scriptTitle'>{TROUBLE_BREWING_SCRIPT.name}</Text>
        <Text className='scriptSubtitle'>角色目录、合法分布与夜晚唤醒顺序</Text>
      </View>

      <View className='scriptSection'>
        <Text className='sectionTitle'>玩家人数 [Player Count]</Text>
        {[5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15].map((playerCount) => (
          <Text className='roleCountLine' key={playerCount}>{roleCountText(playerCount)}</Text>
        ))}
        <Text className='scriptHint'>男爵 [Baron] 在场时，减少 2 个镇民并增加 2 个外来者。</Text>
      </View>

      <View className='scriptSection'>
        <Text className='sectionTitle'>角色 [Characters]</Text>
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
        <Text className='sectionTitle'>首夜顺序 [First Night]</Text>
        {firstNight.map((step) => (
          <View className='wakeRow' key={`first-${step.order}-${step.characterId}`}>
            <Text className='wakeOrder'>{step.order}</Text>
            <View className='wakeInfo'>
              <Text className='wakeName'>
                {TROUBLE_BREWING_SCRIPT.characters.find((character) => character.id === step.characterId)?.name ?? step.characterId}
              </Text>
              <Text className='scriptHint'>{step.prompt}</Text>
            </View>
          </View>
        ))}
      </View>

      <View className='scriptSection'>
        <Text className='sectionTitle'>后续夜晚 [Later Nights]</Text>
        {laterNight.map((step) => (
          <View className='wakeRow' key={`later-${step.order}-${step.characterId}`}>
            <Text className='wakeOrder'>{step.order}</Text>
            <View className='wakeInfo'>
              <Text className='wakeName'>
                {TROUBLE_BREWING_SCRIPT.characters.find((character) => character.id === step.characterId)?.name ?? step.characterId}
              </Text>
              <Text className='scriptHint'>{step.prompt}</Text>
            </View>
          </View>
        ))}
      </View>

      <Button className='backButton' onClick={closePage}>返回房间</Button>
    </ScrollView>
  );
}

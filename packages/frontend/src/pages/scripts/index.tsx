import { Button, Input, Picker, Text, View } from '@tarojs/components';
import { useState } from 'react';
import { TROUBLE_BREWING_SCRIPT, type NightWakeStep } from '@clocktower/core';
import { SessionShell } from '../../components/session-shell';
import { CHARACTER_TYPE_LABELS, CHARACTER_TYPE_ORDER, charactersByType, roleCountText, getFirstNightOrder, getLaterNightOrder } from './utils';
import './index.css';

function wakeStepName(step: NightWakeStep): string {
  const character = TROUBLE_BREWING_SCRIPT.characters.find((item) => item.id === step.characterId);
  if (character) return character.name;
  if (step.characterType) return CHARACTER_TYPE_LABELS[step.characterType] ?? step.characterType;
  return step.characterId;
}

export default function ScriptsPage() {
  const [section, setSection] = useState<'characters' | 'night' | 'counts'>('characters');
  const [query, setQuery] = useState('');
  const [typeIndex, setTypeIndex] = useState(0);
  const search = query.trim().toLowerCase();
  const visibleGroups = CHARACTER_TYPE_ORDER.filter((type) => typeIndex === 0 || CHARACTER_TYPE_ORDER[typeIndex - 1] === type)
    .map((type) => ({ type, characters: charactersByType(type).filter((character) => `${character.name} ${character.ability} ${character.id}`.toLowerCase().includes(search)) }))
    .filter((group) => group.characters.length > 0);

  return (
    <SessionShell title={TROUBLE_BREWING_SCRIPT.name} eyebrow='角色图鉴' section='library' description='认识每一个角色，找到属于你的线索。'>
      <View className='pageTabs'>
        <Button className={`pageTab ${section === 'characters' ? 'pageTabActive' : ''}`} onClick={() => setSection('characters')}>角色图鉴</Button>
        <Button className={`pageTab ${section === 'night' ? 'pageTabActive' : ''}`} onClick={() => setSection('night')}>夜晚顺序</Button>
        <Button className={`pageTab ${section === 'counts' ? 'pageTabActive' : ''}`} onClick={() => setSection('counts')}>人数配置</Button>
      </View>

      {section === 'characters' && <View>
        <View className='catalogToolbar'>
          <Input className='catalogSearch' value={query} placeholder='搜索角色或能力关键词' aria-label='搜索角色' onInput={(event) => setQuery(event.detail.value)} />
          <Picker mode='selector' range={['全部角色', ...CHARACTER_TYPE_ORDER.map((type) => CHARACTER_TYPE_LABELS[type])]} value={typeIndex} onChange={(event) => setTypeIndex(Number(event.detail.value))}>
            <Button className='secondaryButton'>{typeIndex === 0 ? '全部角色' : CHARACTER_TYPE_LABELS[CHARACTER_TYPE_ORDER[typeIndex - 1]!]} ▾</Button>
          </Picker>
        </View>
        <Text className='catalogCount'>{visibleGroups.reduce((count, group) => count + group.characters.length, 0)} 个角色 · 暗流涌动</Text>
        {visibleGroups.map(({ type, characters }) => <View className='characterGroup' key={type}>
          <View className='catalogGroupHeading'><Text className='groupTitle'>{CHARACTER_TYPE_LABELS[type]}</Text><Text className='tag'>{characters.length}</Text></View>
          <View className='characterCardGrid'>{characters.map((character) => <View className={`characterCard characterCard-${type}`} key={character.id}>
            <View className='characterCardHeading'><Text className='characterToken'>{character.name.slice(0, 1)}</Text><View><Text className='characterName'>{character.name}</Text><Text className='characterType'>{CHARACTER_TYPE_LABELS[type]}</Text></View></View>
            <Text className='characterAbility'>{character.ability}</Text>
          </View>)}</View>
        </View>)}
        {visibleGroups.length === 0 && <View className='emptyBand'><Text className='emptyBandTitle'>没有找到匹配角色</Text><Text className='emptyBandText'>换一个关键词，或切换到全部角色。</Text></View>}
      </View>}

      {section === 'counts' && <View className='sectionBand'>
        <Text className='sectionHeading'>为这一桌选人数</Text>
        <Text className='sectionDescription'>玩家人数不包含说书人。</Text>
        <View className='roleCountGrid'>{[5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15].map((count) => <View className='roleCountCard' key={count}><Text className='roleCountLine'>{roleCountText(count)}</Text></View>)}</View>
        <Text className='scriptHint'>男爵在场时，减少 2 个镇民并增加 2 个外来者。</Text>
      </View>}

      {section === 'night' && <View className='nightOrderColumns'>
        {[{ title: '首夜顺序', steps: getFirstNightOrder(), key: 'first' }, { title: '后续夜晚', steps: getLaterNightOrder(), key: 'later' }].map((night) => <View className='sectionBand' key={night.key}>
          <Text className='sectionHeading'>{night.title}</Text>
          {night.steps.map((step) => <View className='wakeRow' key={`${night.key}-${step.order}-${step.characterId || step.characterType || step.actionType}`}>
            <Text className='wakeOrder'>{step.order}</Text><View className='wakeInfo'><Text className='wakeName'>{wakeStepName(step)}</Text><Text className='scriptHint'>{step.prompt}</Text></View>
          </View>)}
        </View>)}
      </View>}
    </SessionShell>
  );
}

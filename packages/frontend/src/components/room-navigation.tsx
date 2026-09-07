import React from 'react';
import { Button, Text, View } from '@tarojs/components';

export type ShellDestination = 'main' | 'library' | 'history' | 'tools';
interface Props {
  readonly active: ShellDestination;
  readonly hasRoom: boolean;
  readonly approvalCount: number;
  readonly onSelect: (destination: ShellDestination) => void;
}
export function RoomNavigation({ active, hasRoom, approvalCount, onSelect }: Props): React.ReactElement {
  const items = [
    { id: 'main' as const, label: hasRoom ? '主桌' : '大厅', icon: '⌂' },
    { id: 'library' as const, label: '图鉴', icon: '♜' },
    { id: 'history' as const, label: '记录', icon: '≡' },
    { id: 'tools' as const, label: '更多', icon: '⋯' },
  ];
  return <View className='bottomNav'>
    {items.map(item => <Button key={item.id} className={`navItem ${active === item.id ? 'navItemActive' : ''}`} aria-pressed={active === item.id} onClick={() => onSelect(item.id)}>
      <Text className='navIcon'>{item.icon}</Text><Text className='navLabel'>{item.label}</Text>
      {item.id === 'tools' && approvalCount > 0 && <Text className='navBadge'>{approvalCount}</Text>}
    </Button>)}
  </View>;
}

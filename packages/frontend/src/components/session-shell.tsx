import { Button, ScrollView, Text, View } from '@tarojs/components';
import type { PropsWithChildren, ReactNode } from 'react';
import { scriptDisplayName } from '../lib/character-display';
import { useRoomSession } from '../lib/room-session-store';

const STATUS_LABELS = {
  connected: '已连接',
  connecting: '连接中',
  disconnected: '未连接',
  error: '连接异常',
} as const;

interface SessionShellProps extends PropsWithChildren {
  readonly eyebrow: string;
  readonly title: string;
  readonly description?: string;
  readonly actions?: ReactNode;
}

export function SessionShell({ eyebrow, title, description, actions, children }: SessionShellProps) {
  const status = useRoomSession((state) => state.status);
  const room = useRoomSession((state) => state.experience.roomState);
  const errorMessage = useRoomSession((state) => state.errorMessage);
  const clearError = useRoomSession((state) => state.clearError);
  const connect = useRoomSession((state) => state.connect);
  const canReconnect = status === 'disconnected' || status === 'error';

  return (
    <ScrollView className='appPage' scrollY>
      <View className='appMasthead'>
        <View className='mastheadCopy'>
          <Text className='mastheadBrand'>血染钟楼</Text>
          <Text className='mastheadEyebrow'>{eyebrow}</Text>
          <Text className='mastheadTitle'>{title}</Text>
          {description && <Text className='mastheadDescription'>{description}</Text>}
        </View>
        <View className='connectionControls'>
          <View className={`connectionBadge connection-${status}`}>
            <View className='connectionDot' />
            <Text>{STATUS_LABELS[status]}</Text>
          </View>
          {canReconnect && <Button className='reconnectButton' onClick={connect}>重新连接</Button>}
        </View>
      </View>

      {room && (
        <View className='roomStrip'>
          <Text className='roomStripLabel'>房间</Text>
          <Text className='roomStripCode'>{room.roomId}</Text>
          <Text className='roomStripMeta'>{scriptDisplayName(room.scriptId, room.scriptName)}</Text>
        </View>
      )}

      {errorMessage && (
        <View className='feedback feedbackError'>
          <Text className='feedbackText'>{errorMessage}</Text>
          <Button className='iconButton' aria-label='关闭错误提示' onClick={clearError}><Text className='iconGlyph'>×</Text></Button>
        </View>
      )}

      <View className='appContent'>{children}</View>
      {actions && <View className='stickyActions'>{actions}</View>}
    </ScrollView>
  );
}

export function LoadingState({ label = '正在恢复房间...' }: { readonly label?: string }) {
  return (
    <View className='emptyBand'>
      <View className='loadingMark' />
      <Text className='emptyBandTitle'>{label}</Text>
    </View>
  );
}

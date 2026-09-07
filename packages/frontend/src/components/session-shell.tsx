import { Button, ScrollView, Text, View } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { useEffect, useState, type PropsWithChildren, type ReactNode } from 'react';
import { scriptDisplayName } from '../lib/character-display';
import { useRoomSession } from '../lib/room-session-store';
import { RoomRecoveryPanel } from './room-recovery-panel';
import { RecoveryApprovalPanel } from './recovery-approval-panel';
import { GameHistoryPanel } from './game-history-panel';
import { RoomNavigation, type ShellDestination } from './room-navigation';
import { routeForExperience, SESSION_ROUTES } from '../lib/session-routing';
import { returnToTable } from '../lib/table-navigation';

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
  readonly hero?: ReactNode;
  readonly utilityContent?: ReactNode;
  readonly section?: 'library';
}

export function SessionShell({ eyebrow, title, description, actions, hero, utilityContent, section, children }: SessionShellProps) {
  const status = useRoomSession((state) => state.status);
  const room = useRoomSession((state) => state.experience.roomState);
  const experience = useRoomSession(state => state.experience);
  const identity = useRoomSession(state => state.identity);
  const playerName = useRoomSession(state => state.playerName);
  const approvalCount = useRoomSession(state => state.recoveryRequests.length);
  const errorMessage = useRoomSession((state) => state.errorMessage);
  const clearError = useRoomSession((state) => state.clearError);
  const connect = useRoomSession((state) => state.connect);
  const playerId = useRoomSession((state) => state.playerId);
  const pendingCommand = useRoomSession((state) => state.pendingCommand);
  const undoGame = useRoomSession((state) => state.undoGame);
  const redoGame = useRoomSession((state) => state.redoGame);
  const canReconnect = status === 'disconnected' || status === 'error';
  const [panel, setPanel] = useState<ShellDestination>('main');
  useEffect(() => { setPanel('main'); }, [room?.roomId, room?.phase]);
  function select(destination: ShellDestination): void {
    if (destination === 'library') {
      if (section === 'library') setPanel('main');
      else void Taro.navigateTo({ url: '/pages/scripts/index' });
      return;
    }
    if (destination === 'main' && section === 'library') {
      void returnToTable(routeForExperience(experience, Boolean(identity), true) ?? SESSION_ROUTES.lobby);
      return;
    }
    setPanel(destination);
  }
  const active = panel === 'main' && section === 'library' ? 'library' : panel;
  const heading = panel === 'history' ? '这一局的故事' : panel === 'tools' ? '桌边工具' : title;
  const caption = panel === 'history' ? 'GAME JOURNAL' : panel === 'tools' ? 'TABLE COMPANION' : eyebrow;

  return (
    <View className={`appShell ${actions && panel === 'main' ? 'appShellHasActions' : ''}`}>
      <View className='appTopBar'>
        <View className='brandLockup'>
          <View className='brandEmblem'><Text>Ⅻ</Text></View>
          <View><Text className='brandName'>血染钟楼</Text><Text className='brandCaption'>THE CLOCKTOWER · 面杀助手</Text></View>
        </View>
        <View className='connectionControls'>
          <View className={`connectionBadge connection-${status}`}>
            <View className='connectionDot' />
            <Text>{STATUS_LABELS[status]}</Text>
          </View>
          {canReconnect && <Button className='reconnectButton' onClick={connect}>重连</Button>}
        </View>
      </View>
      <ScrollView className='appPage shellScroll' scrollY key={`${panel}-${room?.phase ?? 'lobby'}`}>
      {panel === 'main' && hero ? hero : <View className='pageHero'>
        <Text className='mastheadEyebrow'>{caption}</Text><Text className='mastheadTitle'>{heading}</Text>
        {panel === 'main' && description && <Text className='mastheadDescription'>{description}</Text>}
        <View className='heroRule'><Text>◆</Text></View>
      </View>}
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

      <View className='appContent'>
        {panel === 'main' && children}
        {panel === 'history' && (room ? <>
          <View className='sectionBand'>
            <Text className='sectionHeading'>公开事件簿</Text><Text className='sectionDescription'>只记录已公开事件，不展示隐藏身份或夜间信息。</Text>
            {(room.nominationResults ?? []).map((entry, index) => <View className='timelineEntry' key={`vote-${index}`}>
              <Text className='timelineDay'>第 {entry.dayNumber} 天</Text>
              <Text className='playerName'>{room.players.find(player => player.id === entry.nomineeId)?.name ?? '玩家'} · 获得 {entry.yesVotes} 票</Text>
              <Text className='playerMeta'>提名人：{room.players.find(player => player.id === entry.nominatorId)?.name ?? '玩家'} · 门槛 {entry.requiredVotes} 票</Text>
            </View>)}
            {(room.deaths ?? []).map((entry, index) => <View className='timelineEntry' key={`death-${index}`}><Text className='timelineDay'>第 {entry.dayNumber} 天</Text><Text className='playerName'>{room.players.find(player => player.id === entry.playerId)?.name ?? '玩家'} 已死亡</Text></View>)}
            {!room.nominationResults?.length && !room.deaths?.length && <View className='emptyBand'><Text className='emptyBandTitle'>故事尚未开始</Text><Text className='emptyBandText'>提名、投票和公开死亡将在这里留下记录。</Text></View>}
          </View>
          <GameHistoryPanel room={room} actorId={playerId} busy={pendingCommand !== null} onUndo={undoGame} onRedo={redoGame} />
        </> : <View className='emptyBand'><Text className='emptyBandTitle'>还没有进行中的故事</Text><Text className='emptyBandText'>加入房间后，在这里查看本局的公开记录。</Text><Button className='secondaryButton' onClick={() => select('main')}>返回大厅</Button></View>)}
        {panel === 'tools' && <>
          <View className='sectionBand profileCard'><View className='profileAvatar'><Text>{(playerName || '客').slice(0, 1)}</Text></View><View><Text className='sectionHeading'>{playerName || '钟楼访客'}</Text><Text className='sectionDescription'>{room ? '本局身份保存在此设备' : '与朋友围坐一桌，开始新的故事'}</Text></View></View>
          {room && <RecoveryApprovalPanel />}
          <View className='sectionBand'><Text className='sectionHeading'>设备与恢复</Text><Text className='sectionDescription'>换设备时使用私人恢复码，确认后找回原座位。</Text><RoomRecoveryPanel /></View>
          <View className='sectionBand'><Text className='sectionHeading'>安心面杀</Text><Text className='sectionDescription'>私密信息按住查看，松手或切后台隐藏。请勿录屏或向其他玩家展示私人信息。</Text><View className='choiceRow'><Text className='tag'>无震动</Text><Text className='tag'>无自动音效</Text><Text className='tag'>真人说书人</Text></View></View>
          {utilityContent}
        </>}
      </View>
      </ScrollView>
      {actions && panel === 'main' && <View className='stickyActions'>{actions}</View>}
      <RoomNavigation active={active} hasRoom={Boolean(room)} approvalCount={approvalCount} onSelect={select} />
    </View>
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

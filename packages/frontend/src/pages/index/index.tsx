import { Button, Input, Text, View } from '@tarojs/components';
import Taro, { useRouter, useShareAppMessage } from '@tarojs/taro';
import { useEffect, useState } from 'react';
import { SessionShell } from '../../components/session-shell';
import {
  isDevelopmentBuild,
  useRoomSession,
} from '../../lib/room-session-store';
import { SESSION_ROUTES } from '../../lib/session-routing';
import { useSessionRoute } from '../../lib/use-session-route';
import { eventValue, type InputEvent } from '../../lib/utils';
import './index.css';

export default function LobbyPage() {
  const router = useRouter<{ roomId?: string }>();
  const initialized = useRoomSession((state) => state.initialized);
  const status = useRoomSession((state) => state.status);
  const identity = useRoomSession((state) => state.identity);
  const identityStatus = useRoomSession((state) => state.experience.identityStatus);
  const playerName = useRoomSession((state) => state.playerName);
  const endpoint = useRoomSession((state) => state.endpoint);
  const maxPlayers = useRoomSession((state) => state.maxPlayers);
  const pendingCommand = useRoomSession((state) => state.pendingCommand);
  const setInviteRoomId = useRoomSession((state) => state.setInviteRoomId);
  const setPlayerName = useRoomSession((state) => state.setPlayerName);
  const setEndpoint = useRoomSession((state) => state.setEndpoint);
  const setMaxPlayers = useRoomSession((state) => state.setMaxPlayers);
  const connect = useRoomSession((state) => state.connect);
  const createRoom = useRoomSession((state) => state.createRoom);
  const joinRoom = useRoomSession((state) => state.joinRoom);
  const resumeLastRoom = useRoomSession((state) => state.resumeLastRoom);
  const rejoinRoom = useRoomSession((state) => state.rejoinRoom);
  const forgetRoomIdentity = useRoomSession((state) => state.forgetRoomIdentity);
  const [roomId, setRoomId] = useState('');
  const [endpointDraft, setEndpointDraft] = useState(endpoint);
  const [developerToolsOpen, setDeveloperToolsOpen] = useState(false);

  useSessionRoute(SESSION_ROUTES.lobby);

  useEffect(() => {
    const invitedRoomId = router.params.roomId?.trim() ?? '';
    if (!invitedRoomId) return;
    setRoomId(invitedRoomId);
    setInviteRoomId(invitedRoomId);
  }, [router.params.roomId, setInviteRoomId]);

  useEffect(() => {
    setEndpointDraft(endpoint);
  }, [endpoint]);

  useShareAppMessage(() => ({
    title: roomId ? `加入血染钟楼房间 ${roomId}` : '血染钟楼',
    path: roomId ? `${SESSION_ROUTES.lobby}?roomId=${encodeURIComponent(roomId)}` : SESSION_ROUTES.lobby,
  }));

  const busy = pendingCommand !== null;
  const hasDifferentInvitation = Boolean(identity && roomId && identity.roomId !== roomId);
  const retainedIdentity = identityStatus?.status === 'retained';
  const canRejoin = retainedIdentity && identityStatus.canRejoin;
  const rejoinUnavailable = retainedIdentity && !identityStatus.canRejoin;

  function confirmForgetIdentity(): void {
    void Taro.showModal({
      title: '放弃房间身份？',
      content: '此设备将不能再恢复该房间，除非重新获得可加入的邀请。',
      confirmText: '确认放弃',
      confirmColor: '#7b2028',
    }).then((result) => {
      if (result.confirm) forgetRoomIdentity();
    });
  }

  function openScripts(): void {
    void Taro.navigateTo({ url: '/pages/scripts/index' });
  }

  if (!initialized) {
    return (
      <SessionShell eyebrow='大厅' title='正在准备房间'>
        <View className='emptyBand'><Text className='emptyBandTitle'>正在载入身份...</Text></View>
      </SessionShell>
    );
  }

  return (
    <SessionShell
      eyebrow='大厅'
      title='创建或加入一局'
      description='邀请制房间，不公开展示正在进行的游戏。'
    >
      {identity && (
        <View className='sectionBand identityBand'>
          <Text className='sectionHeading'>可恢复的房间</Text>
          <Text className='identityRoomCode'>{identity.roomId}</Text>
          <Text className='sectionDescription'>
            {identityStatus?.status === 'retained'
              ? identityStatus.canRejoin
                ? '你已暂时离开座位，可以在身份锁定前重新加入。'
                : '成员名单已经锁定，此身份不能再加入该房间。'
              : '正在使用此设备保存的凭证恢复权威状态。'}
          </Text>
          {hasDifferentInvitation && (
            <View className='feedback feedbackError inlineFeedback'>
              <Text className='feedbackText'>邀请房间 {roomId} 与当前身份不一致，请先恢复并按房间规则结束旧房间。</Text>
            </View>
          )}
          <View className='commandRow'>
            {!rejoinUnavailable && (
              <Button
                className='secondaryButton'
                disabled={busy || status !== 'connected'}
                onClick={canRejoin ? rejoinRoom : resumeLastRoom}
              >
                {canRejoin ? '重新加入' : '立即恢复'}
              </Button>
            )}
            {rejoinUnavailable && (
              <Button className='dangerButton' disabled={busy} onClick={confirmForgetIdentity}>清理失效身份</Button>
            )}
          </View>
        </View>
      )}

      {!identity && (
        <View className='lobbyColumns'>
          <View className='sectionBand lobbyColumn'>
            <Text className='sectionHeading'>创建房间</Text>
            <Text className='sectionDescription'>实际玩家 5-15 人，说书人不计入人数。</Text>
            <View className='fieldGroup'>
              <Text className='fieldLabel'>你的昵称</Text>
              <Input
                className='textInput'
                maxlength={24}
                value={playerName}
                placeholder='输入桌上使用的昵称'
                onInput={(event: InputEvent) => setPlayerName(eventValue(event))}
              />
            </View>
            <View className='fieldGroup'>
              <Text className='fieldLabel'>实际玩家上限</Text>
              <View className='stepper'>
                <Button className='iconButton stepperButton' disabled={maxPlayers <= 5} onClick={() => setMaxPlayers(maxPlayers - 1)}><Text className='iconGlyph'>−</Text></Button>
                <Text className='stepperValue'>{maxPlayers}</Text>
                <Button className='iconButton stepperButton' disabled={maxPlayers >= 15} onClick={() => setMaxPlayers(maxPlayers + 1)}><Text className='iconGlyph'>+</Text></Button>
              </View>
            </View>
            <Button
              className='commandButton fullWidthButton'
              disabled={busy || status !== 'connected' || !playerName.trim()}
              onClick={createRoom}
            >
              创建房间
            </Button>
          </View>

          <View className='sectionBand lobbyColumn'>
            <Text className='sectionHeading'>加入房间</Text>
            <Text className='sectionDescription'>邀请会自动填入房间号，但不会替你加入。</Text>
            <View className='fieldGroup'>
              <Text className='fieldLabel'>你的昵称</Text>
              <Input
                className='textInput'
                maxlength={24}
                value={playerName}
                placeholder='输入桌上使用的昵称'
                onInput={(event: InputEvent) => setPlayerName(eventValue(event))}
              />
            </View>
            <View className='fieldGroup'>
              <Text className='fieldLabel'>房间号</Text>
              <Input
                className='textInput roomCodeInput'
                maxlength={32}
                value={roomId}
                placeholder='输入邀请中的房间号'
                onInput={(event: InputEvent) => setRoomId(eventValue(event).trim())}
              />
            </View>
            <Button
              className='secondaryButton fullWidthButton'
              disabled={busy || status !== 'connected' || !playerName.trim() || !roomId.trim()}
              onClick={() => joinRoom(roomId)}
            >
              加入房间
            </Button>
          </View>
        </View>
      )}

      <View className='sectionBand utilityBand'>
        <Text className='sectionHeading'>剧本</Text>
        <Text className='sectionDescription'>当前可用剧本：暗流涌动。</Text>
        <Button className='quietButton utilityButton' onClick={openScripts}>
          <Text className='buttonLabelDark'>查看角色与唤醒顺序</Text>
        </Button>
      </View>

      {status !== 'connected' && (
        <View className='sectionBand'>
          <Text className='sectionHeading'>连接未就绪</Text>
          <Button className='secondaryButton utilityButton' disabled={status === 'connecting'} onClick={connect}>重新连接</Button>
        </View>
      )}

      {isDevelopmentBuild && (
        <View className='sectionBand developerBand'>
          <Button className='quietButton utilityButton' onClick={() => setDeveloperToolsOpen((open) => !open)}>
            <Text className='buttonLabelDark'>{developerToolsOpen ? '收起开发工具' : '开发工具'}</Text>
          </Button>
          {developerToolsOpen && (
            <View className='developerTools'>
              <Text className='sectionDescription'>仅开发构建显示。</Text>
              <Input
                className='textInput'
                value={endpointDraft}
                placeholder='ws://localhost:8080/ws'
                onInput={(event: InputEvent) => setEndpointDraft(eventValue(event))}
              />
              <View className='commandRow'>
                <Button className='quietButton' onClick={() => setEndpoint(endpointDraft.trim())}>
                  <Text className='buttonLabelDark'>应用地址</Text>
                </Button>
                {identity && <Button className='dangerButton' disabled={busy} onClick={confirmForgetIdentity}>强制清理本地身份</Button>}
              </View>
            </View>
          )}
        </View>
      )}
    </SessionShell>
  );
}

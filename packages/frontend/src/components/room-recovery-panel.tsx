import { Button, Input, Text, View } from '@tarojs/components';
import Taro, { useDidHide } from '@tarojs/taro';
import { useEffect, useState } from 'react';
import { createPrivateCharacterVisibility, revealPrivateCharacter, concealPrivateCharacter } from './private-character-visibility';
import { encodeRecoveryCode, parseRecoveryCode } from '../lib/room-recovery';
import { useRoomSession } from '../lib/room-session-store';
import { eventValue, type InputEvent } from '../lib/utils';

export function RoomRecoveryPanel() {
  const identity = useRoomSession((state) => state.identity);
  const importRoomIdentity = useRoomSession((state) => state.importRoomIdentity);
  const pending = useRoomSession((state) => state.pendingCommand);
  const recoveryStatus = useRoomSession((state) => state.recoveryStatus);
  const [open, setOpen] = useState(false);
  const [code, setCode] = useState('');
  const [feedback, setFeedback] = useState('');
  const [visibility, setVisibility] = useState(createPrivateCharacterVisibility);

  function close(): void { setOpen(false); setCode(''); setFeedback(''); setVisibility(concealPrivateCharacter); }
  useDidHide(close);
  useEffect(() => {
    if (visibility.concealAt === null) return;
    const timer = setTimeout(() => setVisibility(concealPrivateCharacter), Math.max(0, visibility.concealAt - Date.now()));
    return () => clearTimeout(timer);
  }, [visibility.concealAt]);

  async function showCode(): Promise<void> {
    if (!identity?.recoveryCredential) { setFeedback('请先连接服务器刷新身份，再生成新版恢复码。'); return; }
    const result = await Taro.showModal({ title: '显示私人恢复码？', content: '恢复码不要发群，也不要展示给其他玩家。仅在自己的新设备输入。30 秒后自动隐藏。', confirmText: '确认显示' });
    if (result.confirm) setVisibility(previous => revealPrivateCharacter(previous, Date.now()));
  }

  function restore(): void {
    const result = parseRecoveryCode(code);
    if (!result.ok) { setFeedback(result.error); return; }
    importRoomIdentity(result.identity);
    close();
  }

  async function exportCode(): Promise<void> {
    if (!identity?.recoveryCredential) { setFeedback('请先连接服务器刷新身份，再生成新版恢复码。'); return; }
    const result = await Taro.showModal({
      title: '复制私人恢复码',
      content: '恢复码用于向说书人申请找回座位。仅保存在自己的安全位置，不要发送到游戏群。恢复码不是加密文件。',
      confirmText: '确认复制',
    });
    if (!result.confirm) return;
    try {
      await Taro.setClipboardData({ data: encodeRecoveryCode({ roomId: identity.roomId, playerId: identity.playerId, recoveryCredential: identity.recoveryCredential }) });
      setFeedback('恢复码已复制，请妥善保管。');
    } catch { setFeedback('复制失败，请检查设备剪贴板权限。'); }
  }

  return (
    <View className='recoveryEntry'>
      {recoveryStatus === 'pending' && <Text className='sectionDescription'>恢复申请已提交，等待说书人或房主审批。批准前不会显示身份。请保持此页面打开。</Text>}
      {recoveryStatus === 'rejected' && <Text className='sectionDescription'>申请未获批准，请与桌上的说书人核实后重试。</Text>}
      {recoveryStatus === 'expired' && <Text className='sectionDescription'>恢复申请已过期，请重新提交。</Text>}
      <Button className='quietButton' onClick={() => setOpen(true)}>{identity ? '保存私人恢复码' : '使用恢复码找回座位'}</Button>
      {open && (
        <View className='recoveryBackdrop'>
          <View className='recoveryDialog'>
            <Text className='sectionHeading'>{identity ? '私人房间身份' : '找回原来的座位'}</Text>
            <Text className='sectionDescription'>恢复码只用于恢复本人身份，请勿公开或分享给其他玩家。</Text>
            {identity ? (
              <>
                <Text className='recoveryMask'>•••• •••• •••• ••••</Text>
                {visibility.revealed && identity.recoveryCredential && <Text selectable className='recoveryVisibleCode'>{encodeRecoveryCode({ roomId: identity.roomId, playerId: identity.playerId, recoveryCredential: identity.recoveryCredential })}</Text>}
                <Text className='sectionDescription'>房间：{identity.roomId}</Text>
                <Button className='commandButton recoveryAction' onClick={() => { void exportCode(); }}>确认并复制恢复码</Button>
                <Button className='quietButton recoveryAction' onClick={() => { void showCode(); }}>显示恢复码（30秒）</Button>
              </>
            ) : (
              <>
                <Input className='textInput recoveryAction' password maxlength={8192} value={code} placeholder='粘贴原设备保存的恢复码' onInput={(event: InputEvent) => setCode(eventValue(event))} />
                <Text className='sectionDescription'>需说书人审批；恢复说书人本人时由房主审批。若你是唯一审批人，请在原设备先指定另一说书人或转移房主。旧版 CT2 恢复码不再支持，请重新导出。</Text>
                <Button className='commandButton recoveryAction' disabled={!code.trim() || pending !== null || recoveryStatus === 'pending'} onClick={restore}>申请恢复身份</Button>
              </>
            )}
            {feedback && <Text className='recoveryFeedback'>{feedback}</Text>}
            <Button className='quietButton recoveryAction' onClick={close}>关闭</Button>
          </View>
        </View>
      )}
    </View>
  );
}

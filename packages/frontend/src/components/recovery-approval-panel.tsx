import { Button, Text, View } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { useRoomSession } from '../lib/room-session-store';

export function RecoveryApprovalPanel() {
  const identity = useRoomSession(state => state.identity);
  const room = useRoomSession(state => state.experience.roomState);
  const requests = useRoomSession(state => state.recoveryRequests);
  const review = useRoomSession(state => state.reviewRecovery);
  const refresh = useRoomSession(state => state.getRecoveryRequests);
  const pending = useRoomSession(state => state.pendingCommand);
  const eligible = identity && room && (identity.playerId === room.storytellerId || identity.playerId === room.creatorId);
  if (!eligible) return null;

  async function decide(requestId: string, playerName: string, decision: boolean): Promise<void> {
    if (decision) {
      const result = await Taro.showModal({ title: '批准设备恢复？', content: `请当面核实 ${playerName} 正在更换设备。批准后原设备将退出，恢复码也会更新。`, confirmText: '确认批准' });
      if (!result.confirm) return;
    }
    review(requestId, decision);
  }

  return <View className='sectionBand'>
    <Text className='sectionHeading'>设备恢复审批</Text>
    <Text className='sectionDescription'>请当面核实申请人。只显示你有权审批的请求。</Text>
    {requests.length === 0 && <Text className='sectionDescription'>暂无待审批请求。</Text>}
    {requests.map(request => <View className='playerRow' key={request.requestId}>
      <View className='playerBody'><Text className='playerName'>{request.playerName}</Text><Text className='playerMeta'>请求更换设备并恢复座位</Text></View>
      <Button className='secondaryButton' disabled={pending !== null} onClick={() => { void decide(request.requestId, request.playerName, true); }}>同意</Button>
      <Button className='dangerButton' disabled={pending !== null} onClick={() => { void decide(request.requestId, request.playerName, false); }}>拒绝</Button>
    </View>)}
    <Button className='quietButton utilityButton' onClick={refresh}>刷新申请</Button>
  </View>;
}

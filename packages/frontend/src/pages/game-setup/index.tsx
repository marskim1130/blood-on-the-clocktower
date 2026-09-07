import { Button, Picker, Text, View } from '@tarojs/components';
import Taro, { useShareAppMessage } from '@tarojs/taro';
import { useEffect, useMemo, useState } from 'react';
import {
  TROUBLE_BREWING_SCRIPT,
  randomizeScriptAssignments,
  swapScriptAssignments,
  validateScriptSetup,
  type ScriptSetup,
} from '@clocktower/core';
import { LoadingState, SessionShell } from '../../components/session-shell';
import { PrivateCharacterCard } from '../../components/private-character-card';
import { characterDisplayName } from '../../lib/character-display';
import { createRoomInvite, getBrowserOrigin } from '../../lib/room-invite';
import { useRoomSession } from '../../lib/room-session-store';
import { SESSION_ROUTES } from '../../lib/session-routing';
import { useSessionRoute } from '../../lib/use-session-route';
import {
  canAssignCharacters,
  membershipKey,
  normalizeDemonBluffs,
  summarizeCharacterConfirmation,
  summarizeReadiness,
} from './utils';
import { SeatOrderEditor } from './seat-order-editor';
import './index.css';

interface PickerChangeEvent {
  readonly detail: { readonly value: string | number };
}

const CHARACTER_BY_ID = new Map(
  TROUBLE_BREWING_SCRIPT.characters.map((character) => [character.id, character]),
);

function setupErrorMessage(code: string): string {
  const labels: Readonly<Record<string, string>> = {
    UNSUPPORTED_PLAYER_COUNT: '需要 5-15 名实际玩家才能配置角色。',
    ASSIGNMENT_COUNT_MISMATCH: '每名玩家都必须且只能获得一个角色。',
    UNKNOWN_CHARACTER: '配置中包含当前剧本不存在的角色。',
    DUPLICATE_CHARACTER: '同一个角色不能重复在场。',
    ROLE_COUNT_MISMATCH: '镇民、外来者、爪牙与恶魔数量不合法；男爵会增加两名外来者。',
    PLAYER_ASSIGNMENT_MISMATCH: '角色映射与当前玩家名单不一致。',
    INVALID_DRUNK_SHOWN_CHARACTER: '酒鬼必须看到一个未实际在场的镇民身份。',
    INVALID_FORTUNE_TELLER_RED_HERRING: '占卜师干扰项必须是另一名善良玩家。',
    INVALID_DEMON_BLUFFS: '必须选择三张不重复、未在场且未作为酒鬼展示的善良身份。',
  };
  return labels[code] ?? '当前角色配置不合法。';
}

export default function GameSetupPage() {
  const room = useRoomSession((state) => state.experience.roomState);
  const identityStatus = useRoomSession((state) => state.experience.identityStatus);
  const playerId = useRoomSession((state) => state.playerId);
  const pendingCommand = useRoomSession((state) => state.pendingCommand);
  const setStoryteller = useRoomSession((state) => state.setStoryteller);
  const transferOwnership = useRoomSession((state) => state.transferOwnership);
  const setSeatOrder = useRoomSession((state) => state.setSeatOrder);
  const setReady = useRoomSession((state) => state.setReady);
  const confirmCharacter = useRoomSession((state) => state.confirmCharacter);
  const assignCharacters = useRoomSession((state) => state.assignCharacters);
  const startGame = useRoomSession((state) => state.startGame);
  const kickPlayer = useRoomSession((state) => state.kickPlayer);
  const leaveRoom = useRoomSession((state) => state.leaveRoom);
  const closeRoom = useRoomSession((state) => state.closeRoom);
  const [draft, setDraft] = useState<ScriptSetup | null>(null);
  const [localError, setLocalError] = useState('');
  const [activePanel, setActivePanel] = useState<'seats' | 'roles' | 'manage'>('seats');

  useSessionRoute(SESSION_ROUTES.setup);

  const playerKey = membershipKey(room?.players ?? []);
  const isCreator = room?.creatorId === playerId;
  const isStoryteller = room?.storytellerId === playerId;
  const frozen = identityStatus?.participantSetFrozen === true;
  const visibleSelf = room?.players.find((player) => player.id === playerId) ?? null;
  const projectedStorytellerName = room
    ? (room as typeof room & { readonly storytellerName?: string }).storytellerName
    : undefined;

  useEffect(() => {
    if (!room || !isStoryteller || frozen || room.players.length < 5 || room.players.length > 15) return;
    try {
      setDraft(randomizeScriptAssignments(room.players));
      setLocalError('');
    } catch (error) {
      setDraft(null);
      setLocalError(error instanceof Error ? error.message : '无法生成角色配置');
    }
  }, [frozen, isStoryteller, playerKey]);

  useShareAppMessage(() => ({
    title: room ? `加入血染钟楼房间 ${room.roomId}` : '血染钟楼',
    path: room ? `${SESSION_ROUTES.lobby}?roomId=${encodeURIComponent(room.roomId)}` : SESSION_ROUTES.lobby,
  }));

  const validation = useMemo(
    () => room && draft ? validateScriptSetup(draft, room.players) : null,
    [draft, room],
  );
  const readiness = summarizeReadiness(room?.players ?? [], playerId);
  const characterConfirmation = summarizeCharacterConfirmation(room?.players ?? [], playerId);
  const assignmentAllowed = canAssignCharacters(validation?.ok === true, readiness.allReady);

  const actualCharacters = useMemo(
    () => new Set(Object.values(draft?.assignments ?? {})),
    [draft],
  );

  const drunkPlayerId = useMemo(
    () => Object.entries(draft?.assignments ?? {}).find(([, characterId]) => characterId === 'drunk')?.[0] ?? null,
    [draft],
  );

  const fortuneTellerPlayerId = useMemo(
    () => Object.entries(draft?.assignments ?? {}).find(([, characterId]) => characterId === 'fortuneteller')?.[0] ?? null,
    [draft],
  );

  const unusedTownsfolk = TROUBLE_BREWING_SCRIPT.characters.filter(
    (character) => character.type === 'townsfolk' && !actualCharacters.has(character.id),
  );

  const eligibleDemonBluffs = TROUBLE_BREWING_SCRIPT.characters.filter(
    (character) => character.team === 'good' &&
      !actualCharacters.has(character.id) &&
      !Object.values(draft?.shownCharacters ?? {}).includes(character.id),
  );

  const redHerringCandidates = (room?.players ?? []).filter((player) => {
    const characterId = draft?.assignments[player.id];
    return player.id !== fortuneTellerPlayerId && CHARACTER_BY_ID.get(characterId ?? '')?.team === 'good';
  });
  const projectedRedHerringIndex = room?.players.findIndex(
    (player) => player.id === room?.fortuneTellerRedHerringId,
  ) ?? -1;
  const projectedRedHerring = projectedRedHerringIndex >= 0
    ? room?.players[projectedRedHerringIndex]
    : undefined;

  const busy = pendingCommand !== null;

  function randomize(): void {
    if (!room) return;
    try {
      setDraft(randomizeScriptAssignments(room.players));
      setLocalError('');
    } catch (error) {
      setLocalError(error instanceof Error ? error.message : '无法生成角色配置');
    }
  }

  function chooseCharacter(targetPlayerId: string, characterId: string): void {
    if (!draft) return;
    const currentOwner = Object.entries(draft.assignments).find(([, assigned]) => assigned === characterId)?.[0];
    const assignments = currentOwner && currentOwner !== targetPlayerId
      ? swapScriptAssignments(draft.assignments, targetPlayerId, currentOwner)
      : { ...draft.assignments, [targetPlayerId]: characterId };
    const shownCharacters = Object.fromEntries(
      Object.entries(draft.shownCharacters).filter(([id]) => assignments[id] === 'drunk'),
    );
    const eligibleIds = TROUBLE_BREWING_SCRIPT.characters
      .filter((character) => character.team === 'good' &&
        !Object.values(assignments).includes(character.id) &&
        !Object.values(shownCharacters).includes(character.id))
      .map((character) => character.id);
    setDraft({
      ...draft,
      assignments,
      shownCharacters,
      demonBluffCharacterIds: normalizeDemonBluffs(eligibleIds, draft.demonBluffCharacterIds),
    });
  }

  function chooseDrunkShownCharacter(characterId: string): void {
    if (!draft || !drunkPlayerId) return;
    const shownCharacters = { [drunkPlayerId]: characterId };
    const eligibleIds = TROUBLE_BREWING_SCRIPT.characters
      .filter((character) => character.team === 'good' &&
        !actualCharacters.has(character.id) &&
        !Object.values(shownCharacters).includes(character.id))
      .map((character) => character.id);
    setDraft({
      ...draft,
      shownCharacters,
      demonBluffCharacterIds: normalizeDemonBluffs(eligibleIds, draft.demonBluffCharacterIds),
    });
  }

  function chooseRedHerring(targetPlayerId: string): void {
    if (!draft) return;
    setDraft({ ...draft, fortuneTellerRedHerringId: targetPlayerId || null });
  }

  function toggleDemonBluff(characterId: string): void {
    if (!draft) return;
    const current = [...draft.demonBluffCharacterIds];
    const selected = current.includes(characterId);
    const demonBluffCharacterIds = selected
      ? current.filter((id) => id !== characterId)
      : [...(current.length >= 3 ? current.slice(1) : current), characterId];
    setDraft({ ...draft, demonBluffCharacterIds });
  }

  function confirmAssignment(): void {
    if (!draft || !assignmentAllowed) return;
    void Taro.showModal({
      title: '确认发放身份？',
      content: '身份发放后将锁定成员名单，不能重新洗牌或修改角色。',
      confirmText: '确认发放',
      confirmColor: '#7b2028',
    }).then((result) => {
      if (!result.confirm) return;
      assignCharacters(
        { ...draft.assignments },
        { ...draft.shownCharacters },
        draft.fortuneTellerRedHerringId ?? undefined,
        [...draft.demonBluffCharacterIds],
      );
    });
  }

  function copyInvite(): void {
    if (!room) return;
    const path = createRoomInvite(room.roomId, getBrowserOrigin());
    void Taro.setClipboardData({ data: `血染钟楼房间：${room.roomId}\n${path}` })
      .then(() => Taro.showToast({ title: '邀请已复制', icon: 'success' }));
  }

  function confirmExit(): void {
    const creatorAction = isCreator;
    void Taro.showModal({
      title: creatorAction ? '关闭房间？' : '离开房间？',
      content: creatorAction ? '房间关闭后所有成员都会退出。' : '身份锁定前可以使用原凭证重新加入。',
      confirmText: creatorAction ? '关闭房间' : '离开',
      confirmColor: '#7b2028',
    }).then((result) => {
      if (!result.confirm) return;
      if (creatorAction) closeRoom(); else leaveRoom();
    });
  }

  const ownershipCandidates = [
    ...(room?.players ?? []).map((player) => ({ id: player.id, name: player.name })),
    ...(room?.storytellerId && !room.players.some((player) => player.id === room.storytellerId)
      ? [{ id: room.storytellerId, name: projectedStorytellerName || '说书人' }] : []),
  ].filter((member) => member.id !== playerId);

  function confirmOwnershipTransfer(event: PickerChangeEvent): void {
    const target = ownershipCandidates[Number(event.detail.value)];
    if (!target || busy || !isCreator) return;
    void Taro.showModal({
      title: '转让房主？',
      content: `将房间管理权交给 ${target.name}。你的说书人或玩家身份保持不变，此后由对方管理和关闭房间。`,
      confirmText: '确认转让',
    }).then((result) => { if (result.confirm) transferOwnership(target.id); });
  }

  if (!room) {
    return <SessionShell eyebrow='游戏设置' title='恢复准备状态'><LoadingState /></SessionShell>;
  }

  const storytellerLabel = projectedStorytellerName || (room.storytellerId ? '已设置' : '未设置');
  const actualPlayerCount = Math.max(0, room.players.length - (room.storytellerId ? 0 : 1));
  const panel = activePanel === 'roles' && !isStoryteller || activePanel === 'manage' && !isCreator ? 'seats' : activePanel;
  const nextStep = !room.storytellerId ? '先指定说书人，再确认座位'
    : frozen ? characterConfirmation.allConfirmed ? '全员已确认，可以开始首夜' : '请查看并确认私密身份'
    : room.players.length < 5 ? '邀请朋友入座，至少需要5位玩家'
    : readiness.allReady ? '全员已准备，等待说书人发放身份' : '核对座位后点击准备';

  return (
    <SessionShell
      eyebrow='游戏设置'
      title={frozen ? '身份已发放' : '组建这一局'}
      description={frozen ? '等待说书人开始首夜。' : '确认成员、说书人和角色配置。'}
      actions={
        frozen && isStoryteller
          ? <Button className='commandButton' disabled={busy || !characterConfirmation.allConfirmed} onClick={startGame}>
              {characterConfirmation.allConfirmed ? '开始游戏' : `等待确认 ${characterConfirmation.confirmedCount}/${characterConfirmation.totalCount}`}
            </Button>
          : <View className='commandRow'>
              {!frozen && isStoryteller && <Button className='commandButton' disabled={busy || !draft} onClick={() => setActivePanel('roles')}>配置角色</Button>}
              {!frozen && room.storytellerId && !isStoryteller && readiness.selfReady !== null && <Button
                className={readiness.selfReady ? 'secondaryButton' : 'commandButton'} disabled={busy} onClick={() => setReady(!readiness.selfReady)}
              >{pendingCommand === 'ready' ? '等待服务器确认…' : readiness.selfReady ? '取消准备' : '确认座位并准备'}</Button>}
              <Button className='quietButton' disabled={busy} onClick={copyInvite}><Text className='buttonLabelDark'>复制邀请</Text></Button>
            </View>
      }
    >
      <View className='sectionBand setupProgress'>
        <View className='setupProgressHeading'><Text className='sectionHeading'>{nextStep}</Text><Text className='tag'>{frozen ? '03 · 身份确认' : room.storytellerId ? '02 · 入座准备' : '01 · 组建游戏'}</Text></View>
        <View className='setupSummary'>
          <View className='setupMetric'>
            <Text className='setupMetricLabel'>{room.storytellerId ? '实际玩家' : '预计玩家'}</Text>
            <Text className='setupMetricValue'>{actualPlayerCount} / {room.maxPlayers}</Text>
          </View>
          <View className='setupMetric'>
            <Text className='setupMetricLabel'>说书人</Text>
            <Text className='setupMetricValue'>{storytellerLabel}</Text>
          </View>
          <View className='setupMetric'>
            <Text className='setupMetricLabel'>名单状态</Text>
            <Text className='setupMetricValue'>{frozen ? '已锁定' : '可调整'}</Text>
          </View>
          <View className='setupMetric'>
            <Text className='setupMetricLabel'>{frozen ? '身份确认' : '准备状态'}</Text>
            <Text className='setupMetricValue'>
              {frozen ? characterConfirmation.confirmedCount : readiness.readyCount} / {frozen ? characterConfirmation.totalCount : readiness.totalCount}
            </Text>
          </View>
        </View>
      </View>

      <View className='pageTabs'>
        <Button className={`pageTab ${panel === 'seats' ? 'pageTabActive' : ''}`} onClick={() => setActivePanel('seats')}>座位</Button>
        {isStoryteller && <Button className={`pageTab ${panel === 'roles' ? 'pageTabActive' : ''}`} onClick={() => setActivePanel('roles')}>角色配置</Button>}
        {isCreator && <Button className={`pageTab ${panel === 'manage' ? 'pageTabActive' : ''}`} onClick={() => setActivePanel('manage')}>房间设置</Button>}
      </View>
      <View className='surfaceGrid'>
      <View className='workspaceMain'>
      {panel === 'seats' && frozen && !isStoryteller && visibleSelf?.character && (
        <View className='sectionBand'>
          <Text className='sectionHeading'>你的身份</Text>
          <PrivateCharacterCard character={visibleSelf.character} roomId={room.roomId} playerName={visibleSelf.name || '未命名玩家'}
            confirmed={visibleSelf.hasConfirmedCharacter} confirmPending={pendingCommand === 'confirm-character'} onConfirm={confirmCharacter} />
        </View>
      )}
      {panel === 'seats' && <View className='sectionBand'>
        <Text className='sectionHeading'>成员与座位</Text>
        <Text className='sectionDescription'>按现实桌面顺时针入座 · {room.players.length} 个座位</Text>
        <View className='seatGrid'>
          {room.players.map((player, index) => (
            <View className={`seatCard ${player.id === playerId ? 'seatCardSelf' : ''}`} key={player.id}>
              <Text className='seatAvatar'>{String(index + 1).padStart(2, '0')}</Text>
              <View className='playerBody'>
                <Text className='playerName'>{player.name || `座位 ${index + 1}`}{player.id === playerId ? '（你）' : ''}</Text>
                <Text className='playerMeta'>
                  {frozen
                    ? isStoryteller ? characterDisplayName(player.character) ?? '未分配' : '身份已私下发放'
                    : '等待配置'}
                </Text>
              </View>
              {room.storytellerId && (
                <Text className={`readyBadge ${(frozen ? player.hasConfirmedCharacter : player.isReady) ? 'readyBadgeReady' : ''}`}>
                  {frozen
                    ? player.hasConfirmedCharacter ? '已确认' : '未确认'
                    : player.isReady ? '已准备' : '未准备'}
                </Text>
              )}
              {isCreator && !frozen && !room.storytellerId && (
                <Button className='quietButton' disabled={busy} onClick={() => {
                  if (!room.storytellerId) { setStoryteller(player.id); return; }
                  void Taro.showModal({ title: '更换说书人？', content: `${player.name} 将成为说书人，原说书人回到该座位，所有玩家需要重新准备。`, confirmText: '确认更换' })
                    .then((result) => { if (result.confirm) setStoryteller(player.id); });
                }}><Text className='buttonLabelDark'>{room.storytellerId ? '换为说书人' : '指定说书人'}</Text></Button>
              )}
            </View>
          ))}
        </View>
      </View>}

      {panel === 'manage' && isCreator && room.storytellerId && !frozen && (
        <View className='sectionBand'>
          <Text className='sectionHeading'>调整顺时针座次</Text>
          <SeatOrderEditor
            key={room.players.map((player) => player.id).join('|')}
            players={room.players}
            currentPlayerId={playerId}
            busy={busy}
            onSubmit={setSeatOrder}
          />
        </View>
      )}

      {!room.storytellerId && (
        <View className='emptyBand'>
          <Text className='emptyBandTitle'>{isCreator ? '请选择一名说书人' : '等待房主指定说书人'}</Text>
        </View>
      )}

      {panel === 'seats' && room.storytellerId && !frozen && readiness.selfReady !== null && (
        <View className='sectionBand readinessPanel'>
          <Text className='sectionHeading'>确认你的座位</Text>
          <Text className='sectionDescription'>请核对顺时针座次。座位或成员变化后，所有玩家都需要重新确认。</Text>
          <View className='readinessActions'>
            <Text className='readinessHint'>当前 {readiness.readyCount} / {readiness.totalCount} 人已准备</Text>
            <Text className='sectionDescription'>{readiness.selfReady ? '你已准备，等待其他玩家。' : '核对完成后，使用下方“确认座位并准备”。'}</Text>
          </View>
        </View>
      )}

      {panel === 'roles' && room.storytellerId && !frozen && isStoryteller && draft && (
        <View className='sectionBand'>
          <Text className='sectionHeading'>角色配置</Text>
          <Text className='sectionDescription'>可重新抽取整个角色池，或逐座位选择；选择已在场角色时会交换两人的身份。</Text>
          <Button className='secondaryButton utilityButton' disabled={busy} onClick={randomize}>重新随机</Button>
          <View className='playerList'>
            {room.players.map((player, index) => {
              const selectedId = draft.assignments[player.id];
              const selectedIndex = Math.max(0, TROUBLE_BREWING_SCRIPT.characters.findIndex((character) => character.id === selectedId));
              return (
                <View className='playerRow roleDraftRow' key={player.id}>
                  <Text className='seatNumber'>{String(index + 1).padStart(2, '0')}</Text>
                  <View className='playerBody'><Text className='playerName'>{player.name || `座位 ${index + 1}`}</Text></View>
                  <Picker
                    className='roleDraftControl'
                    mode='selector'
                    range={TROUBLE_BREWING_SCRIPT.characters.map((character) => character.name)}
                    value={selectedIndex}
                    onChange={(event: PickerChangeEvent) => {
                      const character = TROUBLE_BREWING_SCRIPT.characters[Number(event.detail.value)];
                      if (character) chooseCharacter(player.id, character.id);
                    }}
                  >
                    <View className='pickerField'><Text className='pickerValue'>{CHARACTER_BY_ID.get(selectedId ?? '')?.name ?? '选择角色'}</Text></View>
                  </Picker>
                </View>
              );
            })}
          </View>

          {drunkPlayerId && (
            <View className='specialRule'>
              <Text className='fieldLabel'>酒鬼展示身份</Text>
              <View className='choiceRow'>
                {unusedTownsfolk.map((character) => (
                  <Button
                    className={`choiceButton ${draft.shownCharacters[drunkPlayerId] === character.id ? 'choiceButtonSelected' : ''}`}
                    key={character.id}
                    onClick={() => chooseDrunkShownCharacter(character.id)}
                  >
                    <Text className={draft.shownCharacters[drunkPlayerId] === character.id ? 'buttonLabelSelected' : 'buttonLabelDark'}>{character.name}</Text>
                  </Button>
                ))}
              </View>
            </View>
          )}

          {fortuneTellerPlayerId && (
            <View className='specialRule'>
              <Text className='fieldLabel'>占卜师干扰项</Text>
              <View className='choiceRow'>
                {redHerringCandidates.map((player) => (
                  <Button
                    className={`choiceButton ${draft.fortuneTellerRedHerringId === player.id ? 'choiceButtonSelected' : ''}`}
                    key={player.id}
                    onClick={() => chooseRedHerring(player.id)}
                  >
                    <Text className={draft.fortuneTellerRedHerringId === player.id ? 'buttonLabelSelected' : 'buttonLabelDark'}>{player.name}</Text>
                  </Button>
                ))}
              </View>
            </View>
          )}

          <View className='specialRule'>
            <Text className='fieldLabel'>小恶魔伪装身份（{draft.demonBluffCharacterIds.length} / 3）</Text>
            <Text className='sectionDescription'>首夜只私下告知小恶魔；不能是在场身份或酒鬼拿到的展示身份。</Text>
            <View className='choiceRow'>
              {eligibleDemonBluffs.map((character) => {
                const selected = draft.demonBluffCharacterIds.includes(character.id);
                return (
                  <Button
                    className={`choiceButton ${selected ? 'choiceButtonSelected' : ''}`}
                    key={character.id}
                    onClick={() => toggleDemonBluff(character.id)}
                  >
                    <Text className={selected ? 'buttonLabelSelected' : 'buttonLabelDark'}>{character.name}</Text>
                  </Button>
                );
              })}
            </View>
          </View>

          <Text className={`validationText ${validation?.ok ? 'validationOkay' : ''}`}>
            {validation?.ok
              ? assignmentAllowed
                ? '角色配置合法，可以发放身份。'
                : `角色配置合法；等待玩家准备（${readiness.readyCount} / ${readiness.totalCount}）。`
              : setupErrorMessage(validation?.code ?? '')}
          </Text>
          {localError && <Text className='validationText'>{localError}</Text>}
          <Button className='commandButton fullWidthButton' disabled={busy || !assignmentAllowed} onClick={confirmAssignment}>确认并发放身份</Button>
        </View>
      )}

      {panel === 'seats' && room.storytellerId && !frozen && !isStoryteller && (
        <View className='emptyBand'>
          <Text className='emptyBandTitle'>说书人正在配置身份</Text>
          <Text className='emptyBandText'>身份发放后会仅在你的设备上显示。</Text>
        </View>
      )}

      {panel === 'roles' && frozen && isStoryteller && (
        <View className='sectionBand'>
          <Text className='sectionHeading'>最终配置</Text>
          <Text className='sectionDescription'>身份确认 {characterConfirmation.confirmedCount} / {characterConfirmation.totalCount}；全员确认后才能开始首夜。</Text>
          {projectedRedHerring && (
            <View className='specialRule'>
              <Text className='fieldLabel'>占卜师干扰项</Text>
              <Text className='sectionDescription'>
                {projectedRedHerringIndex + 1} 号 · {projectedRedHerring.name || `座位 ${projectedRedHerringIndex + 1}`}
              </Text>
            </View>
          )}
          <View className='specialRule'>
            <Text className='fieldLabel'>小恶魔伪装身份</Text>
            <Text className='sectionDescription'>
              {(room.demonBluffCharacterIds ?? []).map((characterId) => CHARACTER_BY_ID.get(characterId)?.name ?? characterId).join('、')}
            </Text>
          </View>
          <View className='playerList'>
            {room.players.map((player, index) => (
              <View className='playerRow' key={player.id}>
                <Text className='seatNumber'>{String(index + 1).padStart(2, '0')}</Text>
                <View className='playerBody'>
                  <Text className='playerName'>{player.name}</Text>
                  <Text className='playerMeta'>
                    {characterDisplayName(player.character) ?? '未分配'}
                    {player.shownCharacter ? ` · 展示为 ${characterDisplayName(player.shownCharacter)}` : ''}
                  </Text>
                </View>
              </View>
            ))}
          </View>
        </View>
      )}

      {panel === 'manage' && isCreator && !frozen && <View className='sectionBand'>
        <Text className='sectionHeading'>成员管理</Text>
        <Text className='sectionDescription'>更换说书人后，全员需要重新准备。</Text>
        <View className='playerList'>{room.players.map((player) => <View className='playerRow' key={player.id}>
          <View className='playerBody'><Text className='playerName'>{player.name}{player.id === playerId ? '（你）' : ''}</Text></View>
          <Button className='quietButton' disabled={busy} onClick={() => {
            if (!room.storytellerId) { setStoryteller(player.id); return; }
            void Taro.showModal({ title: '更换说书人？', content: `${player.name} 将成为说书人，原说书人回到该座位，所有玩家需要重新准备。`, confirmText: '确认更换' })
              .then((result) => { if (result.confirm) setStoryteller(player.id); });
          }}>指定说书人</Button>
          {player.id !== playerId && <Button className='iconButton' aria-label={`移出${player.name}`} disabled={busy} onClick={() => kickPlayer(player.id)}><Text className='iconGlyph'>×</Text></Button>}
        </View>)}</View>
      </View>}
      {panel === 'roles' && !frozen && !draft && <View className='emptyBand'><Text className='emptyBandTitle'>等待玩家入座</Text><Text className='emptyBandText'>5–15 名玩家到齐后即可配置角色。</Text></View>}
      </View>
      <View className='workspaceAside'>
      <View className='sectionBand setupNextStep'>
        <Text className='sectionHeading'>这一局</Text>
        <Text className='sectionDescription'>暗流涌动 · 说书人 {storytellerLabel}</Text>
        <Text className='sectionDescription'>{nextStep}</Text>
        <Button className='secondaryButton fullWidthButton' disabled={busy} onClick={copyInvite}>邀请朋友入座</Button>
        {isStoryteller && panel !== 'roles' && <Button className='quietButton fullWidthButton' onClick={() => setActivePanel('roles')}>{frozen ? '查看最终配置' : '配置角色'}</Button>}
        {isCreator && panel !== 'manage' && <Button className='quietButton fullWidthButton' onClick={() => setActivePanel('manage')}>管理房间与座次</Button>}
      </View>
      <View className='sectionBand'>
        {panel === 'manage' && isCreator && ownershipCandidates.length > 0 && (
          <Picker mode='selector' range={ownershipCandidates.map((member) => member.name)} disabled={busy} onChange={confirmOwnershipTransfer}>
            <Button className='quietButton' disabled={busy}><Text className='buttonLabelDark'>选择成员 · 转让房主</Text></Button>
          </Picker>
        )}
        <Button className={isCreator ? 'dangerButton' : 'quietButton'} disabled={busy || (!isCreator && frozen)} onClick={confirmExit}>
          <Text className={isCreator ? 'buttonLabelLight' : 'buttonLabelDark'}>{isCreator ? '关闭房间' : '离开房间'}</Text>
        </Button>
      </View>
      </View>
      </View>
    </SessionShell>
  );
}

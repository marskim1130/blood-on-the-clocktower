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
import { characterDisplayAbility, characterDisplayName } from '../../lib/character-display';
import { useRoomSession } from '../../lib/room-session-store';
import { SESSION_ROUTES } from '../../lib/session-routing';
import { useSessionRoute } from '../../lib/use-session-route';
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
  };
  return labels[code] ?? '当前角色配置不合法。';
}

export default function GameSetupPage() {
  const room = useRoomSession((state) => state.experience.roomState);
  const identityStatus = useRoomSession((state) => state.experience.identityStatus);
  const playerId = useRoomSession((state) => state.playerId);
  const pendingCommand = useRoomSession((state) => state.pendingCommand);
  const setStoryteller = useRoomSession((state) => state.setStoryteller);
  const assignCharacters = useRoomSession((state) => state.assignCharacters);
  const startGame = useRoomSession((state) => state.startGame);
  const kickPlayer = useRoomSession((state) => state.kickPlayer);
  const leaveRoom = useRoomSession((state) => state.leaveRoom);
  const closeRoom = useRoomSession((state) => state.closeRoom);
  const [draft, setDraft] = useState<ScriptSetup | null>(null);
  const [localError, setLocalError] = useState('');

  useSessionRoute(SESSION_ROUTES.setup);

  const playerKey = room?.players.map((player) => player.id).join('|') ?? '';
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
    setDraft({ ...draft, assignments, shownCharacters });
  }

  function chooseDrunkShownCharacter(characterId: string): void {
    if (!draft || !drunkPlayerId) return;
    setDraft({ ...draft, shownCharacters: { [drunkPlayerId]: characterId } });
  }

  function chooseRedHerring(targetPlayerId: string): void {
    if (!draft) return;
    setDraft({ ...draft, fortuneTellerRedHerringId: targetPlayerId || null });
  }

  function confirmAssignment(): void {
    if (!draft || !validation?.ok) return;
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
      );
    });
  }

  function copyInvite(): void {
    if (!room) return;
    const path = `${SESSION_ROUTES.lobby}?roomId=${encodeURIComponent(room.roomId)}`;
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

  if (!room) {
    return <SessionShell eyebrow='游戏设置' title='恢复准备状态'><LoadingState /></SessionShell>;
  }

  const storytellerLabel = projectedStorytellerName || (room.storytellerId ? '已设置' : '未设置');
  const actualPlayerCount = Math.max(0, room.players.length - (room.storytellerId ? 0 : 1));

  return (
    <SessionShell
      eyebrow='游戏设置'
      title={frozen ? '身份已发放' : '组建这一局'}
      description={frozen ? '等待说书人开始首夜。' : '确认成员、说书人和角色配置。'}
      actions={
        frozen && isStoryteller
          ? <Button className='commandButton' disabled={busy} onClick={startGame}>开始游戏</Button>
          : <Button className='quietButton' disabled={busy} onClick={copyInvite}><Text className='buttonLabelDark'>复制邀请</Text></Button>
      }
    >
      <View className='sectionBand'>
        <Text className='sectionHeading'>房间概况</Text>
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
        </View>
      </View>

      <View className='sectionBand'>
        <Text className='sectionHeading'>成员与座位</Text>
        <View className='playerList'>
          {room.players.map((player, index) => (
            <View className='playerRow' key={player.id}>
              <Text className='seatNumber'>{String(index + 1).padStart(2, '0')}</Text>
              <View className='playerBody'>
                <Text className='playerName'>{player.name || `座位 ${index + 1}`}{player.id === playerId ? '（你）' : ''}</Text>
                <Text className='playerMeta'>{characterDisplayName(player.character) ?? (frozen ? '身份已私下发放' : '等待配置')}</Text>
              </View>
              {!room.storytellerId && isCreator && (
                <Button className='quietButton' disabled={busy} onClick={() => setStoryteller(player.id)}><Text className='buttonLabelDark'>指定说书人</Text></Button>
              )}
              {isCreator && !frozen && player.id !== playerId && (
                <Button className='iconButton' aria-label={`移出${player.name}`} disabled={busy} onClick={() => kickPlayer(player.id)}><Text className='iconGlyph'>×</Text></Button>
              )}
            </View>
          ))}
        </View>
      </View>

      {!room.storytellerId && (
        <View className='emptyBand'>
          <Text className='emptyBandTitle'>{isCreator ? '请选择一名说书人' : '等待房主指定说书人'}</Text>
        </View>
      )}

      {room.storytellerId && !frozen && isStoryteller && draft && (
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

          <Text className={`validationText ${validation?.ok ? 'validationOkay' : ''}`}>
            {validation?.ok ? '角色配置合法，可以发放身份。' : setupErrorMessage(validation?.code ?? '')}
          </Text>
          {localError && <Text className='validationText'>{localError}</Text>}
          <Button className='commandButton fullWidthButton' disabled={busy || !validation?.ok} onClick={confirmAssignment}>确认并发放身份</Button>
        </View>
      )}

      {room.storytellerId && !frozen && !isStoryteller && (
        <View className='emptyBand'>
          <Text className='emptyBandTitle'>说书人正在配置身份</Text>
          <Text className='emptyBandText'>身份发放后会仅在你的设备上显示。</Text>
        </View>
      )}

      {frozen && !isStoryteller && visibleSelf?.character && (
        <View className='sectionBand'>
          <Text className='sectionHeading'>你的身份</Text>
          <View className='rolePanel privateRole'>
            <Text className='roleName'>{characterDisplayName(visibleSelf.character)}</Text>
            <Text className='roleAbility'>{characterDisplayAbility(visibleSelf.character)}</Text>
          </View>
        </View>
      )}

      {frozen && isStoryteller && (
        <View className='sectionBand'>
          <Text className='sectionHeading'>最终配置</Text>
          {projectedRedHerring && (
            <View className='specialRule'>
              <Text className='fieldLabel'>占卜师干扰项</Text>
              <Text className='sectionDescription'>
                {projectedRedHerringIndex + 1} 号 · {projectedRedHerring.name || `座位 ${projectedRedHerringIndex + 1}`}
              </Text>
            </View>
          )}
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

      <View className='sectionBand'>
        <Button className={isCreator ? 'dangerButton' : 'quietButton'} disabled={busy || (!isCreator && frozen)} onClick={confirmExit}>
          <Text className={isCreator ? 'buttonLabelLight' : 'buttonLabelDark'}>{isCreator ? '关闭房间' : '离开房间'}</Text>
        </Button>
      </View>
    </SessionShell>
  );
}

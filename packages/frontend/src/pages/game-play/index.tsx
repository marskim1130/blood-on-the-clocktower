import { Button, Picker, Text, Textarea, View } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { useEffect, useMemo, useState } from 'react';
import { TROUBLE_BREWING_SCRIPT } from '@clocktower/core';
import { PrivateCharacterCard } from '../../components/private-character-card';
import { PrivateNightResult } from '../../components/private-night-result';
import { LoadingState, SessionShell } from '../../components/session-shell';
import { characterDisplayName, nightResultDisplay } from '../../lib/character-display';
import {
  selectDeathRecords,
  selectActingPlayerId,
  selectGamePhase,
  selectGhostVotes,
  selectVisibleCharacter,
} from '../../lib/room-experience';
import { useRoomSession } from '../../lib/room-session-store';
import { SESSION_ROUTES } from '../../lib/session-routing';
import { useSessionRoute } from '../../lib/use-session-route';
import { eventValue, type InputEvent } from '../../lib/utils';
import { currentVoterId, executionProposalText } from './utils';
import './index.css';

interface PickerChangeEvent {
  readonly detail: { readonly value: string | number };
}

const PHASE_LABELS = {
  setup: '准备',
  day: '白天',
  voting: '投票',
  night: '夜晚',
  finished: '游戏结束',
} as const;

const ACTION_LABELS: Readonly<Record<string, string>> = {
  poison: '投毒',
  protect: '保护',
  kill: '击杀',
  learn_townsfolk: '确认镇民信息',
  learn_outsider: '确认外来者信息',
  learn_minion: '确认爪牙信息',
  learn_evil_pairs: '确认邪恶相邻对数',
  learn_evil_neighbors: '确认邪恶邻居',
  check_demon: '查验恶魔',
  learn_executed: '确认今日处决身份',
  learn_died: '确认死亡触发',
  learn_master: '选择主人',
  learn_demon: '告知恶魔',
  show_grimoire: '查看私密魔典',
  choose_player: '选择玩家',
  none: '无需行动',
};

const ACTION_PROMPTS: Readonly<Record<string, string>> = {
  poison: '选择今夜中毒的玩家。',
  protect: '选择今夜受到保护的玩家。',
  kill: '选择今夜被攻击的玩家。',
  learn_townsfolk: '向当前角色提供镇民信息。',
  learn_outsider: '向当前角色提供外来者信息。',
  learn_minion: '告知恶魔哪些玩家是爪牙。',
  learn_evil_pairs: '告知当前角色邪恶相邻对数。',
  learn_evil_neighbors: '告知当前角色存活邪恶邻居数量。',
  check_demon: '选择两名玩家并告知其中是否包含恶魔。',
  learn_executed: '告知当前角色今天被处决玩家的身份。',
  learn_died: '处理当前角色死亡后获得的信息。',
  learn_master: '为管家选择今夜的主人。',
  learn_demon: '告知爪牙哪位玩家是恶魔。',
  show_grimoire: '请求查看魔典，等待说书人确认后按住阅读，已阅后闭眼。',
  choose_player: '按当前角色能力选择玩家。',
  none: '本步骤无需选择目标。',
};

const WAKE_TYPE_LABELS: Readonly<Record<string, string>> = {
  townsfolk: '镇民',
  outsider: '外来者',
  minion: '爪牙',
  demon: '恶魔',
};

const DEATH_CAUSES = [
  { id: 'execution', label: '处决' },
  { id: 'night_kill', label: '夜间死亡' },
  { id: 'ability', label: '能力致死' },
] as const;

const DEATH_CAUSE_LABELS: Readonly<Record<string, string>> = Object.fromEntries(
  DEATH_CAUSES.map((cause) => [cause.id, cause.label]),
);

function playerName(players: readonly { readonly id: string; readonly name: string }[], id: string): string {
  const index = players.findIndex((player) => player.id === id);
  return players[index]?.name || (index >= 0 ? `座位 ${index + 1}` : '未知玩家');
}

function playerChoiceLabel(players: readonly { readonly id: string; readonly name: string }[], id: string): string {
  const index = players.findIndex((player) => player.id === id);
  const name = players[index]?.name || '未命名玩家';
  return index >= 0 ? `${index + 1} 号 · ${name}` : name;
}

export default function GamePlayPage() {
  const experience = useRoomSession((state) => state.experience);
  const room = experience.roomState;
  const playerId = useRoomSession((state) => state.playerId);
  const pendingCommand = useRoomSession((state) => state.pendingCommand);
  const nominate = useRoomSession((state) => state.nominate);
  const castVote = useRoomSession((state) => state.castVote);
  const recordVote = useRoomSession((state) => state.recordVote);
  const advanceNominationStage = useRoomSession((state) => state.advanceNominationStage);
  const controlNominationTimer = useRoomSession((state) => state.controlNominationTimer);
  const expireNominationTimer = useRoomSession((state) => state.expireNominationTimer);
  const [clockNow, setClockNow] = useState(Date.now());
  const resolveNomination = useRoomSession((state) => state.resolveNomination);
  const useSlayerAbility = useRoomSession((state) => state.useSlayerAbility);
  const killPlayer = useRoomSession((state) => state.killPlayer);
  const finalizeDay = useRoomSession((state) => state.finalizeDay);
  const submitNightAction = useRoomSession((state) => state.submitNightAction);
  const confirmNightAction = useRoomSession((state) => state.confirmNightAction);
  const acknowledgeNightAction = useRoomSession((state) => state.acknowledgeNightAction);
  const skipNightAction = useRoomSession((state) => state.skipNightAction);
  const prepareDawn = useRoomSession((state) => state.prepareDawn);
  const confirmDawn = useRoomSession((state) => state.confirmDawn);
  const endGame = useRoomSession((state) => state.endGame);
  const closeRoom = useRoomSession((state) => state.closeRoom);
  const [nomineeId, setNomineeId] = useState('');
  const [slayerTargetId, setSlayerTargetId] = useState('');
  const [deathTargetId, setDeathTargetId] = useState('');
  const [deathCause, setDeathCause] = useState<(typeof DEATH_CAUSES)[number]['id']>('execution');
  const [nightTargetIds, setNightTargetIds] = useState<readonly string[]>([]);
  const [dawnDeathIds, setDawnDeathIds] = useState<readonly string[]>([]);
  const [nightResult, setNightResult] = useState('');
  const [endDescription, setEndDescription] = useState('');
  const [localError, setLocalError] = useState('');
  const [activeTab, setActiveTab] = useState<'action' | 'seats' | 'storyteller'>('action');

  useSessionRoute(SESSION_ROUTES.play);

  const phase = selectGamePhase(experience);
  const isStoryteller = room?.storytellerId === playerId;
  const selectedTab = activeTab === 'storyteller' && !isStoryteller ? 'action' : activeTab;
  const isCreator = room?.creatorId === playerId;
  const self = room?.players.find((player) => player.id === playerId) ?? null;
  const visibleCharacter = selectVisibleCharacter(experience, playerId);
  const alivePlayers = useMemo(() => room?.players.filter((player) => player.isAlive) ?? [], [room]);
  const deadPlayers = useMemo(() => room?.players.filter((player) => !player.isAlive) ?? [], [room]);
  const ghostVotes = useMemo(() => selectGhostVotes(experience), [experience]);
  const deathRecords = useMemo(() => selectDeathRecords(experience), [experience]);
  const currentNomination = room?.nomination;
  const nightSteps = room?.nightWakeSteps ?? [];
  const currentNightIndex = room?.currentNightWakeIndex ?? 0;
  const currentNightStep = room?.currentNightWakeStep;
  const nightTurnStatus = room?.nightTurnStatus ?? '';
  const pendingNightAction = room?.pendingNightAction;
  const confirmedNightAction = room?.confirmedNightAction;
  const pendingNightTargetsKey = pendingNightAction?.targetIds.join('|') ?? '';
  const nightNumber = room
    ? (room as typeof room & { readonly nightNumber?: number }).nightNumber ?? Math.max(1, room.dayNumber)
    : 0;
  const busy = pendingCommand !== null;
  const deadline = currentNomination?.deadlineUnixMs ?? 0;
  useEffect(() => { setActiveTab('action'); }, [phase]);
  useEffect(() => {
    if (!currentNomination) return;
    const timer = setInterval(() => setClockNow(Date.now()), 250);
    return () => clearInterval(timer);
  }, [Boolean(currentNomination)]);
  useEffect(() => {
    if (!isStoryteller || !deadline || currentNomination?.paused || busy) return;
    const timer = setTimeout(() => expireNominationTimer(deadline), Math.max(100, deadline - Date.now() + 150));
    return () => clearTimeout(timer);
  }, [deadline, isStoryteller, currentNomination?.paused, busy]);

  useEffect(() => {
    setNightTargetIds(pendingNightAction?.targetIds ?? []);
    setNightResult(pendingNightAction?.result ?? '');
    setLocalError('');
  }, [currentNightIndex, currentNightStep?.actionType, pendingNightAction?.actorId, pendingNightTargetsKey]);

  const proposedDawnDeathsKey = room?.pendingDawnDeathIds?.join('|') ?? '';
  useEffect(() => {
    if (room?.dawnReviewPending) setDawnDeathIds(room.pendingDawnDeathIds ?? []);
  }, [room?.dawnReviewPending, proposedDawnDeathsKey]);

  if (!room) {
    return <SessionShell eyebrow='进行中的游戏' title='恢复游戏状态'><LoadingState /></SessionShell>;
  }

  const executionThreshold = Math.ceil(alivePlayers.length / 2);
  const voted = Boolean(currentNomination && Object.hasOwn(currentNomination.votes, playerId));
  const currentVoterPlayerId = currentNomination ? currentVoterId(currentNomination) : undefined;
  const ballotComplete = Boolean(
    currentNomination && currentNomination.currentVoterIndex >= currentNomination.voterOrder.length,
  );
  const votesOpen = (!currentNomination?.stage || currentNomination.stage === 'voting') && !currentNomination?.paused;
  const canVoteNo = Boolean(votesOpen && self && !voted && currentVoterPlayerId === playerId);
  const canVoteYes = Boolean(canVoteNo && (self?.isAlive || ghostVotes.has(playerId)));
  const canNominate = phase === 'day' && self?.isAlive === true;
  const canUseSlayer = phase === 'day' && self?.isAlive === true && visibleCharacter?.id === 'slayer';
  const currentActionLabel = currentNightStep
    ? ACTION_LABELS[currentNightStep.actionType] ?? currentNightStep.actionType
    : '全部步骤已完成';
  const currentActionPrompt = currentNightStep
    ? ACTION_PROMPTS[currentNightStep.actionType] ?? '按当前步骤完成说书人指示。'
    : '';
  const actingPlayerId = selectActingPlayerId(room.players, currentNightStep?.characterId);
  const excludesActor = currentNightStep?.characterId === 'monk' || currentNightStep?.characterId === 'butler';
  const canPrepareDawn = currentNightIndex >= nightSteps.length && !pendingNightAction && !confirmedNightAction;
  const executionProposal = executionProposalText(
    room.players,
    room.executionCandidateId,
    room.executionCandidateVotes,
    room.executionTied,
  );

  function toggleNightTarget(targetId: string): void {
    if (!currentNightStep) return;
    setLocalError('');
    setNightTargetIds((current) => {
      if (current.includes(targetId)) return current.filter((id) => id !== targetId);
      if (current.length >= currentNightStep.maxTargets) return current;
      return [...current, targetId];
    });
  }

  function submitCurrentNightAction(): void {
    if (!currentNightStep) return;
    if (nightTargetIds.length < currentNightStep.minTargets) {
      setLocalError(`至少选择 ${currentNightStep.minTargets} 名玩家`);
      return;
    }
    if (nightTargetIds.length > currentNightStep.maxTargets) {
      setLocalError(`最多选择 ${currentNightStep.maxTargets} 名玩家`);
      return;
    }
    submitNightAction(
      currentNightStep.actionType,
      nightTargetIds,
      isStoryteller ? nightResult.trim() || undefined : undefined,
    );
  }

  function confirmCurrentNightAction(): void {
    if (!currentNightStep) return;
    if (nightTargetIds.length < currentNightStep.minTargets) {
      setLocalError(`至少选择 ${currentNightStep.minTargets} 名玩家`);
      return;
    }
    if (nightTargetIds.length > currentNightStep.maxTargets) {
      setLocalError(`最多选择 ${currentNightStep.maxTargets} 名玩家`);
      return;
    }
    confirmNightAction(nightTargetIds, nightResult.trim() || undefined);
  }

  function confirmSkipNightStep(): void {
    void Taro.showModal({
      title: '跳过当前夜间步骤？',
      content: confirmedNightAction
        ? '已确认的效果会保留，仅跳过玩家“已阅”。'
        : '未确认的玩家选择会被丢弃，本步骤不产生新效果。',
      confirmText: '确认跳过',
      confirmColor: '#7b2028',
    }).then((result) => {
      if (result.confirm) skipNightAction();
    });
  }

  function toggleDawnDeath(playerID: string): void {
    setDawnDeathIds((current) => current.includes(playerID)
      ? current.filter((id) => id !== playerID)
      : [...current, playerID]);
  }

  function confirmDawnDeaths(): void {
    const names = dawnDeathIds.map((id) => playerName(room?.players ?? [], id));
    void Taro.showModal({
      title: '锁定黎明死亡？',
      content: names.length > 0 ? `锁定死亡：${names.join('、')}。如有守鸦人死亡，将先完成其私密行动再统一天亮。` : '确认今夜无人死亡并天亮。',
      confirmText: '确认锁定',
      confirmColor: '#7b2028',
    }).then((result) => {
      if (result.confirm) confirmDawn(dawnDeathIds);
    });
  }

  function phaseCounter(): string {
    if (phase === 'night') return `第 ${nightNumber} 夜`;
    if (phase === 'day' || phase === 'voting') return `第 ${room?.dayNumber ?? 0} 天`;
    return '';
  }

  function confirmCloseRoom(): void {
    void Taro.showModal({
      title: '关闭房间？',
      content: '关闭后全员会立即退出，当前对局不能再恢复。',
      confirmText: '关闭房间',
      confirmColor: '#7b2028',
    }).then((result) => {
      if (result.confirm) closeRoom();
    });
  }

  function confirmFinalizeDay(): void {
    void Taro.showModal({
      title: '确认结束白天？',
      content: executionProposal,
      confirmText: room?.executionCandidateId ? '确认处决' : '确认入夜',
      confirmColor: '#7b2028',
    }).then((result) => {
      if (result.confirm) finalizeDay();
    });
  }

  return (
    <SessionShell
      eyebrow='进行中的游戏'
      title={PHASE_LABELS[phase]}
      description={isStoryteller ? '说书人控制面板' : visibleCharacter ? '你的身份已私下发放，可随时按住查看。' : '等待说书人推进游戏'}
    >
      <View className='phaseHeader gameStatusBar'>
        <View className='gameStatusTitle'>
          <Text className='phaseName'>{PHASE_LABELS[phase]}</Text>
          <Text className='phaseCounter'>{phaseCounter()}</Text>
        </View>
        <View className='gameStatusMetrics'>
          <Text className='tag tagGood'>存活 {alivePlayers.length}</Text>
          <Text className='tag'>死亡 {deadPlayers.length}</Text>
          <Text className='tag'>上台门槛 {executionThreshold} 票</Text>
        </View>
      </View>

      <View className='pageTabs'>
        <Button aria-pressed={selectedTab === 'action'} className={`pageTab ${selectedTab === 'action' ? 'pageTabActive' : ''}`} onClick={() => setActiveTab('action')}>行动</Button>
        <Button aria-pressed={selectedTab === 'seats'} className={`pageTab ${selectedTab === 'seats' ? 'pageTabActive' : ''}`} onClick={() => setActiveTab('seats')}>座位 · {room.players.length}</Button>
        {isStoryteller && <Button aria-pressed={selectedTab === 'storyteller'} className={`pageTab ${selectedTab === 'storyteller' ? 'pageTabActive' : ''}`} onClick={() => setActiveTab('storyteller')}>说书人</Button>}
      </View>

      <View className='surfaceGrid gameWorkspace'>
        <View className='workspaceMain'>
          {selectedTab === 'seats' && (
          <View className='sectionBand'>
            <Text className='sectionHeading'>镇民座位</Text>
            <Text className='sectionDescription'>按顺时针座次排列；死亡玩家仍可参与讨论。</Text>
            <View className='seatGrid'>
              {room.players.map((player, index) => {
                const death = deathRecords[player.id];
                return (
                  <View className={`seatCard ${player.isAlive ? '' : 'playerDead'} ${player.id === playerId ? 'seatCardSelf' : ''}`} key={player.id}>
                    <Text className='seatAvatar'>{String(index + 1).padStart(2, '0')}</Text>
                    <View className='playerBody'>
                      <Text className='playerName'>{player.name || `座位 ${index + 1}`}{player.id === playerId ? '（你）' : ''}</Text>
                      <Text className='playerMeta'>
                        {player.isAlive
                          ? '存活'
                          : `${isStoryteller ? `${DEATH_CAUSE_LABELS[death?.cause ?? ''] ?? '死亡'} · ` : ''}第 ${death?.dayNumber ?? room.dayNumber} 天死亡`}
                        {isStoryteller && player.character ? ` · ${characterDisplayName(player.character)}` : ''}
                        {player.poisonedUntil !== undefined ? ` · 中毒至第 ${player.poisonedUntil} 天黄昏` : ''}
                      </Text>
                    </View>
                    {!player.isAlive && <Text className={`tag ${ghostVotes.has(player.id) ? 'tagGood' : ''}`}>{ghostVotes.has(player.id) ? '幽灵票可用' : '幽灵票已用'}</Text>}
                  </View>
                );
              })}
            </View>
          </View>
          )}

          {selectedTab === 'action' && phase === 'day' && (
            <View className='sectionBand'>
              <Text className='sectionHeading'>提名</Text>
              <Text className='sectionDescription'>提名后依次进行控方陈述、辩方发言和顺时针逐席投票；从被提名人的下一席开始，被提名人最后投票。</Text>
              {canNominate ? (
              <View>
              <View className='choiceRow'>
                {alivePlayers.filter((player) => player.id !== playerId).map((player) => (
                  <Button
                    className={`choiceButton ${nomineeId === player.id ? 'choiceButtonSelected' : ''}`}
                    key={player.id}
                    disabled={!canNominate || busy}
                    onClick={() => setNomineeId(player.id)}
                  >
                    <Text className={nomineeId === player.id ? 'buttonLabelSelected' : 'buttonLabelDark'}>{playerChoiceLabel(room.players, player.id)}</Text>
                  </Button>
                ))}
              </View>
              <Button className='commandButton fullWidthButton' disabled={!canNominate || !nomineeId || busy} onClick={() => nominate(nomineeId)}>发起提名</Button>
              </View>
              ) : (
                <View className='emptyBand'>
                  <Text className='emptyBandTitle'>{isStoryteller ? '等待玩家发起提名' : '继续参与讨论与投票'}</Text>
                  <Text className='emptyBandText'>{isStoryteller ? '提名开始后，将在这里显示发言和逐席投票控制。' : '死亡后不能发起提名；你仍可以讨论，并在适当时机使用幽灵票。'}</Text>
                </View>
              )}
            </View>
          )}

          {selectedTab === 'action' && phase === 'voting' && currentNomination && (
            <View className='sectionBand'>
              <View className='nominationPanel'>
                <Text className='nominationTitle'>
                  {playerName(room.players, currentNomination.nominatorId)} 提名 {playerName(room.players, currentNomination.nomineeId)}
                </Text>
                <Text className='voteThreshold'>处决需要至少 {executionThreshold} 张赞成票</Text>
                <Text className='voteTurn'>
                  {currentNomination.stage === 'accusation' ? '控方陈述' : currentNomination.stage === 'defense' ? '辩方发言' : '顺时针投票'}
                  {currentNomination.paused ? ' · 已暂停' : deadline > 0 ? ` · 剩余 ${Math.max(0, Math.ceil((deadline - clockNow) / 1000))} 秒` : ''}
                </Text>
                <Text className='voteTurn'>
                  {ballotComplete
                    ? '全部席位已完成表态，等待说书人结算。'
                    : `当前轮到 ${playerName(room.players, currentVoterPlayerId ?? '')}（${currentNomination.currentVoterIndex + 1}/${currentNomination.voterOrder.length}）`}
                </Text>
              </View>
              <View className='commandRow'>
                <Button className='secondaryButton' disabled={!canVoteYes || busy} onClick={() => castVote(true)}>赞成处决</Button>
                <Button className='dangerButton' disabled={!canVoteNo || busy} onClick={() => castVote(false)}>反对处决</Button>
              </View>
              {self && !self.isAlive && (
                <Text className='sectionDescription'>
                  {ghostVotes.has(playerId) ? '只有投赞成时才会消耗你的幽灵票。' : '你的幽灵票已经使用，本轮只能投反对。'}
                </Text>
              )}
              <View className='playerList'>
                {currentNomination.voterOrder.map((voterId) => {
                  const hasDecision = Object.hasOwn(currentNomination.votes, voterId);
                  const decision = currentNomination.votes[voterId];
                  return (
                  <View className={`voteRecord ${voterId === currentVoterPlayerId ? 'voteRecordCurrent' : ''}`} key={voterId}>
                    <Text>{playerName(room.players, voterId)}</Text>
                    <Text className={hasDecision ? (decision ? 'voteDecisionYes' : 'voteDecisionNo') : 'voteDecisionWaiting'}>
                      {hasDecision ? (decision ? '赞成' : '反对') : voterId === currentVoterPlayerId ? '正在表态' : '等待'}
                    </Text>
                  </View>
                  );
                })}
              </View>
              {isStoryteller && (
                <View>
                  {votesOpen && currentVoterPlayerId && (
                    <Button className='dangerButton fullWidthButton' disabled={busy} onClick={() => recordVote(currentVoterPlayerId, false)}>
                      当前席无响应／代记反对
                    </Button>
                  )}
                  <View className='commandRow'>
                    <Button className='secondaryButton' disabled={busy || ballotComplete} onClick={() => controlNominationTimer(currentNomination.paused ? 'resume' : 'pause')}>{currentNomination.paused ? '继续计时' : '暂停计时'}</Button>
                    <Button className='secondaryButton' disabled={busy || ballotComplete} onClick={() => controlNominationTimer('restart')}>重开当前计时</Button>
                    {currentNomination.stage !== 'voting' && <Button className='commandButton' disabled={busy || Boolean(currentNomination.paused)} onClick={advanceNominationStage}>下一阶段</Button>}
                  </View>
                  <Button className='commandButton fullWidthButton' disabled={busy || !ballotComplete} onClick={resolveNomination}>结算投票</Button>
                </View>
              )}
            </View>
          )}

          {selectedTab === 'storyteller' && phase === 'night' && isStoryteller && (
            <View className='sectionBand'>
              <Text className='sectionHeading'>唤醒顺序</Text>
              <View className='wakeList'>
                {nightSteps.map((step, index) => {
                  const character = TROUBLE_BREWING_SCRIPT.characters.find((item) => item.id === step.characterId);
                  const wakeCharacterLabel = character?.name
                    ?? WAKE_TYPE_LABELS[step.characterType ?? '']
                    ?? WAKE_TYPE_LABELS[step.characterId]
                    ?? '阵营信息';
                  const current = index === currentNightIndex;
                  const done = index < currentNightIndex;
                  return (
                    <View className={`wakeStep ${current ? 'wakeStepCurrent' : ''} ${done ? 'wakeStepDone' : ''}`} key={`${step.order}-${step.characterId}-${step.actionType}`}>
                      <Text className='wakeStepIndex'>{done ? '✓' : step.order}</Text>
                      <View>
                        <Text className='wakeStepName'>{wakeCharacterLabel} · {ACTION_LABELS[step.actionType] ?? '当前行动'}</Text>
                        <Text className='wakeStepPrompt'>{ACTION_PROMPTS[step.actionType] ?? '按当前步骤完成说书人指示。'}</Text>
                      </View>
                    </View>
                  );
                })}
              </View>

            </View>
          )}

          {selectedTab === 'action' && phase === 'night' && isStoryteller && (
            <View className='sectionBand nightCurrentWorkspace'>
              <Text className='sectionHeading'>当前夜间行动</Text>
              <Text className='sectionDescription'>进度 {Math.min(currentNightIndex + 1, nightSteps.length)} / {nightSteps.length} · 完整唤醒顺序见「说书人」</Text>

              {currentNightStep ? (
                <View className='sectionBand storytellerTools'>
                  <Text className='sectionHeading'>{currentActionLabel}</Text>
                  <Text className='sectionDescription'>{currentActionPrompt}</Text>
                  <View className='nightReviewStatus'>
                    <Text className='nightReviewTitle'>
                      {nightTurnStatus === 'awaiting_storyteller'
                        ? `${playerName(room.players, pendingNightAction?.actorId ?? '')} 已提交，等待你复核`
                        : nightTurnStatus === 'awaiting_acknowledgement'
                          ? '裁定已发送，等待当前角色确认已阅'
                          : '等待当前角色提交；断线时可由你代办'}
                    </Text>
                    {confirmedNightAction && (
                      <Text className='nightReviewDetail'>
                        {confirmedNightAction.targetIds.map((id) => playerName(room.players, id)).join('、') || '无目标'}
                        {confirmedNightAction.result ? ` · ${nightResultDisplay(confirmedNightAction.result)}` : ''}
                      </Text>
                    )}
                  </View>
                  {!confirmedNightAction && (
                    <View>
                      {currentNightStep.maxTargets > 0 && (
                        <View className='targetGrid'>
                          {room.players.map((player) => {
                            const selected = nightTargetIds.includes(player.id);
                            const excluded = excludesActor && player.id === actingPlayerId;
                            const atLimit = !selected && nightTargetIds.length >= currentNightStep.maxTargets;
                            return (
                              <Button
                                className={`choiceButton ${selected ? 'choiceButtonSelected' : ''}`}
                                key={player.id}
                                disabled={excluded || atLimit || busy}
                                onClick={() => toggleNightTarget(player.id)}
                              >
                                <Text className={selected ? 'buttonLabelSelected' : 'buttonLabelDark'}>{player.name}{player.isAlive ? '' : '（死亡）'}</Text>
                              </Button>
                            );
                          })}
                        </View>
                      )}
                      <View className='fieldGroup'>
                        <Text className='fieldLabel'>说书人最终结果</Text>
                        <Textarea
                          className='textAreaInput'
                          maxlength={240}
                          value={nightResult}
                          placeholder='可留空，由服务器计算；中毒/醉酒时可手动填写'
                          onInput={(event: InputEvent) => setNightResult(eventValue(event))}
                        />
                      </View>
                      {localError && <Text className='localError'>{localError}</Text>}
                      <Button
                        className='commandButton fullWidthButton'
                        disabled={busy}
                        onClick={nightTurnStatus === 'awaiting_storyteller' ? confirmCurrentNightAction : submitCurrentNightAction}
                      >
                        {nightTurnStatus === 'awaiting_storyteller' ? '确认裁定并发送给玩家' : '说书人代办并推进'}
                      </Button>
                    </View>
                  )}
                  <Button className='dangerButton fullWidthButton' disabled={busy} onClick={confirmSkipNightStep}>跳过当前步骤／玩家已阅</Button>
                </View>
              ) : (
                <View className='feedback feedbackSuccess inlineFeedback'><Text className='feedbackText'>今夜全部步骤已完成。</Text></View>
              )}

              {room.dawnReviewPending ? (
                <View className='sectionBand dawnReviewPanel'>
                  <Text className='sectionHeading'>黎明死亡审核</Text>
                  <Text className='sectionDescription'>系统建议已选中；你可以增删为零人或多人。确认前其他玩家仍看不到结果。</Text>
                  <View className='targetGrid'>
                    {alivePlayers.map((player) => {
                      const selected = dawnDeathIds.includes(player.id);
                      return (
                        <Button
                          className={`choiceButton ${selected ? 'choiceButtonSelected' : ''}`}
                          key={player.id}
                          disabled={busy}
                          onClick={() => toggleDawnDeath(player.id)}
                        >
                          <Text className={selected ? 'buttonLabelSelected' : 'buttonLabelDark'}>{playerChoiceLabel(room.players, player.id)}</Text>
                        </Button>
                      );
                    })}
                  </View>
                  <View className='nightReviewStatus'>
                    <Text className='nightReviewTitle'>{dawnDeathIds.length > 0 ? `将公布 ${dawnDeathIds.length} 人死亡` : '将公布今夜无人死亡'}</Text>
                  </View>
                  <Button className='commandButton fullWidthButton' disabled={busy} onClick={confirmDawnDeaths}>锁定黎明结果</Button>
                </View>
              ) : (room.pendingDawnDeathIds?.length ?? 0) > 0 ? (
                <View className='nightReviewStatus'><Text className='nightReviewTitle'>死亡名单已锁定</Text><Text className='sectionDescription'>等待守鸦人完成死亡触发行动，已阅或由说书人跳过后统一公布天亮。</Text></View>
              ) : (
                <Button className='secondaryButton fullWidthButton' disabled={busy || !canPrepareDawn} onClick={prepareDawn}>生成黎明死亡建议</Button>
              )}
            </View>
          )}

          {selectedTab === 'action' && phase === 'night' && !isStoryteller && currentNightStep && (
            <View className='nightActorPanel'>
              <Text className='nightAwakeLabel'>说书人已唤醒你</Text>
              <Text className='emptyBandTitle'>{currentActionLabel}</Text>
              <Text className='emptyBandText'>{currentActionPrompt}</Text>

              {nightTurnStatus === 'awaiting_player' && (
                <View>
                  {currentNightStep.maxTargets > 0 && (
                    <View className='targetGrid'>
                      {room.players.map((player) => {
                        const selected = nightTargetIds.includes(player.id);
                        const excluded = excludesActor && player.id === actingPlayerId;
                        const atLimit = !selected && nightTargetIds.length >= currentNightStep.maxTargets;
                        return (
                          <Button
                            className={`choiceButton ${selected ? 'choiceButtonSelected' : ''}`}
                            key={player.id}
                            disabled={excluded || atLimit || busy}
                            onClick={() => toggleNightTarget(player.id)}
                          >
                            <Text className={selected ? 'buttonLabelSelected' : 'buttonLabelDark'}>{playerChoiceLabel(room.players, player.id)}{player.isAlive ? '' : '（死亡）'}</Text>
                          </Button>
                        );
                      })}
                    </View>
                  )}
                  {localError && <Text className='localError'>{localError}</Text>}
                  <Button className='commandButton fullWidthButton' disabled={busy} onClick={submitCurrentNightAction}>
                    {currentNightStep.maxTargets > 0 ? '提交私密选择' : '我已准备接收信息'}
                  </Button>
                </View>
              )}

              {nightTurnStatus === 'awaiting_storyteller' && (
                <View className='nightResultCard'>
                  <Text className='nightReviewTitle'>选择已私密提交</Text>
                  <Text className='nightReviewDetail'>
                    {pendingNightAction
                      ? pendingNightAction.targetIds.map((id) => playerName(room.players, id)).join('、') || '无需选择目标'
                      : '同阵营当前步骤已提交'}
                  </Text>
                  <Text className='nightReviewDetail'>请保持安静，等待说书人复核。</Text>
                </View>
              )}

              {nightTurnStatus === 'awaiting_acknowledgement' && confirmedNightAction && (
                <PrivateNightResult
                  resultKey={`${room.nightNumber}-${currentNightStep.actionType}`}
                  roomId={room.roomId}
                  playerName={playerName(room.players, actingPlayerId ?? '')}
                  targets={confirmedNightAction.targetIds.map((id) => playerName(room.players, id)).join('、')}
                  result={confirmedNightAction.result ? nightResultDisplay(confirmedNightAction.result) : '行动已确认'}
                  busy={busy}
                  onAcknowledge={acknowledgeNightAction}
                />
              )}
            </View>
          )}

          {selectedTab === 'action' && phase === 'night' && !isStoryteller && !currentNightStep && (
            <View className='waitingNight'>
              <Text className='emptyBandTitle'>夜晚降临</Text>
              <Text className='emptyBandText'>请闭眼等待说书人唤醒；这里不会显示其他玩家的行动。</Text>
            </View>
          )}
          {selectedTab === 'action' && phase === 'day' && (
            <View className='sectionBand'>
              <Text className='sectionHeading'>日终处决提案</Text>
              <View className={`executionProposal ${room.executionCandidateId ? 'executionProposalActive' : ''}`}>
                <Text className='executionProposalText'>{executionProposal}</Text>
              </View>
              {(room.nominationResults ?? []).filter((result) => result.dayNumber === room.dayNumber).map((result, index) => (
                <View className='voteRecord' key={`${result.dayNumber}-${result.nomineeId}-${index}`}>
                  <Text>{playerName(room.players, result.nomineeId)}</Text>
                  <Text>{result.yesVotes} 票赞成 · 门槛 {result.requiredVotes}</Text>
                </View>
              ))}
              {isStoryteller && (
                <Button className='commandButton fullWidthButton' disabled={busy} onClick={confirmFinalizeDay}>确认日终并进入夜晚</Button>
              )}
            </View>
          )}

          {selectedTab === 'storyteller' && isStoryteller && phase === 'day' && (
            <View className='sectionBand storytellerTools'>
              <Text className='sectionHeading'>说书人工具</Text>
              <View className='fieldGroup'>
                <Text className='fieldLabel'>宣告死亡</Text>
                <View className='choiceRow'>
                  {alivePlayers.map((player) => (
                    <Button className={`choiceButton ${deathTargetId === player.id ? 'choiceButtonSelected' : ''}`} key={player.id} onClick={() => setDeathTargetId(player.id)}>
                      <Text className={deathTargetId === player.id ? 'buttonLabelSelected' : 'buttonLabelDark'}>{player.name}</Text>
                    </Button>
                  ))}
                </View>
              </View>
              <Picker
                mode='selector'
                range={DEATH_CAUSES.map((cause) => cause.label)}
                value={Math.max(0, DEATH_CAUSES.findIndex((cause) => cause.id === deathCause))}
                onChange={(event: PickerChangeEvent) => {
                  const cause = DEATH_CAUSES[Number(event.detail.value)];
                  if (cause) setDeathCause(cause.id);
                }}
              >
                <View className='pickerField'><Text className='pickerValue'>{DEATH_CAUSE_LABELS[deathCause]}</Text></View>
              </Picker>
              <Button className='dangerButton fullWidthButton' disabled={busy || !deathTargetId} onClick={() => killPlayer(deathTargetId, deathCause)}>确认死亡</Button>

              <View className='fieldGroup'>
                <Text className='fieldLabel'>手动结束游戏</Text>
                <Textarea
                  className='textAreaInput'
                  maxlength={240}
                  value={endDescription}
                  placeholder='补充结局说明（可选）'
                  onInput={(event: InputEvent) => setEndDescription(eventValue(event))}
                />
              </View>
              <View className='commandRow'>
                <Button className='secondaryButton' disabled={busy} onClick={() => endGame('good', endDescription)}>善良胜利</Button>
                <Button className='dangerButton' disabled={busy} onClick={() => endGame('evil', endDescription)}>邪恶胜利</Button>
              </View>
            </View>
          )}

        </View>
        <View className='workspaceAside'>
          {selectedTab === 'action' && canUseSlayer && (
            <View className='sectionBand'>
              <Text className='sectionHeading'>杀手能力</Text>
              <View className='choiceRow'>
                {alivePlayers.filter((player) => player.id !== playerId).map((player) => (
                  <Button className={`choiceButton ${slayerTargetId === player.id ? 'choiceButtonSelected' : ''}`} key={player.id} onClick={() => setSlayerTargetId(player.id)}>
                    <Text className={slayerTargetId === player.id ? 'buttonLabelSelected' : 'buttonLabelDark'}>{player.name}</Text>
                  </Button>
                ))}
              </View>
              <Button className='dangerButton fullWidthButton' disabled={busy || !slayerTargetId} onClick={() => useSlayerAbility(slayerTargetId)}>使用能力</Button>
            </View>
          )}

          {(selectedTab === 'action' || selectedTab === 'seats') && visibleCharacter && !isStoryteller && (
            <View className='sectionBand'>
              <Text className='sectionHeading'>你的身份</Text>
              <PrivateCharacterCard
                character={visibleCharacter}
                roomId={room.roomId}
                playerName={self?.name || '未命名玩家'}
              />
            </View>
          )}

          {selectedTab === 'seats' && deadPlayers.length > 0 && (
            <View className='sectionBand'>
              <Text className='sectionHeading'>死亡时间线</Text>
              {room.deaths?.map((death) => (
                <View className='timelineEntry' key={`${death.dayNumber}-${death.playerId}`}>
                  <Text className='timelineTitle'>{playerName(room.players, death.playerId)}</Text>
                  <Text className='timelineMeta'>
                    第 {death.dayNumber} 天死亡{isStoryteller && death.cause ? ` · ${DEATH_CAUSE_LABELS[death.cause] ?? death.cause}` : ''}
                  </Text>
                </View>
              ))}
            </View>
          )}

          {selectedTab === 'storyteller' && isStoryteller && phase === 'night' && (room.nightActions?.length ?? 0) > 0 && (
            <View className='sectionBand'>
              <Text className='sectionHeading'>今夜记录</Text>
              {room.nightActions?.map((action, index) => (
                <View className='nightActionEntry' key={`${action.actionType}-${index}`}>
                  <Text className='timelineTitle'>{ACTION_LABELS[action.actionType] ?? action.actionType}</Text>
                  <Text className='nightActionResult'>
                    {(action.targetIds ?? []).map((id) => playerName(room.players, id)).join('、') || '无目标'}
                    {action.result ? ` · ${nightResultDisplay(action.result)}` : ''}
                  </Text>
                </View>
              ))}
            </View>
          )}

          {isCreator && selectedTab === (isStoryteller ? 'storyteller' : 'seats') && (
            <View className='sectionBand'>
              <Text className='sectionHeading'>房间管理</Text>
              <Button className='dangerButton fullWidthButton' disabled={busy} onClick={confirmCloseRoom}>关闭房间</Button>
            </View>
          )}
        </View>
      </View>
    </SessionShell>
  );
}

import { Button, Picker, Text, Textarea, View } from '@tarojs/components';
import Taro from '@tarojs/taro';
import { useEffect, useMemo, useState } from 'react';
import { TROUBLE_BREWING_SCRIPT } from '@clocktower/core';
import { LoadingState, SessionShell } from '../../components/session-shell';
import { characterDisplayAbility, characterDisplayName, nightResultDisplay } from '../../lib/character-display';
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
  const resolveNomination = useRoomSession((state) => state.resolveNomination);
  const useSlayerAbility = useRoomSession((state) => state.useSlayerAbility);
  const killPlayer = useRoomSession((state) => state.killPlayer);
  const changePhase = useRoomSession((state) => state.changePhase);
  const submitNightAction = useRoomSession((state) => state.submitNightAction);
  const resolveNight = useRoomSession((state) => state.resolveNight);
  const endGame = useRoomSession((state) => state.endGame);
  const closeRoom = useRoomSession((state) => state.closeRoom);
  const [nomineeId, setNomineeId] = useState('');
  const [slayerTargetId, setSlayerTargetId] = useState('');
  const [deathTargetId, setDeathTargetId] = useState('');
  const [deathCause, setDeathCause] = useState<(typeof DEATH_CAUSES)[number]['id']>('execution');
  const [nightTargetIds, setNightTargetIds] = useState<readonly string[]>([]);
  const [nightResult, setNightResult] = useState('');
  const [endDescription, setEndDescription] = useState('');
  const [localError, setLocalError] = useState('');

  useSessionRoute(SESSION_ROUTES.play);

  const phase = selectGamePhase(experience);
  const isStoryteller = room?.storytellerId === playerId;
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
  const nightNumber = room
    ? (room as typeof room & { readonly nightNumber?: number }).nightNumber ?? Math.max(1, room.dayNumber)
    : 0;
  const busy = pendingCommand !== null;

  useEffect(() => {
    setNightTargetIds([]);
    setNightResult('');
    setLocalError('');
  }, [currentNightIndex, currentNightStep?.actionType]);

  if (!room) {
    return <SessionShell eyebrow='进行中的游戏' title='恢复游戏状态'><LoadingState /></SessionShell>;
  }

  const executionThreshold = Math.ceil(alivePlayers.length / 2);
  const voted = Boolean(currentNomination && Object.hasOwn(currentNomination.votes, playerId));
  const canVote = Boolean(self && !voted && (self.isAlive || ghostVotes.has(playerId)));
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
  const canResolveNight = currentNightIndex >= nightSteps.length;

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
    submitNightAction(currentNightStep.actionType, nightTargetIds, nightResult.trim() || undefined);
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

  return (
    <SessionShell
      eyebrow='进行中的游戏'
      title={PHASE_LABELS[phase]}
      description={isStoryteller ? '说书人控制面板' : visibleCharacter ? `你的身份：${characterDisplayName(visibleCharacter)}` : '等待说书人推进游戏'}
    >
      <View className='phaseHeader'>
        <Text className='phaseName'>{PHASE_LABELS[phase]}</Text>
        <Text className='phaseCounter'>{phaseCounter()}</Text>
      </View>

      <View className='gameGrid'>
        <View>
          <View className='sectionBand'>
            <Text className='sectionHeading'>玩家状态</Text>
            <View className='playerList'>
              {room.players.map((player, index) => {
                const death = deathRecords[player.id];
                return (
                  <View className={`playerRow ${player.isAlive ? '' : 'playerDead'}`} key={player.id}>
                    <Text className='seatNumber'>{String(index + 1).padStart(2, '0')}</Text>
                    <View className='playerBody'>
                      <Text className='playerName'>{player.name || `座位 ${index + 1}`}{player.id === playerId ? '（你）' : ''}</Text>
                      <Text className='playerMeta'>
                        {player.isAlive ? '存活' : `${DEATH_CAUSE_LABELS[death?.cause ?? ''] ?? '死亡'} · 第 ${death?.dayNumber ?? room.dayNumber} 天`}
                        {player.character ? ` · ${characterDisplayName(player.character)}` : ''}
                        {player.poisonedUntil !== undefined ? ` · 中毒至第 ${player.poisonedUntil} 天黄昏` : ''}
                      </Text>
                    </View>
                    {!player.isAlive && <Text className={`tag ${ghostVotes.has(player.id) ? 'tagGood' : ''}`}>{ghostVotes.has(player.id) ? '幽灵票可用' : '幽灵票已用'}</Text>}
                  </View>
                );
              })}
            </View>
          </View>

          {phase === 'day' && (
            <View className='sectionBand'>
              <Text className='sectionHeading'>提名</Text>
              <Text className='sectionDescription'>选择一名存活玩家后提交，服务器会直接进入投票。</Text>
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
          )}

          {phase === 'voting' && currentNomination && (
            <View className='sectionBand'>
              <View className='nominationPanel'>
                <Text className='nominationTitle'>
                  {playerName(room.players, currentNomination.nominatorId)} 提名 {playerName(room.players, currentNomination.nomineeId)}
                </Text>
                <Text className='voteThreshold'>处决需要至少 {executionThreshold} 张赞成票</Text>
              </View>
              <View className='commandRow'>
                <Button className='secondaryButton' disabled={!canVote || busy} onClick={() => castVote(true)}>赞成处决</Button>
                <Button className='dangerButton' disabled={!canVote || busy} onClick={() => castVote(false)}>反对处决</Button>
              </View>
              {self && !self.isAlive && (
                <Text className='sectionDescription'>{ghostVotes.has(playerId) ? '本次投票会消耗你的幽灵票。' : '你的幽灵票已经使用。'}</Text>
              )}
              <View className='playerList'>
                {Object.entries(currentNomination.votes).map(([voterId, decision]) => (
                  <View className='voteRecord' key={voterId}>
                    <Text>{playerName(room.players, voterId)}</Text>
                    <Text className={decision ? 'voteDecisionYes' : 'voteDecisionNo'}>{decision ? '赞成' : '反对'}</Text>
                  </View>
                ))}
              </View>
              {isStoryteller && (
                <Button className='commandButton fullWidthButton' disabled={busy} onClick={resolveNomination}>结算投票</Button>
              )}
            </View>
          )}

          {phase === 'night' && isStoryteller && (
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

              {currentNightStep ? (
                <View className='sectionBand storytellerTools'>
                  <Text className='sectionHeading'>{currentActionLabel}</Text>
                  <Text className='sectionDescription'>{currentActionPrompt}</Text>
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
                    <Text className='fieldLabel'>行动结果</Text>
                    <Textarea
                      className='textAreaInput'
                      maxlength={240}
                      value={nightResult}
                      placeholder='可留空，由服务器计算的信息会自动生成'
                      onInput={(event: InputEvent) => setNightResult(eventValue(event))}
                    />
                  </View>
                  {localError && <Text className='localError'>{localError}</Text>}
                  <Button className='commandButton fullWidthButton' disabled={busy} onClick={submitCurrentNightAction}>提交当前步骤</Button>
                </View>
              ) : (
                <View className='feedback feedbackSuccess inlineFeedback'><Text className='feedbackText'>今夜全部步骤已完成。</Text></View>
              )}

              <Button className='secondaryButton fullWidthButton' disabled={busy || !canResolveNight} onClick={resolveNight}>结束夜晚</Button>
            </View>
          )}

          {phase === 'night' && !isStoryteller && (
            <View className='waitingNight'>
              <Text className='emptyBandTitle'>夜晚降临</Text>
              <Text className='emptyBandText'>请闭眼等待说书人唤醒；这里不会显示其他玩家的行动。</Text>
            </View>
          )}
        </View>

        <View className='gameAside'>
          {visibleCharacter && !isStoryteller && (
            <View className='sectionBand'>
              <Text className='sectionHeading'>你的身份</Text>
              <View className='rolePanel'>
                <Text className='roleName'>{characterDisplayName(visibleCharacter)}</Text>
                <Text className='roleAbility'>{characterDisplayAbility(visibleCharacter)}</Text>
              </View>
            </View>
          )}

          {isStoryteller && phase === 'day' && (
            <View className='sectionBand storytellerTools'>
              <Text className='sectionHeading'>说书人工具</Text>
              <Button className='secondaryButton fullWidthButton' disabled={busy} onClick={() => changePhase('night')}>结束白天</Button>

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

          {canUseSlayer && (
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

          {deadPlayers.length > 0 && (
            <View className='sectionBand'>
              <Text className='sectionHeading'>死亡时间线</Text>
              {room.deaths?.map((death) => (
                <View className='timelineEntry' key={`${death.dayNumber}-${death.playerId}`}>
                  <Text className='timelineTitle'>{playerName(room.players, death.playerId)}</Text>
                  <Text className='timelineMeta'>第 {death.dayNumber} 天 · {DEATH_CAUSE_LABELS[death.cause] ?? death.cause}</Text>
                </View>
              ))}
            </View>
          )}

          {isStoryteller && phase === 'night' && (room.nightActions?.length ?? 0) > 0 && (
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

          {isCreator && (
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

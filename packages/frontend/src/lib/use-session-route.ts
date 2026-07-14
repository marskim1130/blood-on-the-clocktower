import Taro from '@tarojs/taro';
import { useEffect, useRef } from 'react';
import { useRoomSession } from './room-session-store';
import { routeForExperience, type SessionRoute } from './session-routing';

export function useSessionRoute(currentRoute: SessionRoute): void {
  const experience = useRoomSession((state) => state.experience);
  const identity = useRoomSession((state) => state.identity);
  const initialized = useRoomSession((state) => state.initialized);
  const lastRedirect = useRef<SessionRoute | null>(null);

  useEffect(() => {
    const target = routeForExperience(experience, Boolean(identity), initialized);
    if (!target || target === currentRoute || lastRedirect.current === target) return;
    lastRedirect.current = target;
    void Taro.redirectTo({ url: target });
  }, [currentRoute, experience, identity, initialized]);
}


import { expect, it } from 'vitest';
import { createRoomInvite } from './room-invite';

it('H5 邀请使用可直接打开的绝对 hash 链接，小程序保留页面路径', () => {
  expect(createRoomInvite('abc', 'https://clock.example')).toBe('https://clock.example/#/pages/index/index?roomId=abc');
  expect(createRoomInvite('abc')).toBe('/pages/index/index?roomId=abc');
});

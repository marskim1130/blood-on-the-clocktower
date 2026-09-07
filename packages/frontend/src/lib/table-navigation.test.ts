import { expect, it, vi } from 'vitest';
import Taro from '@tarojs/taro';
import { returnToTable } from './table-navigation';

vi.mock('@tarojs/taro', () => ({ default: { getCurrentPages: vi.fn(() => [{}]), navigateBack: vi.fn(async () => ({})), reLaunch: vi.fn(async () => ({})) } }));
it('returns to the table after refreshing a directly opened library page', async () => {
  await returnToTable('/pages/game-setup/index');
  expect(Taro.reLaunch).toHaveBeenCalledWith({ url: '/pages/game-setup/index' });
  expect(Taro.navigateBack).not.toHaveBeenCalled();
});

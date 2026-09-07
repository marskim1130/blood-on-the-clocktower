import Taro from '@tarojs/taro';

export async function returnToTable(destination: string): Promise<void> {
  const pages = Taro.getCurrentPages();
  const target = destination.replace(/^\//, '').split('?')[0];
  for (let index = pages.length - 2; index >= 0; index--) {
    if (pages[index]?.route?.replace(/^\//, '') === target) {
      try { await Taro.navigateBack({ delta: pages.length - 1 - index }); return; } catch { break; }
    }
  }
  await Taro.reLaunch({ url: destination });
}

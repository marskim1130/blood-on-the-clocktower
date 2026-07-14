const websocketUrl = process.env.TARO_APP_WS_URL ?? '';
const development = process.env.NODE_ENV !== 'production';

const config = {
  projectName: 'blood-on-the-clocktower',
  date: '2026-06-09',
  designWidth: 375,
  deviceRatio: {
    375: 2,
    640: 2.34 / 2,
    750: 1,
    828: 1.81 / 2,
  },
  sourceRoot: 'src',
  outputRoot: 'dist',
  plugins: ['@tarojs/plugin-framework-react', '@tarojs/plugin-platform-weapp', '@tarojs/plugin-platform-h5'],
  defineConstants: {
    __CLOCKTOWER_WS_URL__: JSON.stringify(websocketUrl),
    __CLOCKTOWER_DEV__: JSON.stringify(development),
  },
  copy: {
    patterns: [],
    options: {},
  },
  framework: 'react',
  compiler: {
    type: 'webpack5',
    prebundle: {
      enable: false,
    },
  },
  mini: {},
  h5: {
    postcss: {
      pxtransform: {
        enable: true,
        config: {
          targetUnit: 'px',
        },
      },
    },
  },
};

export default config;

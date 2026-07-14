const { getDefaultConfig, mergeConfig } = require('@react-native/metro-config')
const { getMetroConfig } = require('@tarojs/rn-supporter')
const path = require('node:path')

/**
 * Metro configuration
 * https://facebook.github.io/metro/docs/configuration
 *
 * @type {import('metro-config').MetroConfig}
 */
const config = {
  watchFolders: [path.resolve(__dirname, '../..')],
  resolver: {
    unstable_enableSymlinks: true,
  },
}

module.exports = (async function () {
  return mergeConfig(getDefaultConfig(__dirname), await getMetroConfig(), config)
})()

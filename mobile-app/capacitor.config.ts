import { CapacitorConfig } from '@capacitor/cli';

const config: CapacitorConfig = {
  appId: 'com.mobilecoder.app',
  appName: 'MobileCoder',
  webDir: 'dist',
  server: {
    url: 'http://121.41.69.142:3001',
    cleartext: true,
  },
  android: {
    backgroundColor: '#1f2937',
  },
};

export default config;

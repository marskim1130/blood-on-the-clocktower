import { useEffect, type PropsWithChildren } from 'react';
import { getRoomSessionState } from './lib/room-session-store';
import './app.css';

function App({ children }: PropsWithChildren) {
  useEffect(() => {
    getRoomSessionState().initialize();
  }, []);

  return children;
}

export default App;

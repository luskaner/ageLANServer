import { useState } from 'react';
import { FluentProvider, webDarkTheme } from '@fluentui/react-components';
import { AppConfig, DEFAULT_TOML_CONTENT } from './types';
import { parseTOML } from './tomlUtils';
import { Screen1_FileLoad } from './components/Screen1_FileLoad';
import { Screen2_ConfigEditor } from './components/Screen2_ConfigEditor';

export function App() {
  const [activeScreen, setActiveScreen] = useState<'screen1' | 'screen2'>('screen1');
  const [loadedConfig, setLoadedConfig] = useState<AppConfig>(() => parseTOML(DEFAULT_TOML_CONTENT));
  const [fileName, setFileName] = useState<string>('ageLANServer.toml');

  const handleConfigLoaded = (config: AppConfig, name: string) => {
    setLoadedConfig(config);
    setFileName(name);
    setActiveScreen('screen2');
  };

  // Custom transparent background theme override so Wails 3 backdrop effect is visible
  const customFluentTheme = {
    ...webDarkTheme,
    colorNeutralBackground1: 'transparent',
    colorNeutralBackground2: 'rgba(255, 255, 255, 0.035)',
    colorNeutralBackground3: 'rgba(255, 255, 255, 0.06)',
    colorBrandBackground: '#60cdff',
    colorBrandBackgroundHover: '#4cc2ff',
    colorBrandForeground1: '#60cdff',
  };

  return (
    <FluentProvider theme={customFluentTheme} style={{ background: 'transparent', minHeight: '100vh' }}>
      <div className="win-app-root">
        {/* Wails 3 Backdrop Layer */}
        <div className="win-bg-backdrop" />

        <main className="win-main-content">
          {activeScreen === 'screen1' ? (
            <Screen1_FileLoad onConfigLoaded={handleConfigLoaded} />
          ) : (
            <Screen2_ConfigEditor
              initialConfig={loadedConfig}
              fileName={fileName}
              onBack={() => setActiveScreen('screen1')}
            />
          )}
        </main>

        {/* WinUI 3 Footer bar */}
        <footer className="win-footer">
          <div className="win-footer-brand">
            <span className="win-footer-dot" />
            <span>ageLANServer Configurator • Fluent UI React v9</span>
          </div>
          <div className="win-footer-info">
            <span>Screen: {activeScreen === 'screen1' ? '1 / 2 (Load)' : '2 / 2 (Editor)'}</span>
          </div>
        </footer>
      </div>
    </FluentProvider>
  );
}

export default App;

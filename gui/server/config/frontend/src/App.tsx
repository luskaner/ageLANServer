import { useEffect, useRef, useState } from 'react';
import { Events } from '@wailsio/runtime';
import { ConfigFile, ConfigService, Configuration } from '../bindings/changeme';
import { AppConfig, BattleServerConfig } from './types';
import { Screen1_FileLoad } from './components/Screen1_FileLoad';
import { Screen2_ConfigEditor } from './components/Screen2_ConfigEditor';

function toEditorConfig(config: Configuration): AppConfig {
    const game = (hosts: string[] | null, battleServers: BattleServerConfig[] | null) => ({
        Hosts: hosts ?? [],
        BattleServers: battleServers ?? [],
    });
    return {
        Log: config.Log,
        GeneratePlatformUserId: config.GeneratePlatformUserId,
        Sign: { Enabled: config.Sign.Enabled, PlatformUserIdExceptions: config.Sign.PlatformUserIdExceptions ?? [] },
        Authentication: config.Authentication as AppConfig['Authentication'],
        Internet: config.Internet,
        Games: {
            Enabled: config.Games.Enabled ?? [],
            age1: game(config.Games.age1.Hosts, config.Games.age1.BattleServers),
            age2: game(config.Games.age2.Hosts, config.Games.age2.BattleServers),
            age3: game(config.Games.age3.Hosts, config.Games.age3.BattleServers),
            age4: game(config.Games.age4.Hosts, config.Games.age4.BattleServers),
            athens: game(config.Games.athens.Hosts, config.Games.athens.BattleServers),
        },
        Announcement: config.Announcement,
        CustomFiles: {
            ...config.CustomFiles,
            MaxFileSize: toEditorSize(config.CustomFiles.MaxFileSize),
            MaxTotalSize: toEditorSize(config.CustomFiles.MaxTotalSize),
        },
    };
}

/**
 * The server also reads single-letter sizes (32M, 256G): the editor offers MB
 * and GB only, so those spellings are taken as their full unit.
 */
function toEditorSize(size: string): string {
    return size.replace(/^(\d+)\s*M$/i, '$1MB').replace(/^(\d+)\s*G$/i, '$1GB');
}

export function App() {
    const [loadedFile, setLoadedFile] = useState<ConfigFile | null>(null);
    const [errorMsg, setErrorMsg] = useState<string | null>(null);
    // The editor keeps its form state from the moment it mounts, so loading a
    // different document while it is open has to remount it: the key changes on
    // every load, including dropping the same file again after editing it.
    const [loadKey, setLoadKey] = useState(0);
    const hasStartedLoading = useRef(false);

    const showLoadedFile = (file: ConfigFile | null) => {
        if (!file) return;
        setLoadedFile(file);
        setLoadKey((key) => key + 1);
    };

    const openConfigFile = () => {
        setErrorMsg(null);
        ConfigService.OpenConfigFile()
            .then(showLoadedFile)
            .catch((error: Error) => setErrorMsg(error.message));
    };

    const selectConfigFile = () => {
        setErrorMsg(null);
        ConfigService.SelectConfigFile()
            .then(showLoadedFile)
            .catch((error: Error) => setErrorMsg(error.message));
    };

    const createConfigFile = () => {
        setErrorMsg(null);
        ConfigService.OpenTemplate()
            .then(showLoadedFile)
            .catch((error: Error) => setErrorMsg(error.message));
    };

    // Try the default configuration on startup. When there is none the load
    // screen stays in place: the picker only opens when the user asks for it.
    useEffect(() => {
        if (hasStartedLoading.current) return;
        hasStartedLoading.current = true;
        openConfigFile();
    }, []);

    useEffect(() => {
        return Events.On('files-dropped', (event) => {
            const data = event.data as { files?: string[] } | undefined;
            const droppedPath = data?.files?.[0];
            if (!droppedPath) {
                setErrorMsg('No file path was received from the dropped item.');
                return;
            }
            setErrorMsg(null);
            ConfigService.OpenConfigFileAtPath(droppedPath)
                .then(showLoadedFile)
                .catch((error: Error) => setErrorMsg(error.message));
        });
    }, []);

    return loadedFile ? (
        <Screen2_ConfigEditor
            key={loadKey}
            initialConfig={toEditorConfig(loadedFile.config!)}
            backendValidationErrors={loadedFile.validationErrors ?? []}
            fileName={loadedFile.path}
            isNew={!loadedFile.path}
            onSaved={showLoadedFile}
            onBack={() => setLoadedFile(null)}
        />
    ) : (
        <Screen1_FileLoad
            onSelectConfig={selectConfigFile}
            onCreateConfig={createConfigFile}
            errorMsg={errorMsg || undefined}
            validationErrors={[]}
        />
    );
}

export default App;

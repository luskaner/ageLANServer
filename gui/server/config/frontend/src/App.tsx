import { useEffect, useRef, useState } from 'react';
import { Events } from '@wailsio/runtime';
import { ConfigFile, ConfigService, Configuration } from '../bindings/changeme';
import { AppConfig } from './types';
import { Screen1_FileLoad } from './components/Screen1_FileLoad';
import { Screen2_ConfigEditor } from './components/Screen2_ConfigEditor';

function toEditorConfig(config: Configuration): AppConfig {
    const game = (hosts: string[] | null) => ({ Hosts: hosts ?? [] });
    return {
        Log: config.Log,
        GeneratePlatformUserId: config.GeneratePlatformUserId,
        Sign: config.Sign,
        Authentication: config.Authentication as AppConfig['Authentication'],
        Internet: config.Internet,
        Games: {
            Enabled: config.Games.Enabled ?? [],
            age1: game(config.Games.age1.Hosts),
            age2: game(config.Games.age2.Hosts),
            age3: game(config.Games.age3.Hosts),
            age4: game(config.Games.age4.Hosts),
            athens: game(config.Games.athens.Hosts),
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
    const hasStartedLoading = useRef(false);

    const openConfigFile = () => {
        setErrorMsg(null);
        ConfigService.OpenConfigFile()
            .then((file) => {
                if (!file) return;
                setLoadedFile(file);
            })
            .catch((error: Error) => setErrorMsg(error.message));
    };

    const selectConfigFile = () => {
        setErrorMsg(null);
        ConfigService.SelectConfigFile()
            .then((file) => {
                if (!file) return;
                setLoadedFile(file);
            })
            .catch((error: Error) => setErrorMsg(error.message));
    };

    const createConfigFile = () => {
        setErrorMsg(null);
        ConfigService.OpenTemplate()
            .then((file) => {
                if (!file) return;
                setLoadedFile(file);
            })
            .catch((error: Error) => setErrorMsg(error.message));
    };

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
                .then((file) => {
                    if (!file) return;
                    setLoadedFile(file);
                })
                .catch((error: Error) => setErrorMsg(error.message));
        });
    }, []);

    return loadedFile ? (
        <Screen2_ConfigEditor
            initialConfig={toEditorConfig(loadedFile.config!)}
            backendValidationErrors={loadedFile.validationErrors ?? []}
            fileName={loadedFile.path}
            isNew={!loadedFile.path}
            onSaved={setLoadedFile}
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

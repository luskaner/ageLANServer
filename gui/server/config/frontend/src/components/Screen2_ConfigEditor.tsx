import React, { useState, useEffect } from 'react';
import {
    Card,
    Button,
    Switch,
    Dropdown,
    Option,
    SpinButton,
    Field,
    Badge,
    MessageBar,
    MessageBarBody,
    MessageBarTitle,
    Dialog,
    DialogSurface,
    DialogBody,
    DialogTitle,
    DialogContent,
    DialogActions,
} from '@fluentui/react-components';
import {
    ArrowLeft24Regular,
    Save24Regular,
    Code24Regular,
    Checkmark24Filled,
    DocumentText24Regular,
    Shield24Regular,
    Games24Regular,
    Megaphone24Regular,
    Globe24Regular,
} from '@fluentui/react-icons';
import { AppConfig, ValidationErrors, AVAILABLE_GAMES } from '../types';
import { validateIPv4, validateMulticastIPv4 } from '../validation';
import { IPv4Input } from './WinUI/IPv4Input';
import { ConfigFile, ConfigService, Configuration, ValidationError } from '../../bindings/changeme';

function toServiceConfiguration(config: AppConfig): Configuration {
    const game = (gameId: string) => ({ Hosts: config.Games[gameId]?.Hosts ?? [] });
    return {
        Log: config.Log,
        GeneratePlatformUserId: config.GeneratePlatformUserId,
        Authentication: config.Authentication,
        Games: {
            Enabled: config.Games.Enabled,
            age1: game('age1'),
            age2: game('age2'),
            age3: game('age3'),
            age4: game('age4'),
            athens: game('athens'),
        },
        Announcement: config.Announcement,
    };
}

interface Screen2Props {
    initialConfig: AppConfig;
    originalContent: string;
    backendValidationErrors: ValidationError[];
    fileName: string;
    isNew: boolean;
    onSaved: (file: ConfigFile) => void;
    onBack: () => void;
}

export const Screen2_ConfigEditor: React.FC<Screen2Props> = ({
    initialConfig,
    originalContent,
    backendValidationErrors,
    fileName,
    isNew,
    onSaved,
    onBack,
}) => {
    const [config, setConfig] = useState<AppConfig>(initialConfig);
    const [errors, setErrors] = useState<ValidationErrors>({});
    const [backendErrors, setBackendErrors] = useState(backendValidationErrors);
    const [toastMessage, setToastMessage] = useState<string | null>(null);
    const [showTomlModal, setShowTomlModal] = useState(false);
    const [previewContent, setPreviewContent] = useState('');
    const [isPreviewPending, setIsPreviewPending] = useState(true);
    const [previewError, setPreviewError] = useState<string | null>(null);

    useEffect(() => {
        let isCurrentPreview = true;
        setIsPreviewPending(true);
        ConfigService.PreviewConfiguration(originalContent, toServiceConfiguration(config))
            .then((content) => {
                if (!isCurrentPreview) return;
                setPreviewContent(content);
                setPreviewError(null);
                setIsPreviewPending(false);
            })
            .catch((error: Error) => {
                if (!isCurrentPreview) return;
                setPreviewError(error.message);
                setIsPreviewPending(false);
            });

        return () => {
            isCurrentPreview = false;
        };
    }, [config, originalContent]);

    // TODO: react validations need their own files
    // Validate fields whenever config changes
    useEffect(() => {
        const newErrors: ValidationErrors = {};

        if (config.GeneratePlatformUserId && config.Authentication !== 'disabled') {
            newErrors.GeneratePlatformUserId = 'Authentication must be disabled when generating platform user IDs.';
        }

        if (config.Games.Enabled.length === 0) {
            newErrors['Games.Enabled'] = 'At least one game must be enabled.';
        }
        config.Games.Enabled.forEach((gameId, index) => {
            if (!AVAILABLE_GAMES.some((game) => game.id === gameId)) {
                newErrors[`Games.Enabled.${index}`] = `Unsupported game: ${gameId}`;
            }
        });

        // Validate Announcement Port
        if (isNaN(config.Announcement.Port) || config.Announcement.Port < 1 || config.Announcement.Port > 65535) {
            newErrors['Announcement.Port'] = 'Port must be an integer between 1 and 65535.';
        }

        // Validate Announcement MulticastGroup (Class D Multicast 224.0.0.0 - 239.255.255.255)
        if (config.Announcement.Enabled && config.Announcement.Multicast &&
            !validateMulticastIPv4(config.Announcement.MulticastGroup)) {
            newErrors['Announcement.MulticastGroup'] = 'Valid Multicast IPv4 address required (Range: 224.0.0.0 to 239.255.255.255, e.g., 239.31.97.8)';
        }

        // Validate Game Hosts
        AVAILABLE_GAMES.forEach((game) => {
            if (!config.Games.Enabled.includes(game.id)) return;

            const gameConf = config.Games[game.id];
            const hosts = gameConf && Array.isArray(gameConf.Hosts) ? gameConf.Hosts : [];
            hosts.forEach((host: string, idx: number) => {
                if (!validateIPv4(host)) {
                    newErrors[`Games.${game.id}.Hosts.${idx}`] = `Invalid IP for ${game.id} (e.g., 0.0.0.0)`;
                }
            });
        });

        const usedHosts = new Set<string>();
        config.Games.Enabled.forEach((gameId) => {
            const hosts = config.Games[gameId]?.Hosts ?? [];
            hosts.forEach((host, index) => {
                const normalizedHost = host.trim();
                if (usedHosts.has(normalizedHost)) {
                    newErrors[`Games.${gameId}.Hosts.${index}`] = 'A host cannot be shared by enabled games.';
                }
                usedHosts.add(normalizedHost);
            });
        });

        setErrors(newErrors);
        setBackendErrors((previous) => previous.filter((error) => newErrors[error.field]));
    }, [config]);

    const backendErrorMap: ValidationErrors = Object.fromEntries(
        backendErrors.map((error) => [error.field, error.message]),
    );
    const allErrors = { ...backendErrorMap, ...errors };
    const hasErrors = Object.keys(allErrors).length > 0;

    // State handlers
    const handleToggleLog = (checked: boolean) => setConfig({ ...config, Log: checked });
    const handleToggleGenUserId = (checked: boolean) => setConfig({ ...config, GeneratePlatformUserId: checked });
    const handleAuthChange = (val: string) => setConfig({ ...config, Authentication: val as any });

    const handleGameToggle = (gameId: string) => {
        const currentEnabled = [...config.Games.Enabled];
        const index = currentEnabled.indexOf(gameId);
        if (index === -1) {
            currentEnabled.push(gameId);
        } else {
            currentEnabled.splice(index, 1);
        }
        setConfig({
            ...config,
            Games: {
                ...config.Games,
                Enabled: currentEnabled,
            },
        });
    };

    const handleGameHostChange = (gameId: string, index: number, value: string) => {
        const hosts = [...(config.Games[gameId]?.Hosts ?? [])];
        hosts[index] = value;
        setConfig({
            ...config,
            Games: {
                ...config.Games,
                [gameId]: {
                    Hosts: hosts,
                },
            },
        });
    };

    const handleAnnouncementToggle = (checked: boolean) => {
        setConfig({
            ...config,
            Announcement: { ...config.Announcement, Enabled: checked },
        });
    };

    const handleMulticastToggle = (checked: boolean) => {
        setConfig({
            ...config,
            Announcement: { ...config.Announcement, Multicast: checked },
        });
    };

    const handlePortChange = (val: number | string | null | undefined) => {
        if (val == null || val === '') return;
        const port = typeof val === 'string' ? Number.parseInt(val, 10) : val;
        if (Number.isNaN(port)) return;
        setConfig({
            ...config,
            Announcement: { ...config.Announcement, Port: port },
        });
    };

    const handleMulticastGroupChange = (val: string) => {
        setConfig({
            ...config,
            Announcement: { ...config.Announcement, MulticastGroup: val },
        });
    };

    const triggerToast = (msg: string) => {
        setToastMessage(msg);
        setTimeout(() => {
            setToastMessage(null);
        }, 4000);
    };

    const handleSave = async () => {
        if (hasErrors) return;
        try {
            if (!isNew) {
                await ConfigService.SaveConfiguration(fileName, toServiceConfiguration(config));
                triggerToast('Configuration saved successfully.');
                return;
            }

            const destination = await ConfigService.SelectConfigDestination();
            if (!destination) return;
            try {
                const created = await ConfigService.CreateConfigFile(destination, toServiceConfiguration(config), false);
                if (!created) throw new Error('The created configuration was not returned.');
                onSaved(created);
                triggerToast('Configuration created successfully.');
            } catch (error) {
                const message = error instanceof Error ? error.message : String(error);
                if (!message.includes('confirmation required') || !window.confirm('The selected file already exists. Overwrite it?')) {
                    throw error;
                }
                const created = await ConfigService.CreateConfigFile(destination, toServiceConfiguration(config), true);
                if (!created) throw new Error('The overwritten configuration was not returned.');
                onSaved(created);
                triggerToast('Configuration overwritten successfully.');
            }
        } catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            triggerToast(`Error saving configuration: ${message}`);
        }
    };

    const handleSaveAs = async () => {
        if (hasErrors) return;
        try {
            const destination = await ConfigService.SelectConfigSaveAsDestination(fileName);
            if (!destination) return;

            try {
                const saved = await ConfigService.SaveConfigurationAs(
                    fileName,
                    destination,
                    toServiceConfiguration(config),
                    false,
                );
                if (!saved) throw new Error('The saved configuration was not returned.');
                onSaved(saved);
                triggerToast('Configuration saved successfully to the new location.');
            } catch (error) {
                const message = error instanceof Error ? error.message : String(error);
                if (!message.includes('confirmation required') || !window.confirm('The selected file already exists. Overwrite it?')) {
                    throw error;
                }
                const saved = await ConfigService.SaveConfigurationAs(
                    fileName,
                    destination,
                    toServiceConfiguration(config),
                    true,
                );
                if (!saved) throw new Error('The overwritten configuration was not returned.');
                onSaved(saved);
                triggerToast('Configuration overwritten successfully.');
            }
        } catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            triggerToast(`Error saving configuration: ${message}`);
        }
    };

    const handleCopyToClipboard = async () => {
        try {
            await navigator.clipboard.writeText(previewContent);
            triggerToast('TOML content copied to clipboard.');
        } catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            triggerToast(`Error copying TOML content: ${message}`);
        }
    };

    const authOptions = [
        { value: 'disabled', label: 'No authentication (disabled)' },
        { value: 'required', label: 'Required (required)' },
        { value: 'cached', label: 'Cached (cached)' },
        { value: 'adaptive', label: 'Adaptive (adaptive)' },
    ];

    return (
        <div className="win-screen-container">
            {/* Top Navbar */}
            <div className="win-top-bar">
                <Button
                    appearance="subtle"
                    icon={<ArrowLeft24Regular />}
                    onClick={onBack}
                >
                    Back
                </Button>

                <div className="win-top-title-group">
                    <h2>Configuration Editor</h2>
                    <Badge appearance="tint" color="brand">{fileName || 'New configuration'}</Badge>
                </div>
            </div>

            <div className="win-editor-sticky-bar">
                <div className="win-top-actions">
                    <Button
                        appearance="secondary"
                        icon={<Code24Regular />}
                        onClick={() => setShowTomlModal(true)}
                    >
                        View TOML
                    </Button>

                    {!isNew && (
                        <Button
                            appearance="secondary"
                            icon={<Save24Regular />}
                            onClick={handleSaveAs}
                            disabled={hasErrors}
                        >
                            Save As
                        </Button>
                    )}

                    <Button
                        appearance="primary"
                        icon={<Save24Regular />}
                        onClick={handleSave}
                        disabled={hasErrors}
                    >
                        Save Configuration
                    </Button>
                </div>

                {hasErrors && (
                    <MessageBar className="win-validation-bar" intent="warning">
                        <MessageBarBody>
                            <MessageBarTitle>Validation Errors</MessageBarTitle>
                            There are fields with errors. Please correct them to enable saving.
                            {Object.keys(allErrors).length > 0 && (
                                <ul className="win-validation-errors">
                                    {Object.entries(allErrors).map(([field, message]) => (
                                        <li key={`${field}-${message}`}>
                                            <strong>{field}</strong>: {message}
                                        </li>
                                    ))}
                                </ul>
                            )}
                        </MessageBarBody>
                    </MessageBar>
                )}
            </div>

            {/* SECTION 1: GENERAL */}
            <div className="win-card-group">
                <h3 className="win-section-header">General</h3>
                <div className="win-card-stack">
                    {/* Card 1: Log */}
                    <Card className="win-card">
                        <div className="win-setting-row-main win-game-selection-row">
                            <div className="win-setting-icon"><DocumentText24Regular /></div>
                            <div className="win-setting-text">
                                <div className="win-setting-title">Log Information</div>
                                <div className="win-setting-subtitle">
                                    Enables logging to terminal and saves data to a file for debugging.
                                </div>
                            </div>
                            <div className="win-setting-control">
                                <Switch
                                    checked={config.Log}
                                    onChange={(_, data) => handleToggleLog(data.checked)}
                                    label={config.Log ? 'Enabled' : 'Disabled'}
                                    labelPosition="before"
                                />
                            </div>
                        </div>
                    </Card>

                    {/* Card 2: GeneratePlatformUserId */}
                    <Card className="win-card">
                        <div className="win-setting-row-main">
                            <div className="win-setting-icon"><Shield24Regular /></div>
                            <div className="win-setting-text">
                                <div className="win-setting-title">Generate Platform User ID</div>
                                <div className="win-setting-subtitle">
                                    Generate a unique ID. *ONLY* if multiple users share a launcher. Incompatible with active authentication.
                                </div>
                            </div>
                            <div className="win-setting-control">
                                <Switch
                                    checked={config.GeneratePlatformUserId}
                                    onChange={(_, data) => handleToggleGenUserId(data.checked)}
                                    label={config.GeneratePlatformUserId ? 'Enabled' : 'Disabled'}
                                    labelPosition="before"
                                />
                            </div>
                        </div>
                    </Card>

                    {/* Card 3: Authentication */}
                    <Card className="win-card">
                        <div className="win-setting-row-main">
                            <div className="win-setting-icon"><Shield24Regular /></div>
                            <div className="win-setting-text">
                                <div className="win-setting-title">Authentication Method</div>
                                <div className="win-setting-subtitle">
                                    Define how users are authenticated: required, cached, adaptive, or disabled.
                                </div>
                            </div>
                            <div className="win-setting-control">
                                <Dropdown
                                    value={authOptions.find(o => o.value === config.Authentication)?.label || 'No authentication (disabled)'}
                                    onOptionSelect={(_, data) => handleAuthChange(data.optionValue as string)}
                                    style={{ minWidth: 200, backgroundColor: 'var(--colorNeutralBackground1)', padding: '4px', borderRadius: '4px' }}
                                >
                                    {authOptions.map((opt) => (
                                        <Option key={opt.value} value={opt.value}>
                                            {opt.label}
                                        </Option>
                                    ))}
                                </Dropdown>
                            </div>
                        </div>
                    </Card>
                </div>
            </div>

            {/* SECTION 2: GAMES & HOSTS */}
            <div className="win-card-group">
                <h3 className="win-section-header">Supported Games and Binding Hosts</h3>
                <div className="win-card-stack">
                    {/* Card 4: Games and Binding Hosts */}
                    <Card className="win-card">
                        <div className="win-setting-row-main">
                            <div className="win-setting-icon"><Games24Regular /></div>
                            <div className="win-setting-text">
                                <div className="win-setting-title">Games and Network Addresses</div>
                                <div className="win-setting-subtitle">
                                    Select the games the LAN server will accept and configure each binding IP (single IPv4). Use 0.0.0.0 to bind all interfaces; hosts must be unique across enabled games.
                                </div>
                            </div>
                        </div>
                        <div className="win-setting-row-details">
                            <div className="win-hosts-grid">
                                {AVAILABLE_GAMES.map((game) => {
                                    const isEnabled = config.Games.Enabled.includes(game.id);
                                    const gameConf = config.Games[game.id];
                                    const gameHosts = gameConf && Array.isArray(gameConf.Hosts) ? gameConf.Hosts : ['0.0.0.0'];
                                    const hosts = gameHosts.length > 0 ? gameHosts : ['0.0.0.0'];

                                    return (
                                        <div key={game.id} className="win-host-card">
                                            <div className="win-host-info">
                                                <span className="win-host-name">{game.name}</span>
                                                <Switch
                                                    checked={isEnabled}
                                                    onChange={() => handleGameToggle(game.id)}
                                                    aria-label={`Enable ${game.name}`}
                                                />
                                            </div>
                                            {hosts.map((host, index) => {
                                                return (
                                                    <Field className="win-host-field" key={index}>
                                                        <IPv4Input
                                                            error={allErrors[`Games.${game.id}.Hosts.${index}`]}
                                                            value={host}
                                                            onChange={(value) => handleGameHostChange(game.id, index, value)}
                                                            placeholder="0.0.0.0"
                                                            disabled={!isEnabled}
                                                        />
                                                    </Field>
                                                );
                                            })}
                                        </div>
                                    );
                                })}
                            </div>
                        </div>
                    </Card>
                </div>
            </div>

            {/* SECTION 3: ANNOUNCEMENT */}
            <div className="win-card-group">
                <h3 className="win-section-header">Announcement and Network Discovery</h3>
                <div className="win-card-stack">
                    {/* Card 6: Announcement Enabled */}
                    <Card className="win-card">
                        <div className="win-setting-row-main">
                            <div className="win-setting-icon"><Megaphone24Regular /></div>
                            <div className="win-setting-text">
                                <div className="win-setting-title">LAN Announcement</div>
                                <div className="win-setting-subtitle">
                                    Respond to automatic discovery queries on the local LAN.
                                </div>
                            </div>
                            <div className="win-setting-control">
                                <Switch
                                    checked={config.Announcement.Enabled}
                                    onChange={(_, data) => handleAnnouncementToggle(data.checked)}
                                    label={config.Announcement.Enabled ? 'Enabled' : 'Disabled'}
                                    labelPosition="before"
                                />
                            </div>
                        </div>
                    </Card>

                    {/* Card 7: Multicast */}
                    <Card className="win-card">
                        <div className="win-setting-row-main">
                            <div className="win-setting-icon"><Globe24Regular /></div>
                            <div className="win-setting-text">
                                <div className="win-setting-title">Multicast</div>
                                <div className="win-setting-subtitle">
                                    Respond to discovery queries at the multicast address.
                                </div>
                            </div>
                            <div className="win-setting-control">
                                <Switch
                                    checked={config.Announcement.Multicast}
                                    onChange={(_, data) => handleMulticastToggle(data.checked)}
                                    label={config.Announcement.Multicast ? 'Enabled' : 'Disabled'}
                                    labelPosition="before"
                                />
                            </div>
                        </div>
                    </Card>

                    {/* Card 8: Port */}
                    <Card className="win-card">
                        <div className="win-setting-row-main">
                            <div className="win-setting-icon"><Globe24Regular /></div>
                            <div className="win-setting-text">
                                <div className="win-setting-title">Announcement Port</div>
                                <div className="win-setting-subtitle">
                                    UDP port for announcing the server to launchers (Default: 31978).
                                </div>
                            </div>
                            <div className="win-setting-control">
                                <Field
                                    validationMessage={allErrors['Announcement.Port']}
                                    validationState={allErrors['Announcement.Port'] ? 'error' : 'none'}
                                >
                                    <SpinButton
                                        min={1}
                                        max={65535}
                                        value={config.Announcement.Port}
                                        onChange={(_, data) => handlePortChange(data.value ?? data.displayValue)}
                                        aria-valuemin={1}
                                        aria-valuemax={65535}
                                        style={{ minWidth: 180 }}
                                    />
                                </Field>
                            </div>
                        </div>
                    </Card>

                    {/* Card 9: MulticastGroup */}
                    <Card className="win-card">
                        <div className="win-setting-row-main">
                            <div className="win-setting-icon"><Globe24Regular /></div>
                            <div className="win-setting-text">
                                <div className="win-setting-title">Multicast Group</div>
                                <div className="win-setting-subtitle">
                                    IPv4 Multicast group IP for responding to announcements (Default: 239.31.97.8, Class D range: 224.0.0.0 - 239.255.255.255).
                                </div>
                            </div>
                            <div className="win-setting-control">
                                <IPv4Input
                                    error={allErrors['Announcement.MulticastGroup']}
                                    max="239.255.255.255"
                                    min="224.0.0.0"
                                    onChange={handleMulticastGroupChange}
                                    placeholder="239.31.97.8"
                                    value={config.Announcement.MulticastGroup}
                                />
                            </div>
                        </div>
                    </Card>
                </div>
            </div>

            {/* Toast Notification */}
            {toastMessage && (
                <div className="win-toast">
                    <Checkmark24Filled style={{ color: '#60cdff' }} />
                    <span>{toastMessage}</span>
                </div>
            )}

            {/* Fluent UI Dialog Modal */}
            <Dialog open={showTomlModal} onOpenChange={(_, data) => setShowTomlModal(data.open)}>
                <DialogSurface>
                    <DialogBody>
                        <DialogTitle>TOML preview</DialogTitle>
                        <DialogContent>
                            {previewError && (
                                <MessageBar intent="error">
                                    <MessageBarBody>
                                        <MessageBarTitle>Preview unavailable</MessageBarTitle>
                                        {previewError}
                                    </MessageBarBody>
                                </MessageBar>
                            )}
                            {isPreviewPending && <p role="status">Updating preview…</p>}
                            <pre className="win-toml-code">{previewContent}</pre>
                        </DialogContent>
                        <DialogActions>
                            <Button appearance="secondary" onClick={handleCopyToClipboard} disabled={isPreviewPending || !!previewError}>
                                Copy to clipboard
                            </Button>
                            <Button appearance="primary" onClick={() => setShowTomlModal(false)}>
                                Close
                            </Button>
                        </DialogActions>
                    </DialogBody>
                </DialogSurface>
            </Dialog>
        </div>
    );
};

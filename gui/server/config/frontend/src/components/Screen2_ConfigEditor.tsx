import React, { useState, useEffect } from 'react';
import {
  Card,
  Button,
  Switch,
  Dropdown,
  Option,
  Input,
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
  ChevronDown24Regular,
  ChevronUp24Regular,
} from '@fluentui/react-icons';
import { AppConfig, ValidationErrors, AVAILABLE_GAMES } from '../types';
import { stringifyTOML, validateIPv4, validateMulticastIPv4 } from '../tomlUtils';

interface Screen2Props {
  initialConfig: AppConfig;
  fileName: string;
  onBack: () => void;
}

export const Screen2_ConfigEditor: React.FC<Screen2Props> = ({
  initialConfig,
  fileName,
  onBack,
}) => {
  const [config, setConfig] = useState<AppConfig>(initialConfig);
  const [errors, setErrors] = useState<ValidationErrors>({});
  const [toastMessage, setToastMessage] = useState<string | null>(null);
  const [showTomlModal, setShowTomlModal] = useState(false);
  const [hostsExpanded, setHostsExpanded] = useState(true);

  // Validate fields whenever config changes
  useEffect(() => {
    const newErrors: ValidationErrors = {};

    // Validate Announcement Port
    if (isNaN(config.Announcement.Port) || config.Announcement.Port < 1 || config.Announcement.Port > 65535) {
      newErrors['Announcement.Port'] = 'Port must be an integer between 1 and 65535.';
    }

    // Validate Announcement MulticastGroup (Class D Multicast 224.0.0.0 - 239.255.255.255)
    if (!validateMulticastIPv4(config.Announcement.MulticastGroup)) {
      newErrors['Announcement.MulticastGroup'] = 'Valid Multicast IPv4 address required (Range: 224.0.0.0 to 239.255.255.255, e.g., 239.31.97.8)';
    }

    // Validate Game Hosts
    AVAILABLE_GAMES.forEach((game) => {
      const gameConf = config.Games[game.id];
      const hosts = gameConf && Array.isArray(gameConf.Hosts) ? gameConf.Hosts : [];
      hosts.forEach((host: string, idx: number) => {
        if (!validateIPv4(host)) {
          newErrors[`Games.${game.id}.Hosts.${idx}`] = `Invalid IP for ${game.id} (e.g., 0.0.0.0)`;
        }
      });
    });

    setErrors(newErrors);
  }, [config]);

  const hasErrors = Object.keys(errors).length > 0;

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

  const handleGameHostChange = (gameId: string, value: string) => {
    const hostList = value.split(',').map((s) => s.trim());
    setConfig({
      ...config,
      Games: {
        ...config.Games,
        [gameId]: {
          Hosts: hostList,
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

  const handlePortChange = (val: number | null) => {
    setConfig({
      ...config,
      Announcement: { ...config.Announcement, Port: val ?? 0 },
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

  const handleSave = () => {
    if (hasErrors) return;
    const tomlStr = stringifyTOML(config);
    const blob = new Blob([tomlStr], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = fileName || 'ageLANServer.toml';
    link.click();
    URL.revokeObjectURL(url);
    triggerToast('Configuration saved and downloaded successfully!');
  };

  const handleCopyToClipboard = () => {
    const tomlStr = stringifyTOML(config);
    navigator.clipboard.writeText(tomlStr);
    triggerToast('TOML content copied to clipboard.');
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
          <Badge appearance="tint" color="brand">{fileName}</Badge>
        </div>

        <div className="win-top-actions">
          <Button
            appearance="secondary"
            icon={<Code24Regular />}
            onClick={() => setShowTomlModal(true)}
          >
            View TOML
          </Button>

          <Button
            appearance="primary"
            icon={<Save24Regular />}
            onClick={handleSave}
            disabled={hasErrors}
          >
            Save Configuration
          </Button>
        </div>
      </div>

      {hasErrors && (
        <MessageBar intent="warning" style={{ marginBottom: 24 }}>
          <MessageBarBody>
            <MessageBarTitle>Validation Errors</MessageBarTitle>
            There are fields with errors. Please correct them to enable saving.
          </MessageBarBody>
        </MessageBar>
      )}

      {/* SECTION 1: GENERAL */}
      <div className="win-card-group">
        <h3 className="win-section-header">General</h3>
        <div className="win-card-stack">
          {/* Card 1: Log */}
          <Card className="win-card">
            <div className="win-setting-row-main">
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
          {/* Card 4: Games Enabled */}
          <Card className="win-card">
            <div className="win-setting-row-main">
              <div className="win-setting-icon"><Games24Regular /></div>
              <div className="win-setting-text">
                <div className="win-setting-title">Enabled Games (Games.Enabled)</div>
                <div className="win-setting-subtitle">
                  Select which Age of Empires games the LAN server will accept.
                </div>
              </div>
              <div className="win-setting-control">
                <div className="win-pills-group">
                  {AVAILABLE_GAMES.map((game) => {
                    const isEnabled = config.Games.Enabled.includes(game.id);
                    return (
                      <Button
                        key={game.id}
                        appearance={isEnabled ? 'primary' : 'secondary'}
                        size="small"
                        shape="rounded"
                        onClick={() => handleGameToggle(game.id)}
                        style={{ backgroundColor: 'var(--colorNeutralBackground1)', color: 'var(--colorNeutralForeground1)', margin: '2px' }}
                      >
                        {isEnabled ? '✓ ' : '+ '}
                        {game.name}
                      </Button>
                    );
                  })}
                </div>
              </div>
            </div>
          </Card>

          {/* Card 5: Hosts Per Game */}
          <Card className="win-card">
            <div className="win-setting-row-main" onClick={() => setHostsExpanded(!hostsExpanded)} style={{ cursor: 'pointer' }}>
              <div className="win-setting-icon"><Globe24Regular /></div>
              <div className="win-setting-text">
                <div className="win-setting-title">Network Addresses per Game (Games.&lt;game&gt;.Hosts)</div>
                <div className="win-setting-subtitle">
                  Configure the IP to which the server will bind for each game (single IPv4).
                </div>
              </div>
              <div className="win-setting-control">
                <Button appearance="subtle" icon={hostsExpanded ? <ChevronUp24Regular /> : <ChevronDown24Regular />} />
              </div>
            </div>

            {hostsExpanded && (
              <div className="win-setting-row-details">
                <div className="win-hosts-grid">
                  {AVAILABLE_GAMES.map((game) => {
                    const gameConf = config.Games[game.id];
                    const gameHosts = gameConf && Array.isArray(gameConf.Hosts) ? gameConf.Hosts : ['0.0.0.0'];
                    const hostStr = gameHosts.join(', ');
                    const hostError = errors[`Games.${game.id}.Hosts.0`];

                    return (
                      <div key={game.id} className="win-host-card">
                        <div className="win-host-info">
                          <span className="win-host-name">{game.name}</span>
                          <Badge appearance="outline" size="extra-small">[{game.id}]</Badge>
                        </div>
                        <Field validationMessage={hostError} validationState={hostError ? 'error' : 'none'}>
                          <Input
                            value={hostStr}
                            onChange={(_, data) => handleGameHostChange(game.id, data.value)}
                            placeholder="0.0.0.0"
                          />
                        </Field>
                      </div>
                    );
                  })}
                </div>
              </div>
            )}
          </Card>
        </div>
      </div>

      {/* SECTION 3: ANNOUNCEMENT */}
      <div className="win-card-group">
        <h3 className="win-section-header">Announcement and Network Discovery (Announcement)</h3>
        <div className="win-card-stack">
          {/* Card 6: Announcement Enabled */}
          <Card className="win-card">
            <div className="win-setting-row-main">
              <div className="win-setting-icon"><Megaphone24Regular /></div>
              <div className="win-setting-text">
                <div className="win-setting-title">LAN Announcement (Enabled)</div>
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
                  validationMessage={errors['Announcement.Port']}
                  validationState={errors['Announcement.Port'] ? 'error' : 'none'}
                >
                  <SpinButton
                    min={1}
                    max={65535}
                    value={config.Announcement.Port}
                    onChange={(_, data) => handlePortChange(data.value ?? 0)}
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
                <Field
                  validationMessage={errors['Announcement.MulticastGroup']}
                  validationState={errors['Announcement.MulticastGroup'] ? 'error' : 'none'}
                >
                  <Input
                    value={config.Announcement.MulticastGroup}
                    onChange={(_, data) => handleMulticastGroupChange(data.value)}
                    placeholder="239.31.97.8"
                    style={{ minWidth: 180 }}
                  />
                </Field>
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
            <DialogTitle>Preview of Generated TOML</DialogTitle>
            <DialogContent>
              <pre className="win-toml-code">{stringifyTOML(config)}</pre>
            </DialogContent>
            <DialogActions>
              <Button appearance="secondary" onClick={handleCopyToClipboard}>
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

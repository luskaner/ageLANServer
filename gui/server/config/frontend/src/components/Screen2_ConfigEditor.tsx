import React, { useEffect, useRef, useState } from 'react';
import {
    ActionIcon,
    Alert,
    AppShell,
    Box,
    Burger,
    Button,
    Group,
    Input,
    List,
    Menu,
    NavLink,
    NumberInput,
    Radio,
    ScrollArea,
    SimpleGrid,
    Stack,
    Text,
    Title,
    Tooltip,
    UnstyledButton,
} from '@mantine/core';
import { useForm } from '@mantine/form';
import { useDebouncedValue, useDisclosure, useMediaQuery } from '@mantine/hooks';
import { notifications } from '@mantine/notifications';
import {
    IconAlertTriangle,
    IconBroadcast,
    IconChevronDown,
    IconCircleCheck,
    IconClock,
    IconDatabase,
    IconDeviceFloppy,
    IconDeviceGamepad2,
    IconFile,
    IconFiles,
    IconFolder,
    IconFolderOpen,
    IconId,
    IconLogs,
    IconNetwork,
    IconPlus,
    IconSettings,
    IconShield,
    IconSignature,
    IconSpeakerphone,
    IconTrash,
    IconWorld,
} from '@tabler/icons-react';
import { AppConfig, AVAILABLE_GAMES, ValidationErrors, fieldLabel, humanMessage } from '../types';
import { AutoManualCard } from './AutoManualCard';
import { FolderPickerInput } from './FolderPickerInput';
import { IPv4Input } from './IPv4Input';
import { SettingCard } from './SettingCard';
import { SwitchSettingCard } from './SwitchSettingCard';
import { UnitInput, UnitInputUnit } from './UnitInput';
import { ConfigFile, ConfigService, Configuration, ValidationError } from '../../bindings/changeme';
function toServiceConfiguration(config: AppConfig): Configuration {
    const game = (gameId: string) => ({ Hosts: config.Games[gameId]?.Hosts ?? [] });
    return {
        Log: config.Log,
        GeneratePlatformUserId: config.GeneratePlatformUserId,
        Sign: config.Sign,
        Authentication: config.Authentication,
        Internet: { Enabled: config.Internet.Enabled, IP: config.Internet.IP },
        Games: {
            Enabled: config.Games.Enabled,
            age1: game('age1'),
            age2: game('age2'),
            age3: game('age3'),
            age4: game('age4'),
            athens: game('athens'),
        },
        Announcement: config.Announcement,
        CustomFiles: config.CustomFiles,
    };
}
type SectionKey = 'general' | 'internet' | 'games' | 'announcement' | 'customFiles';
const SECTION_LABELS: Record<SectionKey, { title: string; subtitle: string }> = {
    general: { title: 'General', subtitle: 'Basic server configuration options.' },
    internet: { title: 'Internet', subtitle: 'Internet access and the address the server announces.' },
    games: { title: 'Games', subtitle: 'Configure which games the server accepts and their binding addresses.' },
    announcement: { title: 'Announcement', subtitle: 'Configure LAN discovery and announcement settings.' },
    customFiles: { title: 'Custom files', subtitle: 'Storage of the savegames and replays uploaded by the game.' },
};
const AUTH_OPTIONS = [
    { value: 'disabled', label: 'Disabled', description: 'No authentication as it worked before.' },
    { value: 'required', label: 'Required', description: 'Authenticate the user with the official servers.' },
    {
        value: 'cached',
        label: 'Cached',
        description: 'Authenticate with the official server at least once every 30 days.',
    },
    {
        value: 'adaptive',
        label: 'Adaptive',
        description: "Resolve to 'cached' if internet is detected, or 'disabled' otherwise.",
    },
];
/** Sizes are stored with their suffix, and the server reads MB and GB. */ const SIZE_UNITS: UnitInputUnit[] = [
    { label: 'MB', storedUnit: 'MB' },
    { label: 'GB', storedUnit: 'GB' },
];
/** * The server parses Go durations, which have hours but no days: a day picked * in the dropdown is written as its 24 hours. */ const EXPIRY_UNITS: UnitInputUnit[] =
    [
        { label: 'hours', storedUnit: 'h' },
        { label: 'days', storedUnit: 'h', factor: 24 },
    ];
/** Section that owns a validation error, so the summary can jump to it. */ function sectionForField(
    field: string,
): SectionKey {
    if (field.startsWith('Internet.')) return 'internet';
    if (field.startsWith('Games.')) return 'games';
    if (field.startsWith('Announcement.')) return 'announcement';
    if (field.startsWith('CustomFiles.')) return 'customFiles';
    return 'general';
}
/** Splits the result of the Go validation into blocking errors and warnings. */ function splitIssues(
    issues: ValidationError[] | null | undefined,
): { errors: ValidationErrors; warnings: ValidationErrors } {
    const errors: ValidationErrors = {};
    const warnings: ValidationErrors = {};
    (issues ?? []).forEach((issue) => {
        if (!issue?.field) return;
        if (issue.severity === 'warning') warnings[issue.field] = humanMessage(issue.message);
        else errors[issue.field] = humanMessage(issue.message);
    });
    return { errors, warnings };
}
/** * Keeps the messages that the live pass still reports while it reflects the * current values: one goes away as soon as its field is fine again, without * waiting for another blur. The live pass never brings a message up. */ function currentMessages(
    shown: ValidationErrors,
    live: ValidationErrors,
    liveIsCurrent: boolean,
): ValidationErrors {
    if (!liveIsCurrent) return shown;
    return Object.fromEntries(
        Object.entries(shown)
            .filter(([field]) => live[field])
            .map(([field]) => [field, live[field]]),
    );
}
interface Screen2Props {
    initialConfig: AppConfig;
    backendValidationErrors: ValidationError[];
    fileName: string;
    isNew: boolean;
    onSaved: (file: ConfigFile) => void;
    onBack: () => void;
}
export const Screen2_ConfigEditor: React.FC<Screen2Props> = ({
    initialConfig,
    backendValidationErrors,
    fileName,
    isNew,
    onSaved,
    onBack,
}) => {
    const form = useForm<AppConfig>({ mode: 'controlled', initialValues: initialConfig });
    const config = form.values;
    const [authExpanded, authExpandedHandlers] = useDisclosure(false);
    const [multicastExpanded, multicastExpandedHandlers] = useDisclosure(false); // Multicast sits under the LAN announcement: switching the announcement off
    // takes the whole card away, so its details close instead of staying open
    // behind a row that can no longer be clicked.
    const closeMulticastDetails = multicastExpandedHandlers.close;
    useEffect(() => {
        if (!config.Announcement.Enabled) closeMulticastDetails();
    }, [config.Announcement.Enabled, closeMulticastDetails]);
    const [navbarOpened, navbarHandlers] = useDisclosure(true); // Below the App Shell breakpoint the sidebar becomes an overlay on top of the
    // content instead of a column beside it, so it keeps its own state (closed to
    // begin with) and a surface of its own: transparent, it would let the content
    // show through as it slid over it.
    const [overlayNavbarOpened, overlayNavbarHandlers] = useDisclosure(false);
    const isNarrowWindow = useMediaQuery('(max-width: 47.99375em)'); // Matches the App Shell media query that switches the navbar into an
    // overlay (breakpoint minus a tenth of a pixel).
    const sidebarOpened = isNarrowWindow ? overlayNavbarOpened : navbarOpened;
    const toggleSidebar = isNarrowWindow ? overlayNavbarHandlers.toggle : navbarHandlers.toggle;
    const sidebarWidth = 240;
    const [activeSection, setActiveSection] = useState<SectionKey>('general'); // The file's own validation result seeds the messages shown when it opens;
    // every blur (or committed control) replaces them with the fresh result.
    const [backendIssues] = useState(() => splitIssues(backendValidationErrors));
    const [errors, setErrors] = useState<ValidationErrors>(backendIssues.errors);
    const [warnings, setWarnings] = useState<ValidationErrors>(backendIssues.warnings); // Values that were last validated. They only change when a control loses
    // focus, so neither the field errors nor the general summary react to
    // every keystroke.
    const [validationValues, setValidationValues] = useState<AppConfig>(initialConfig); // Validation rules live in Go (ConfigService.ValidateConfiguration). The
    // frontend only asks Go to evaluate them and renders the resulting errors.
    useEffect(() => {
        let cancelled = false;
        ConfigService.ValidateConfiguration(toServiceConfiguration(validationValues))
            .then((result) => {
                if (cancelled) return;
                const issues = splitIssues(result);
                setErrors(issues.errors);
                setWarnings(issues.warnings);
            })
            .catch(() => {
                // Keep the last known errors if validation cannot be reached.
            });
        return () => {
            cancelled = true;
        };
    }, [validationValues]); // Blur is the moment a value is considered finished, and the same request
    // is made by the actions that edit a list as a whole. The timeout lets the
    // control's own blur handling settle first (NumberInput clamps on blur),
    // so the values that reach Go are the ones the control shows.
    const latestValues = useRef(form.values);
    latestValues.current = form.values;
    const requestValidation = () => {
        window.setTimeout(() => setValidationValues(latestValues.current), 0);
    }; // Save, though, has to know whether the form is valid right now: an invalid
    // value keeps it disabled while it is still being typed, before the blur
    // that brings up the message. This pass runs on a short debounce, and its
    // result is never rendered as an error by itself.
    const [liveIssues, setLiveIssues] = useState(backendIssues); // Values the live result was computed for. The debounce catching up is not
    // an answer yet: until the pass for the values on screen arrives, the last
    // result stays out of the way instead of briefly filtering the messages
    // that are already shown (which reads as the summary flickering).
    const [liveValues, setLiveValues] = useState<AppConfig>(initialConfig);
    const [debouncedValues] = useDebouncedValue(form.values, 200);
    useEffect(() => {
        let cancelled = false;
        ConfigService.ValidateConfiguration(toServiceConfiguration(debouncedValues))
            .then((result) => {
                if (cancelled) return;
                setLiveIssues(splitIssues(result));
                setLiveValues(debouncedValues);
            })
            .catch(() => {
                // Keep the last known state if validation cannot be reached.
            });
        return () => {
            cancelled = true;
        };
    }, [debouncedValues]); // Controls that commit a value by themselves (a switch, a radio) have no blur
    // of their own, so they ask for validation as soon as they change; only text
    // typed inside a field waits until it is left.
    const commitProps = <Props extends { onChange?: (value: any) => void }>(props: Props): Props => ({
        ...props,
        onChange: (value: any) => {
            props.onChange?.(value);
            requestValidation();
        },
    }); // Switch state as the form holds it, with the change marked as finished:
    // SwitchSettingCard renders the control and validates through `onCommit`.
    const switchProps = (path: string) => {
        const props = form.getInputProps(path, { type: 'checkbox' });
        return {
            checked: Boolean(props.checked),
            onChange: (checked: boolean) => props.onChange(checked),
            onCommit: requestValidation,
        };
    }; // While the live pass reflects the values as they are right now, it also
    // decides which messages stay: one goes away as soon as its field is fine
    // again, without waiting for another blur. A message never *appears* this
    // way, only disappears, so nothing shows up mid-typing.
    const liveIsCurrent = liveValues === form.values;
    const allErrors = currentMessages(errors, liveIssues.errors, liveIsCurrent);
    const allWarnings = currentMessages(warnings, liveIssues.warnings, liveIsCurrent);
    const errorEntries = Object.entries(allErrors); // Save follows the live pass: the messages above are a snapshot taken when a
    // control was left, so a value typed right now is not in it yet. Warnings are
    // advisories and never block saving.
    const hasErrors = Object.keys(liveIssues.errors).length > 0;
    const handleGameToggle = (gameId: string) => {
        const enabled = [...form.values.Games.Enabled];
        const index = enabled.indexOf(gameId);
        if (index === -1) {
            enabled.push(gameId);
        } else {
            enabled.splice(index, 1);
        }
        form.setFieldValue('Games.Enabled', enabled);
        requestValidation();
    };
    const handleGameHostChange = (gameId: string, index: number, value: string) => {
        form.setFieldValue(`Games.${gameId}.Hosts.${index}`, value);
    }; // Adding and removing a host change the shape of the list, which is not
    // something that produces a blur: they count as finished edits themselves.
    const handleAddGameHost = (gameId: string) => {
        const hosts = form.values.Games[gameId]?.Hosts ?? [];
        form.setFieldValue(`Games.${gameId}.Hosts`, [...hosts, '0.0.0.0']);
        requestValidation();
    };
    const handleRemoveGameHost = (gameId: string, index: number) => {
        const hosts = [...(form.values.Games[gameId]?.Hosts ?? [])];
        hosts.splice(index, 1);
        form.setFieldValue(`Games.${gameId}.Hosts`, hosts);
        requestValidation();
    };
    const handlePortChange = (value: string | number) => {
        const port = typeof value === 'string' ? Number.parseInt(value, 10) : value;
        if (Number.isNaN(port)) return;
        form.setFieldValue('Announcement.Port', port);
    };
    const showSuccessToast = (message: string) => {
        notifications.show({ message, color: 'accent', icon: <IconCircleCheck size={20} />, autoClose: 4000 });
    };
    const showErrorToast = (message: string) => {
        notifications.show({ message, color: 'red', icon: <IconAlertTriangle size={20} />, autoClose: 6000 });
    };
    const handleSave = async () => {
        if (hasErrors) return;
        try {
            if (!isNew) {
                await ConfigService.SaveConfiguration(fileName, toServiceConfiguration(form.values));
                showSuccessToast('Configuration saved successfully.');
                return;
            }
            const destination = await ConfigService.SelectConfigDestination();
            if (!destination) return;
            try {
                const created = await ConfigService.CreateConfigFile(
                    destination,
                    toServiceConfiguration(form.values),
                    false,
                );
                if (!created) throw new Error('The created configuration was not returned.');
                onSaved(created);
                showSuccessToast('Configuration created successfully.');
            } catch (error) {
                const message = error instanceof Error ? error.message : String(error);
                if (
                    !message.includes('confirmation required') ||
                    !window.confirm('The selected file already exists. Overwrite it?')
                ) {
                    throw error;
                }
                const created = await ConfigService.CreateConfigFile(
                    destination,
                    toServiceConfiguration(form.values),
                    true,
                );
                if (!created) throw new Error('The overwritten configuration was not returned.');
                onSaved(created);
                showSuccessToast('Configuration overwritten successfully.');
            }
        } catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            showErrorToast(`Error saving configuration: ${message}`);
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
                    toServiceConfiguration(form.values),
                    false,
                );
                if (!saved) throw new Error('The saved configuration was not returned.');
                onSaved(saved);
                showSuccessToast('Configuration saved successfully to the new location.');
            } catch (error) {
                const message = error instanceof Error ? error.message : String(error);
                if (
                    !message.includes('confirmation required') ||
                    !window.confirm('The selected file already exists. Overwrite it?')
                ) {
                    throw error;
                }
                const saved = await ConfigService.SaveConfigurationAs(
                    fileName,
                    destination,
                    toServiceConfiguration(form.values),
                    true,
                );
                if (!saved) throw new Error('The overwritten configuration was not returned.');
                onSaved(saved);
                showSuccessToast('Configuration overwritten successfully.');
            }
        } catch (error) {
            const message = error instanceof Error ? error.message : String(error);
            showErrorToast(`Error saving configuration: ${message}`);
        }
    };
    const authLabel = AUTH_OPTIONS.find((option) => option.value === config.Authentication)?.label ?? 'Disabled';
    const navItems: Array<{ key: SectionKey; label: string; icon: React.ReactNode }> = [
        { key: 'general', label: 'General', icon: <IconSettings size={18} stroke={1.6} /> },
        { key: 'internet', label: 'Internet', icon: <IconWorld size={18} stroke={1.6} /> },
        { key: 'games', label: 'Games', icon: <IconDeviceGamepad2 size={18} stroke={1.6} /> },
        { key: 'announcement', label: 'Announcement', icon: <IconSpeakerphone size={18} stroke={1.6} /> },
        { key: 'customFiles', label: 'Custom files', icon: <IconFiles size={18} stroke={1.6} /> },
    ]; // WinUI 3 navigation items: rounded rows inset from the pane edges, with
    // a subtle highlight on hover and when selected instead of an accent
    // fill. The states cannot go through the styles prop (it only supports
    // flat properties), so they live in this small stylesheet.
    const winuiNavItemClass = 'winui-nav-item';
    const winuiNavItemCss = `        .${winuiNavItemClass} [data-position='left'] {            color: var(--mantine-color-dimmed);        }        @media (hover: hover) {            .${winuiNavItemClass}:hover {                background-color: color-mix(in srgb, var(--mantine-color-text) 5%, transparent);            }        }        .${winuiNavItemClass}[data-active] {            background-color: color-mix(in srgb, var(--mantine-color-text) 7%, transparent);            color: var(--mantine-color-text);        }        .${winuiNavItemClass}[data-active] [data-position='left'] {            color: var(--mantine-color-text);        }    `;
    return (
        <AppShell
            header={{ height: 64 }}
            navbar={{
                width: sidebarWidth,
                breakpoint: 'sm',
                collapsed: { mobile: !overlayNavbarOpened, desktop: !navbarOpened },
            }}
            padding="lg" // The window must not scroll as a whole (see AppShell.Main below), so
            // the shell keeps clear of the fixed header with padding: a margin
            // here would collapse out of the shell and leave the window
            // scrollable by exactly the header's height.
            style={{ paddingTop: 'var(--app-shell-header-offset, 0rem)' }}
        >
            <AppShell.Header>
                <Group h="100%" px="md" justify="space-between" wrap="nowrap" gap="md">
                    <Group gap="md" wrap="nowrap" style={{ flex: 1, minWidth: 0 }}>
                        <Burger
                            opened={sidebarOpened}
                            onClick={toggleSidebar}
                            size="sm"
                            aria-label={sidebarOpened ? 'Hide sidebar' : 'Show sidebar'}
                        />
                        <Box style={{ minWidth: 0 }}>
                            <Title order={4}>{SECTION_LABELS[activeSection].title}</Title>
                            <Text size="xs" c="dimmed" truncate>
                                {SECTION_LABELS[activeSection].subtitle}
                            </Text>
                        </Box>
                    </Group>
                    {isNew ? (
                        <Button leftSection={<IconDeviceFloppy size={16} />} onClick={handleSave} disabled={hasErrors}>
                            Save
                        </Button>
                    ) : (
                        <Button.Group>
                            <Button
                                leftSection={<IconDeviceFloppy size={16} />}
                                onClick={handleSave}
                                disabled={hasErrors}
                            >
                                Save
                            </Button>
                            <Menu position="bottom-end" width={190}>
                                <Menu.Target>
                                    <Button px={8} aria-label="More save options" disabled={hasErrors}>
                                        <IconChevronDown size={16} />
                                    </Button>
                                </Menu.Target>
                                <Menu.Dropdown>
                                    <Menu.Item leftSection={<IconDeviceFloppy size={16} />} onClick={handleSaveAs}>
                                        Save as├ö├ç┬¬
                                    </Menu.Item>
                                </Menu.Dropdown>
                            </Menu>
                        </Button.Group>
                    )}
                </Group>
            </AppShell.Header>
            <AppShell.Navbar
                style={
                    isNarrowWindow
                        ? {
                              // As an overlay the sidebar lies on top of the
                              // content, so it cannot be transparent like it is
                              // beside it, and Mantine gives an overlay navbar
                              // the whole window width: a 240 px acrylic pane,
                              // like the one a Windows 11 app slides in, with a
                              // frosted fill that blurs what is behind.
                              width: sidebarWidth,
                              background: 'color-mix(in srgb, var(--mantine-color-default) 90%, transparent)',
                              backdropFilter: 'blur(28px) saturate(150%)',
                              boxShadow: 'var(--mantine-shadow-md)',
                          }
                        : undefined
                }
            >
                <style>{winuiNavItemCss}</style>
                <AppShell.Section grow component={ScrollArea} px={12} py="sm">
                    {navItems.map((item) => (
                        <NavLink
                            key={item.key}
                            label={item.label}
                            leftSection={item.icon}
                            active={activeSection === item.key}
                            onClick={() => setActiveSection(item.key)}
                            className={winuiNavItemClass}
                            styles={{
                                root: {
                                    // WinUI 3 navigation item: rounded. The
                                    // inset comes from the section padding
                                    // (the items are width: 100%, so a margin
                                    // would overflow the pane and bring up a
                                    // horizontal scrollbar).
                                    borderRadius: 4,
                                    paddingBlock: 8,
                                    paddingInline: 8,
                                },
                            }}
                        />
                    ))}
                </AppShell.Section>
                <AppShell.Section px={12} py="xs">
                    <Group gap={8} wrap="nowrap">
                        <Box w={8} h={8} bg="teal" style={{ borderRadius: 999, flexShrink: 0 }} />
                        <Text size="xs" c="dimmed" truncate style={{ flex: 1 }}>
                            {fileName || 'New configuration'}
                        </Text>
                        <Tooltip label="Open another configuration file" withArrow openDelay={400}>
                            <ActionIcon
                                variant="subtle"
                                size="sm"
                                onClick={onBack}
                                aria-label="Open another configuration file"
                            >
                                <IconFolderOpen size={15} stroke={1.6} />
                            </ActionIcon>
                        </Tooltip>
                    </Group>
                </AppShell.Section>
            </AppShell.Navbar>
            <AppShell.Main // Validation is requested when a control in the content gives
                // up focus (see requestValidation).
                onBlur={requestValidation} // The window itself does not scroll: the content is the scroll
                // area, so it stays clipped below the header instead of sliding
                // behind it (the header is transparent and lets the window
                // backdrop through). The header offset is a padding of the shell
                // (see .app-shell in style.css) and is dropped from the padding
                // here, so the content is not offset twice.
                style={
                    {
                        paddingTop: 'var(--app-shell-padding, 1.5rem)',
                        height: 'calc(100dvh - var(--app-shell-header-offset, 4rem))',
                        minHeight: 0,
                        overflowX: 'hidden',
                        overflowY: 'auto',
                    } as React.CSSProperties
                }
            >
                <Stack gap="md">
                    {errorEntries.length > 0 && (
                        <Alert
                            color="yellow"
                            variant="light"
                            icon={<IconAlertTriangle size={20} />}
                            title={`${errorEntries.length} validation ${errorEntries.length === 1 ? 'error' : 'errors'}`}
                        >
                            <List size="sm" spacing={4}>
                                {errorEntries.map(([field, message]) => (
                                    <List.Item key={`${field}-${message}`}>
                                        <UnstyledButton
                                            onClick={() => setActiveSection(sectionForField(field))}
                                            style={{ textAlign: 'left' }}
                                        >
                                            <Text size="sm">
                                                <Text span fw={600}>
                                                    {fieldLabel(field)}
                                                </Text>
                                                : {message}
                                            </Text>
                                        </UnstyledButton>
                                    </List.Item>
                                ))}
                            </List>
                        </Alert>
                    )}
                    {/* SECTION 1: GENERAL */}
                    {activeSection === 'general' && (
                        <Stack gap="sm">
                            <SwitchSettingCard
                                icon={<IconLogs size={20} stroke={1.6} />}
                                title="Log information"
                                description="Enables logging to terminal and saves data to a file for debugging."
                                {...switchProps('Log')}
                            />
                            <SwitchSettingCard
                                icon={<IconId size={20} stroke={1.6} />}
                                title="Generate platform user ID"
                                description="Generate a unique ID. Only if multiple users share a launcher. Incompatible with active authentication."
                                error={allErrors.GeneratePlatformUserId}
                                {...switchProps('GeneratePlatformUserId')}
                            />
                            <SwitchSettingCard
                                icon={<IconSignature size={20} stroke={1.6} />}
                                title="Re-sign catalogues"
                                description="Sign the catalogues with this server's key instead of the captured signatures. Only for clients that trust this server's signing key."
                                error={allErrors.Sign}
                                {...switchProps('Sign')}
                            />
                            <SettingCard
                                icon={<IconShield size={20} stroke={1.6} />}
                                title="Authentication method"
                                description="Define how users are authenticated: required, cached, adaptive, or disabled."
                                value={authLabel}
                                error={allErrors.Authentication}
                                expanded={authExpanded}
                                onToggle={authExpandedHandlers.toggle}
                                details={
                                    <Radio.Group {...commitProps(form.getInputProps('Authentication'))}>
                                        <Stack gap="xs">
                                            {AUTH_OPTIONS.map((option) => (
                                                <Radio
                                                    key={option.value}
                                                    value={option.value}
                                                    label={option.label}
                                                    description={option.description}
                                                />
                                            ))}
                                        </Stack>
                                    </Radio.Group>
                                }
                            />
                        </Stack>
                    )}
                    {/* SECTION 2: INTERNET */}
                    {activeSection === 'internet' && (
                        <Stack gap="sm">
                            <SwitchSettingCard
                                icon={<IconWorld size={20} stroke={1.6} />}
                                title="Internet access"
                                description="Allow the server to use the Internet: connectivity check, authentication and public IP detection."
                                error={allErrors['Internet.Enabled']}
                                {...switchProps('Internet.Enabled')}
                            />
                            <AutoManualCard
                                icon={<IconNetwork size={20} stroke={1.6} />}
                                title="Public IP address"
                                description="IPv4 address other networks use to reach this server. Automatic asks api.ipify.org for it."
                                error={allErrors['Internet.IP']}
                                value={config.Internet.IP}
                                onChange={(value) => form.setFieldValue('Internet.IP', value)}
                                onCommit={requestValidation}
                                disabled={!config.Internet.Enabled}
                            >
                                {(manual) => <IPv4Input {...manual} />}
                            </AutoManualCard>
                        </Stack>
                    )}
                    {/* SECTION 3: GAMES & HOSTS */}
                    {activeSection === 'games' && (
                        <Stack gap="sm">
                            {allWarnings['Games.Enabled'] && (
                                <Group gap={6} wrap="nowrap" align="flex-start">
                                    <IconAlertTriangle
                                        size={16}
                                        stroke={1.8}
                                        style={{ color: 'var(--mantine-color-warning)', flexShrink: 0, marginTop: 2 }}
                                    />
                                    <Text size="sm" style={{ color: 'var(--mantine-color-warning)' }}>
                                        {allWarnings['Games.Enabled']}
                                    </Text>
                                </Group>
                            )}
                            <SimpleGrid cols={{ base: 1, xs: 2 }} spacing="sm">
                                {AVAILABLE_GAMES.map((game) => {
                                    const isEnabled = config.Games.Enabled.includes(game.id);
                                    const gameConf = config.Games[game.id];
                                    const hosts = gameConf && Array.isArray(gameConf.Hosts) ? gameConf.Hosts : []; // A game holds as many hosts as it needs: each row
                                    // carries the message of its own address, and the card
                                    // paints its border for any of them.
                                    const rowErrors = hosts.map(
                                        (_, index) => allErrors[`Games.${game.id}.Hosts.${index}`],
                                    );
                                    // A game that is not enabled only warns about a shared
                                    // address, which does not paint the card.
                                    const rowWarnings = hosts.map(
                                        (_, index) => allWarnings[`Games.${game.id}.Hosts.${index}`],
                                    );
                                    const hostsError = allErrors[`Games.${game.id}.Hosts`];
                                    return (
                                        <SwitchSettingCard
                                            key={game.id}
                                            title={game.name}
                                            description="IPv4 addresses the server binds to for this game."
                                            error={hostsError}
                                            invalid={rowErrors.some((message) => Boolean(message))}
                                            checked={isEnabled}
                                            onChange={() => handleGameToggle(game.id)}
                                            ariaLabel={`Enable ${game.name}`}
                                        >
                                            <Stack gap="xs">
                                                {hosts.map((host, index) => (
                                                    <Stack key={index} gap={4}>
                                                        <Group gap="xs" wrap="nowrap" justify="space-between">
                                                            <IPv4Input
                                                                value={host}
                                                                onChange={(value) =>
                                                                    handleGameHostChange(game.id, index, value)
                                                                }
                                                                disabled={!isEnabled}
                                                            />
                                                            <ActionIcon
                                                                variant="subtle"
                                                                color="gray"
                                                                size="sm" // A game needs a host to bind to, so the
                                                                // last remaining row cannot be removed.
                                                                disabled={!isEnabled || hosts.length <= 1}
                                                                aria-label={`Remove host ${index + 1} of ${game.name}`}
                                                                onClick={() => handleRemoveGameHost(game.id, index)}
                                                            >
                                                                <IconTrash size={16} stroke={1.6} />
                                                            </ActionIcon>
                                                        </Group>
                                                        {rowErrors[index] && (
                                                            <Input.Error>{rowErrors[index]}</Input.Error>
                                                        )}
                                                        {!rowErrors[index] && rowWarnings[index] && (
                                                            <Group gap={6} wrap="nowrap" align="flex-start">
                                                                <IconAlertTriangle
                                                                    size={13}
                                                                    stroke={1.8}
                                                                    style={{
                                                                        color: 'var(--mantine-color-warning)',
                                                                        flexShrink: 0,
                                                                        marginTop: 2,
                                                                    }}
                                                                />
                                                                <Text size="xs" style={{ color: 'var(--mantine-color-warning)' }}>
                                                                    {rowWarnings[index]}
                                                                </Text>
                                                            </Group>
                                                        )}
                                                    </Stack>
                                                ))}
                                                <Button
                                                    variant="subtle"
                                                    size="compact-sm"
                                                    w="fit-content"
                                                    leftSection={<IconPlus size={15} stroke={1.8} />}
                                                    disabled={!isEnabled}
                                                    onClick={() => handleAddGameHost(game.id)}
                                                >
                                                    Add host
                                                </Button>
                                            </Stack>
                                        </SwitchSettingCard>
                                    );
                                })}
                            </SimpleGrid>
                        </Stack>
                    )}
                    {/* SECTION 4: ANNOUNCEMENT */}
                    {activeSection === 'announcement' && (
                        <Stack gap="sm">
                            <SwitchSettingCard
                                icon={<IconSpeakerphone size={20} stroke={1.6} />}
                                title="LAN announcement"
                                description="Respond to automatic discovery queries on the local LAN."
                                error={allErrors['Announcement.Enabled']}
                                {...switchProps('Announcement.Enabled')}
                            />
                            <SettingCard
                                icon={<IconNetwork size={20} stroke={1.6} />}
                                title="Announcement port"
                                description="UDP port for announcing the server to launchers (default: 31978)."
                                error={allErrors['Announcement.Port']}
                                warning={allWarnings['Announcement.Port']}
                                control={
                                    <NumberInput
                                        min={1}
                                        max={65535}
                                        clampBehavior="blur"
                                        value={config.Announcement.Port}
                                        onChange={handlePortChange} // The port only means something while the server
                                        // announces itself on the LAN.
                                        disabled={!config.Announcement.Enabled}
                                        aria-label="Announcement port"
                                        w={140}
                                    />
                                }
                            />
                            <Text size="sm" fw={600} mt="md">
                                Multicast
                            </Text>
                            <SwitchSettingCard
                                icon={<IconBroadcast size={20} stroke={1.6} />}
                                title="Multicast"
                                description="Respond to discovery queries at the multicast address."
                                error={allErrors['Announcement.MulticastGroup']}
                                {...switchProps('Announcement.Multicast')} // Multicast announcements are one way of announcing:
                                // without the plain announcement there is nothing to
                                // send over multicast.
                                disabled={!config.Announcement.Enabled}
                                expanded={multicastExpanded}
                                onToggle={multicastExpandedHandlers.toggle}
                                details={
                                    <Group wrap="nowrap" justify="space-between" gap="md">
                                        <Box style={{ minWidth: 0 }}>
                                            <Text size="sm" fw={500}>
                                                Multicast group
                                            </Text>
                                            <Text size="xs" c="dimmed" mt={2}>
                                                IPv4 multicast group for responding to announcements (default:
                                                239.31.97.8; class D range 224.0.0.0 ├ö├ç├┤ 239.255.255.255).
                                            </Text>
                                        </Box>
                                        <IPv4Input
                                            onChange={(value) =>
                                                form.setFieldValue('Announcement.MulticastGroup', value)
                                            }
                                            value={config.Announcement.MulticastGroup} // The group is only used while multicast is on
                                            // (and multicast needs the announcement).
                                            disabled={!config.Announcement.Enabled || !config.Announcement.Multicast}
                                        />
                                    </Group>
                                }
                            />
                        </Stack>
                    )}
                    {/* SECTION 5: CUSTOM FILES */}
                    {activeSection === 'customFiles' && (
                        <Stack gap="sm">
                            <SwitchSettingCard
                                icon={<IconFiles size={20} stroke={1.6} />}
                                title="Custom files"
                                description="Serve the endpoints the game uses to upload and download multiplayer savegames and replays."
                                error={allErrors['CustomFiles.Enabled']}
                                {...switchProps('CustomFiles.Enabled')}
                            />
                            <AutoManualCard
                                icon={<IconFolder size={20} stroke={1.6} />}
                                title="Storage folder"
                                description="Folder where the files are stored. Automatic uses the system temporary folder."
                                error={allErrors['CustomFiles.Folder']}
                                value={config.CustomFiles.Folder}
                                onChange={(value) => form.setFieldValue('CustomFiles.Folder', value)}
                                onCommit={requestValidation}
                                disabled={!config.CustomFiles.Enabled}
                            >
                                {(manual) => <FolderPickerInput {...manual} onSelect={requestValidation} />}
                            </AutoManualCard>
                            <SettingCard
                                icon={<IconFile size={20} stroke={1.6} />}
                                title="Maximum file size"
                                description="Largest single file accepted, in megabytes or gigabytes (default: 32MB)."
                                error={allErrors['CustomFiles.MaxFileSize']}
                                control={
                                    <UnitInput
                                        value={config.CustomFiles.MaxFileSize}
                                        onChange={(value) => form.setFieldValue('CustomFiles.MaxFileSize', value)}
                                        units={SIZE_UNITS}
                                        ariaLabel="Maximum file size"
                                        onCommit={requestValidation}
                                        disabled={!config.CustomFiles.Enabled}
                                    />
                                }
                            />
                            <SettingCard
                                icon={<IconDatabase size={20} stroke={1.6} />}
                                title="Total size limit"
                                description="Maximum accumulated size per game. When exceeded, the oldest files are evicted first (default: 256MB)."
                                error={allErrors['CustomFiles.MaxTotalSize']}
                                control={
                                    <UnitInput
                                        value={config.CustomFiles.MaxTotalSize}
                                        onChange={(value) => form.setFieldValue('CustomFiles.MaxTotalSize', value)}
                                        units={SIZE_UNITS}
                                        ariaLabel="Total size limit"
                                        onCommit={requestValidation}
                                        disabled={!config.CustomFiles.Enabled}
                                    />
                                }
                            />
                            <SettingCard
                                icon={<IconClock size={20} stroke={1.6} />}
                                title="File expiry"
                                description="How long a file is kept since its upload, in hours or days (default: 12h)."
                                error={allErrors['CustomFiles.Expiry']}
                                control={
                                    <UnitInput
                                        value={config.CustomFiles.Expiry}
                                        onChange={(value) => form.setFieldValue('CustomFiles.Expiry', value)}
                                        units={EXPIRY_UNITS}
                                        ariaLabel="File expiry"
                                        onCommit={requestValidation}
                                        disabled={!config.CustomFiles.Enabled}
                                    />
                                }
                            />
                        </Stack>
                    )}
                </Stack>
            </AppShell.Main>
        </AppShell>
    );
};

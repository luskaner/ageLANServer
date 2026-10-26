import React from 'react';
import { ActionIcon, Box, Button, Group, Input, Stack, Text } from '@mantine/core';
import { useDisclosure } from '@mantine/hooks';
import { IconAlertTriangle, IconPlus, IconServerCog, IconTrash, IconWorld } from '@tabler/icons-react';
import { AppConfig, BattleServerConfig, ValidationErrors } from '../types';
import { BattleServerDialog } from './BattleServerDialog';
import { IPv4Input } from './IPv4Input';
import { SettingCard } from './SettingCard';

export interface GameInfo {
    id: string;
    name: string;
    code: string;
}

/**
 * Summary shown on the right of the battle servers card: the server regions in
 * order, separated by commas. The region is unique within the game, so it is
 * enough to identify each server.
 */
function battleServerSummary(servers: BattleServerConfig[]): string {
    if (servers.length === 0) return 'None';
    return servers.map((server) => server.Region.trim() || '—').join(', ');
}

export interface GameSectionProps {
    game: GameInfo;
    config: AppConfig;
    errors: ValidationErrors;
    warnings: ValidationErrors;
    onHostChange: (gameId: string, index: number, value: string) => void;
    onAddHost: (gameId: string) => void;
    onRemoveHost: (gameId: string, index: number) => void;
    onChangeBattleServers: (gameId: string, servers: BattleServerConfig[]) => void;
    onCommit: () => void;
}

/**
 * One game's page: two collapsible cards. Hosts shows the bound addresses as a
 * comma-separated summary and lets them be edited when expanded; Battle servers
 * lists its regions the same way (they are unique within the game) and
 * expanding lists each server's address/ports with a delete button. Adding a
 * server opens a single-entry dialog.
 */
export const GameSection: React.FC<GameSectionProps> = ({
    game,
    config,
    errors,
    warnings,
    onHostChange,
    onAddHost,
    onRemoveHost,
    onChangeBattleServers,
    onCommit,
}) => {
    const [hostsExpanded, hostsHandlers] = useDisclosure(false);
    const [serversExpanded, serversHandlers] = useDisclosure(false);
    const [addServerOpen, addServerHandlers] = useDisclosure(false);

    const gameConfig = config.Games[game.id];
    const hosts = Array.isArray(gameConfig?.Hosts) ? gameConfig.Hosts : [];
    const servers = Array.isArray(gameConfig?.BattleServers) ? gameConfig.BattleServers : [];

    const hostsError = errors[`Games.${game.id}.Hosts`];
    const hostRowErrors = hosts.map((_, index) => errors[`Games.${game.id}.Hosts.${index}`]);
    const hostRowWarnings = hosts.map((_, index) => warnings[`Games.${game.id}.Hosts.${index}`]);
    const battleServersError = errors[`Games.${game.id}.BattleServers`];
    const battleServerEntryErrors = Object.keys(errors).filter((field) =>
        field.startsWith(`Games.${game.id}.BattleServers.`),
    );

    return (
        <Stack gap="sm">
            {warnings['Games.Enabled'] && (
                <Group gap={6} wrap="nowrap" align="flex-start">
                    <IconAlertTriangle
                        size={16}
                        stroke={1.8}
                        style={{ color: 'var(--mantine-color-warning)', flexShrink: 0, marginTop: 2 }}
                    />
                    <Text size="sm" style={{ color: 'var(--mantine-color-warning)' }}>
                        {warnings['Games.Enabled']}
                    </Text>
                </Group>
            )}

            <SettingCard
                icon={<IconWorld size={20} stroke={1.6} />}
                title="Hosts"
                description="IPv4 addresses the server binds to for this game."
                value={hosts.length > 0 ? hosts.join(', ') : 'None'}
                error={hostsError}
                invalid={hostRowErrors.some((message) => Boolean(message))}
                expanded={hostsExpanded}
                onToggle={hostsHandlers.toggle}
                details={
                    <Stack gap="xs">
                        {hosts.map((host, index) => (
                            <Stack key={index} gap={4}>
                                <Group gap="xs" wrap="nowrap" justify="space-between">
                                    <IPv4Input
                                        value={host}
                                        onChange={(value) => onHostChange(game.id, index, value)}
                                    />
                                    <ActionIcon
                                        variant="subtle"
                                        color="gray"
                                        size="sm"
                                        aria-label={`Remove host ${index + 1} of ${game.name}`}
                                        // A game needs a host to bind to, so the last row stays.
                                        disabled={hosts.length <= 1}
                                        onClick={() => onRemoveHost(game.id, index)}
                                    >
                                        <IconTrash size={16} stroke={1.6} />
                                    </ActionIcon>
                                </Group>
                                {hostRowErrors[index] && <Input.Error>{hostRowErrors[index]}</Input.Error>}
                                {!hostRowErrors[index] && hostRowWarnings[index] && (
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
                                            {hostRowWarnings[index]}
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
                            onClick={() => onAddHost(game.id)}
                        >
                            Add host
                        </Button>
                    </Stack>
                }
            />

            <SettingCard
                icon={<IconServerCog size={20} stroke={1.6} />}
                title="Battle servers"
                description="Battle servers the server announces for this game."
                value={battleServerSummary(servers)}
                error={battleServersError}
                invalid={battleServerEntryErrors.length > 0}
                expanded={serversExpanded}
                onToggle={serversHandlers.toggle}
                details={
                    <Stack gap="sm">
                        {servers.length === 0 && (
                            <Text size="sm" c="dimmed">
                                No battle servers configured for this game.
                            </Text>
                        )}
                        {servers.map((server, index) => {
                            const prefix = `Games.${game.id}.BattleServers.${index}`;
                            const entryErrors = Object.entries(errors)
                                .filter(([field]) => field.startsWith(`${prefix}.`))
                                .map(([, message]) => message);
                            return (
                                <Group key={index} gap="sm" wrap="nowrap" justify="space-between" align="flex-start">
                                    <Box style={{ minWidth: 0 }}>
                                        <Text size="sm" fw={500}>
                                            {server.Region || `Battle server ${index + 1}`}
                                            {server.Name ? ` — ${server.Name}` : ''}
                                        </Text>
                                        <Text size="xs" c="dimmed">
                                            IPv4: {server.IPv4} · Battle server port: {server.BsPort} · WebSocket
                                            port: {server.WebSocketPort}
                                            {game.id !== 'age1'
                                                ? ` · Out-of-band port: ${server.OutOfBandPort}`
                                                : ''}
                                        </Text>
                                        {entryErrors.length > 0 && (
                                            <Input.Error>{entryErrors.join(' · ')}</Input.Error>
                                        )}
                                    </Box>
                                    <ActionIcon
                                        variant="subtle"
                                        color="red"
                                        size="sm"
                                        aria-label={`Remove battle server ${index + 1} of ${game.name}`}
                                        onClick={() => {
                                            const next = [...servers];
                                            next.splice(index, 1);
                                            onChangeBattleServers(game.id, next);
                                            onCommit();
                                        }}
                                    >
                                        <IconTrash size={16} stroke={1.6} />
                                    </ActionIcon>
                                </Group>
                            );
                        })}
                        <Button
                            variant="subtle"
                            size="compact-sm"
                            w="fit-content"
                            leftSection={<IconPlus size={15} stroke={1.8} />}
                            onClick={addServerHandlers.open}
                        >
                            Add battle server
                        </Button>
                    </Stack>
                }
            />

            <BattleServerDialog
                opened={addServerOpen}
                gameId={game.id}
                gameName={game.name}
                servers={servers}
                onAdd={(server) => {
                    onChangeBattleServers(game.id, [...servers, server]);
                    onCommit();
                    addServerHandlers.close();
                }}
                onClose={addServerHandlers.close}
            />
        </Stack>
    );
};

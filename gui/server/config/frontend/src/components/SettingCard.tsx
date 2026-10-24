import React from 'react';
import { Box, Card, Collapse, Group, Input, Text, ThemeIcon, UnstyledButton } from '@mantine/core';
import { IconAlertTriangle, IconChevronDown } from '@tabler/icons-react';

export interface SettingCardProps {
    /** Icon shown in the leading ThemeIcon box. */
    icon?: React.ReactNode;
    title: string;
    description?: string;
    /** Control rendered at the end of the row (a Switch, an input, ...). */
    control?: React.ReactNode;
    /** Current value shown just before the control (used by expandable rows). */
    value?: string;
    /** Extra content always shown under the row. */
    children?: React.ReactNode;
    /** Makes the whole row clickable; clicking toggles `details`. */
    onToggle?: () => void;
    /** Whether `details` is currently revealed. */
    expanded?: boolean;
    /** Content revealed under the row while `expanded`. */
    details?: React.ReactNode;
    /** Validation message: paints the card border red and renders the message below the row. */
    error?: string;
    /** Marks the card invalid when the message sits inside it, next to the row it belongs to. */
    invalid?: boolean;
    /** Advisory message: renders below the row without marking the card as invalid. */
    warning?: string;
    /** Disabled cards keep their details closed and cannot be expanded. */
    disabled?: boolean;
}

/**
 * Translucent card for a single setting over the window's Mica backdrop: icon,
 * title/description and a control on the right. Rows can be expandable
 * (`onToggle` + `details`), hold extra content (`children`) and show a
 * card-level validation `error`.
 */
export const SettingCard: React.FC<SettingCardProps> = ({
    icon,
    title,
    description,
    control,
    value,
    children,
    onToggle,
    expanded = false,
    details,
    error,
    invalid,
    warning,
    disabled = false,
}) => {
    // A disabled setting has nothing to reveal: its row no longer toggles the
    // details and the chevron is gone, so the collapsible cannot be opened.
    const expandable = Boolean(onToggle) && !disabled;
    const open = expandable && expanded;
    const row = (
        <Group wrap="nowrap" justify="space-between" gap="md" px="md" py="sm">
            <Group wrap="nowrap" gap="md" style={{ flex: 1, minWidth: 0 }}>
                {icon && (
                    <ThemeIcon variant="light" size={38} radius="md">
                        {icon}
                    </ThemeIcon>
                )}
                <Box style={{ minWidth: 0 }}>
                    <Text size="sm" fw={500}>
                        {title}
                    </Text>
                    {description && (
                        <Text size="xs" c="dimmed" mt={2}>
                            {description}
                        </Text>
                    )}
                </Box>
            </Group>
            <Group gap="sm" wrap="nowrap">
                {value && (
                    <Text size="sm">
                        {value}
                    </Text>
                )}
                {control}
                {expandable && (
                    <IconChevronDown
                        size={16}
                        stroke={1.6}
                        style={{
                            flexShrink: 0,
                            transition: 'transform 150ms ease',
                            transform: open ? 'rotate(180deg)' : undefined,
                        }}
                    />
                )}
            </Group>
        </Group>
    );

    return (
        <Card
            withBorder
            radius="md"
            p={0}
            className="app-card"
            style={{
                backgroundColor: 'var(--app-card-bg, rgba(255, 255, 255, 0.06))',
                borderColor: error || invalid
                    ? 'var(--mantine-color-error)'
                    : 'var(--app-card-border, transparent)',
            }}
        >
            {expandable ? (
                <UnstyledButton className="app-card-row" w="100%" onClick={onToggle} aria-expanded={open}>
                    {row}
                </UnstyledButton>
            ) : (
                row
            )}
            {error && (
                <Input.Error
                    pb="sm"
                    my={0}
                    pr="md"
                    pl={icon ? 'calc(38px + 2 * var(--mantine-spacing-md))' : 'md'}
                >
                    {error}
                </Input.Error>
            )}
            {warning && (
                <Group
                    gap={6}
                    wrap="nowrap"
                    align="flex-start"
                    pb="sm"
                    pr="md"
                    pl={icon ? 'calc(38px + 2 * var(--mantine-spacing-md))' : 'md'}
                >
                    <IconAlertTriangle
                        size={14}
                        stroke={1.8}
                        style={{ color: 'var(--mantine-color-warning)', flexShrink: 0, marginTop: 2 }}
                    />
                    <Text size="xs" style={{ color: 'var(--mantine-color-warning)' }}>
                        {warning}
                    </Text>
                </Group>
            )}
            {children && (
                <Box px="md" pb="sm">
                    {children}
                </Box>
            )}
            {details && (
                <Collapse expanded={open}>
                    <Box pt="sm" pb="md" pr="md" pl="calc(38px + 2 * var(--mantine-spacing-md))">
                        {details}
                    </Box>
                </Collapse>
            )}
        </Card>
    );
};

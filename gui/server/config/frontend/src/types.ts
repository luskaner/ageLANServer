export interface GameHostsConfig {
  Hosts: string[];
}

export interface GamesConfig {
  Enabled: string[];
  age1?: GameHostsConfig;
  age2?: GameHostsConfig;
  age3?: GameHostsConfig;
  age4?: GameHostsConfig;
  athens?: GameHostsConfig;
  [key: string]: any;
}

export interface AnnouncementConfig {
  Enabled: boolean;
  Multicast: boolean;
  Port: number;
  MulticastGroup: string;
}

export interface InternetConfig {
  Enabled: boolean;
  IP: string;
}

export interface CustomFilesConfig {
  Enabled: boolean;
  Folder: string;
  MaxFileSize: string;
  MaxTotalSize: string;
  Expiry: string;
}

export interface AppConfig {
  Log: boolean;
  GeneratePlatformUserId: boolean;
  Sign: boolean;
  Authentication: 'required' | 'cached' | 'adaptive' | 'disabled';
  Internet: InternetConfig;
  Games: GamesConfig;
  Announcement: AnnouncementConfig;
  CustomFiles: CustomFilesConfig;
}

export interface ValidationErrors {
  [key: string]: string | undefined;
}

export const AVAILABLE_GAMES = [
  { id: 'age1', name: 'Age of Empires: DE', code: 'AoE: DE' },
  { id: 'age2', name: 'Age of Empires II: DE', code: 'AoE II: DE' },
  { id: 'age3', name: 'Age of Empires III: DE', code: 'AoE III: DE' },
  { id: 'age4', name: 'Age of Empires IV: AE', code: 'AoE IV: AE' },
  { id: 'athens', name: 'Age of Mythology: Retold', code: 'AoM: RT' },
];

const GAME_NAMES = new Map(AVAILABLE_GAMES.map((game) => [game.id, game.name]));
const GAME_ID_PATTERN = new RegExp(`\\b(?:${AVAILABLE_GAMES.map((game) => game.id).join('|')})\\b`, 'g');

/**
 * Label of a validation field as the form shows it: the card the user sees,
 * never the configuration key or the list position behind it
 * (`Announcement.Port`, `Games.age2.Hosts.0`). Go keeps those keys: they are
 * what the editor uses to map every answer back to its control.
 */
export function fieldLabel(field: string): string {
  if (field === 'Authentication') return 'Authentication method';
  if (field === 'GeneratePlatformUserId') return 'Generate platform user ID';
  if (field === 'Sign') return 'Re-sign catalogues';
  if (field === 'Internet.Enabled') return 'Internet access';
  if (field === 'Internet.IP') return 'Public IP address';
  if (field === 'Announcement.Port') return 'Announcement port';
  if (field === 'Announcement.MulticastGroup') return 'Multicast group';
  if (field === 'CustomFiles.Enabled') return 'Custom files';
  if (field === 'CustomFiles.Folder') return 'Custom files folder';
  if (field === 'CustomFiles.MaxFileSize') return 'Maximum file size';
  if (field === 'CustomFiles.MaxTotalSize') return 'Total size limit';
  if (field === 'CustomFiles.Expiry') return 'File expiry';
  if (field.startsWith('Games.Enabled')) return 'Enabled games';

  const hosts = /^Games\.([^.]+)\.Hosts(?:\.\d+)?$/.exec(field);
  if (hosts) return `Hosts of ${GAME_NAMES.get(hosts[1]) ?? hosts[1]}`;

  // A rule pointing at something this map does not know yet: the name of its
  // last step, spread out into words, so nothing technical reaches the user.
  const words = field.split('.').filter((segment) => !/^\d+$/.test(segment));
  return (words[words.length - 1] ?? field).replace(/([a-z0-9])([A-Z])/g, '$1 $2');
}

/**
 * A Go validation message as it reads to the user: a game is named the way the
 * form names it, instead of by its identifier in the configuration file.
 */
export function humanMessage(message: string): string {
  return message.replace(GAME_ID_PATTERN, (id) => GAME_NAMES.get(id) ?? id);
}

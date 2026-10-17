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

export interface AppConfig {
  Log: boolean;
  GeneratePlatformUserId: boolean;
  Authentication: 'required' | 'cached' | 'adaptive' | 'disabled';
  Games: GamesConfig;
  Announcement: AnnouncementConfig;
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

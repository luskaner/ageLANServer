import { AppConfig } from './types';

export function parseTOML(tomlText: string): AppConfig {
  const config: AppConfig = {
    Log: false,
    GeneratePlatformUserId: false,
    Authentication: 'disabled',
    Games: {
      Enabled: [],
      age1: { Hosts: ['0.0.0.0'] },
      age2: { Hosts: ['0.0.0.0'] },
      age3: { Hosts: ['0.0.0.0'] },
      age4: { Hosts: ['0.0.0.0'] },
      athens: { Hosts: ['0.0.0.0'] }
    },
    Announcement: {
      Enabled: true,
      Multicast: true,
      Port: 31978,
      MulticastGroup: '239.31.97.8'
    }
  };

  const lines = tomlText.split(/\r?\n/);
  let currentSection = '';

  for (let line of lines) {
    line = line.trim();
    if (!line || line.startsWith('#')) continue;

    const hashIndex = line.indexOf('#');
    let content = line;
    if (hashIndex !== -1) {
      const beforeHash = line.substring(0, hashIndex);
      const quoteCount = (beforeHash.match(/['"]/g) || []).length;
      if (quoteCount % 2 === 0) {
        content = beforeHash.trim();
      }
    }

    if (content.startsWith('[') && content.endsWith(']')) {
      currentSection = content.slice(1, -1).trim();
      continue;
    }

    const eqIndex = content.indexOf('=');
    if (eqIndex === -1) continue;

    const key = content.substring(0, eqIndex).trim();
    const rawVal = content.substring(eqIndex + 1).trim();

    const parseValue = (valStr: string) => {
      if (valStr === 'true') return true;
      if (valStr === 'false') return false;
      if (!isNaN(Number(valStr)) && valStr !== '') return Number(valStr);
      if ((valStr.startsWith("'") && valStr.endsWith("'")) || (valStr.startsWith('"') && valStr.endsWith('"'))) {
        return valStr.slice(1, -1);
      }
      if (valStr.startsWith('[') && valStr.endsWith(']')) {
        const inner = valStr.slice(1, -1).trim();
        if (!inner) return [];
        return inner.split(',').map(item => {
          let s = item.trim();
          if ((s.startsWith("'") && s.endsWith("'")) || (s.startsWith('"') && s.endsWith('"'))) {
            s = s.slice(1, -1);
          }
          return s;
        }).filter(Boolean);
      }
      return valStr;
    };

    const val = parseValue(rawVal);

    if (currentSection === '') {
      if (key === 'Log') config.Log = Boolean(val);
      else if (key === 'GeneratePlatformUserId') config.GeneratePlatformUserId = Boolean(val);
      else if (key === 'Authentication') config.Authentication = val as any;
    } else if (currentSection === 'Games') {
      if (key === 'Enabled' && Array.isArray(val)) {
        config.Games.Enabled = val;
      }
    } else if (currentSection.startsWith('Games.')) {
      const gameKey = currentSection.replace('Games.', '');
      if (!config.Games[gameKey]) {
        config.Games[gameKey] = { Hosts: [] };
      }
      if (key === 'Hosts' && Array.isArray(val)) {
        config.Games[gameKey].Hosts = val;
      }
    } else if (currentSection === 'Announcement') {
      if (key === 'Enabled') config.Announcement.Enabled = Boolean(val);
      else if (key === 'Multicast') config.Announcement.Multicast = Boolean(val);
      else if (key === 'Port') config.Announcement.Port = Number(val);
      else if (key === 'MulticastGroup') config.Announcement.MulticastGroup = String(val);
    }
  }

  return config;
}

export function stringifyTOML(config: AppConfig): string {
  let output = `# ageLANServer configuration file

Log = ${config.Log}
GeneratePlatformUserId = ${config.GeneratePlatformUserId}
Authentication = '${config.Authentication}'

[Games]
Enabled = [${config.Games.Enabled.map(g => `'${g}'`).join(', ')}]

`;

  const allGames = ['age1', 'age2', 'age3', 'age4', 'athens'];
  allGames.forEach(gameKey => {
    const gameConf = config.Games[gameKey];
    const hosts = gameConf && Array.isArray(gameConf.Hosts) ? gameConf.Hosts : ['0.0.0.0'];
    output += `[Games.${gameKey}]\nHosts = [${hosts.map(h => `'${h}'`).join(', ')}]\n\n`;
  });

  output += `[Announcement]
Enabled = ${config.Announcement.Enabled}
Multicast = ${config.Announcement.Multicast}
Port = ${config.Announcement.Port}
MulticastGroup = '${config.Announcement.MulticastGroup}'
`;

  return output;
}

export function validateIPv4(ip: string): boolean {
  if (!ip) return false;
  const ipv4Regex = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
  return ipv4Regex.test(ip.trim());
}

export function validateMulticastIPv4(ip: string): boolean {
  if (!validateIPv4(ip)) return false;
  const parts = ip.trim().split('.').map(Number);
  // IPv4 Multicast range is Class D: 224.0.0.0 to 239.255.255.255
  return parts[0] >= 224 && parts[0] <= 239;
}

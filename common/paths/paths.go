package paths

import "path/filepath"

const ResourcesDir = "resources"
const ConfigDir = "config"

var ConfigsPath = filepath.Join(ResourcesDir, ConfigDir)

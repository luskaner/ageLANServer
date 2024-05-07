package userData

import "os"

type Data struct {
	Path     string
	Original bool
}

const finalPath = `Games\Age of Empires 2 DE`

func (d *Data) backupPath() string {
	return d.absolutePath() + `.bak`
}

func (d *Data) temporaryPath() string {
	return d.absolutePath() + `.tmp`
}

func (d *Data) absolutePath() string {
	return Path() + `\` + d.Path
}

func Path() string {
	return os.Getenv("USERPROFILE") + `\` + finalPath
}

func (d *Data) srcDestPath() (string, string) {
	if d.Original {
		return d.absolutePath(), d.backupPath()
	} else {
		return d.backupPath(), d.absolutePath()
	}
}

func (d *Data) switchPath() bool {
	srcPath, destPath := d.srcDestPath()
	tempPath := d.temporaryPath()
	err := os.Rename(srcPath, tempPath)
	if err != nil {
		return false
	}
	err = os.Rename(destPath, srcPath)
	if err != nil {
		return false
	}
	err = os.Rename(tempPath, destPath)
	if err != nil {
		return false
	}
	d.Original = !d.Original
	return true
}

func (d *Data) Backup() bool {
	if !d.Original {
		return true
	}
	absolutePath := d.absolutePath()
	info, err := os.Stat(absolutePath)

	if err != nil {
		return false
	}

	backupPath := d.backupPath()

	if _, err := os.Stat(backupPath); err != nil {
		err = os.Rename(absolutePath, backupPath)
		if err != nil {
			return false
		}
		err = os.Mkdir(absolutePath, info.Mode())
		if err != nil {
			return false
		}
		d.Original = false
		return true
	} else {
		return d.switchPath()
	}
}

func (d *Data) Restore() bool {
	if d.Original {
		return true
	}

	absolutePath := d.absolutePath()
	_, err := os.Stat(absolutePath)

	if err != nil {
		return false
	}

	backupPath := d.backupPath()

	if _, err := os.Stat(backupPath); err != nil {
		return false
	}

	return d.switchPath()
}

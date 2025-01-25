package whitelist

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type Whitelist struct {
	FilePath string
	ModTime  time.Time
	whitelistData
}

func (wa Whitelist) IsModified() bool {
	if fileInfo, err := os.Stat(wa.FilePath); err != nil {
		return false
	} else {
		return fileInfo.ModTime().After(wa.ModTime)
	}
}

func NewWhitelist(filePath string) (*Whitelist, error) {
	_, err := os.Stat(filePath)
	wa := &Whitelist{}
	wa.FilePath = filePath

	if err != nil {
		if pathErr, ok := err.(*os.PathError); ok && os.IsNotExist(pathErr) {
			if err := wa.DumpFile(); err != nil {
				log.Printf("Failed to dump file: %v", err)
				return nil, err
			}
		} else {
			log.Printf("Failed to stat file: %v", err)
			return nil, err
		}
	} else {
		if err := wa.LoadFile(); err != nil {
			return nil, err
		}
	}

	return wa, nil
}

func (wa *Whitelist) AddUser(userId int64) error {
	if wa.IsModified() {
		log.Printf("Whitelist file %s has been modified", wa.FilePath)
		if err := wa.LoadFile(); err != nil {
			return err
		}
	}
	wa.whitelistData.addUser(userId)
	return wa.DumpFile()
}

func (wa *Whitelist) AddGroup(groupId int64) error {
	if wa.IsModified() {
		log.Printf("Whitelist file %s has been modified", wa.FilePath)
		if err := wa.LoadFile(); err != nil {
			return err
		}
	}
	wa.whitelistData.addGroup(groupId)
	return wa.DumpFile()
}

func (wa Whitelist) Show() (*string, error) {
	if wa.IsModified() {
		log.Printf("Whitelist file %s has been modified", wa.FilePath)
		if err := wa.LoadFile(); err != nil {
			return nil, err
		}
	}

	b, err := json.Marshal(wa.whitelistData)
	if err != nil {
		return nil, err
	}
	s := string(b)
	return &s, nil
}

func (wa *Whitelist) DumpFile() error {
	if err := wa.whitelistData.DumpFile(wa.FilePath); err != nil {
		return err
	}

	if fileInfo, err := os.Stat(wa.FilePath); err != nil {
		return err
	} else {
		wa.ModTime = fileInfo.ModTime()
	}

	return nil
}

func (wa *Whitelist) LoadFile() error {
	whitelist, err := loadFile(wa.FilePath)
	if err != nil {
		return err
	}

	fileInfo, err := os.Stat(wa.FilePath)
	if err != nil {
		return err
	}

	wa.ModTime = fileInfo.ModTime()
	wa.whitelistData = *whitelist

	return nil
}

func (wa *Whitelist) HasGroup(groupId int64) bool {
	if wa.IsModified() {
		log.Printf("Whitelist file %s has been modified", wa.FilePath)
		if err := wa.LoadFile(); err != nil {
			log.Printf("Failed to load file: %v", err)
			return false
		}
	}
	return wa.whitelistData.hasGroup(groupId)
}

func (wa *Whitelist) HasUser(userId int64) bool {
	if wa.IsModified() {
		log.Printf("Whitelist file %s has been modified", wa.FilePath)
		if err := wa.LoadFile(); err != nil {
			log.Printf("Failed to load file: %v", err)
			return false
		}
	}
	return wa.whitelistData.hasUser(userId)
}

func (wa Whitelist) IsAdmin(userId int64) bool {
	return wa.whitelistData.isAdmin(userId)
}

type whitelistData struct {
	UserIds  []int64 `json:"user_ids"`
	GroupIds []int64 `json:"group_ids"`
	Admin    *int64  `json:"admin,omitempty"`
}

func (wd whitelistData) DumpFile(path string) error {
	bytes, err := json.Marshal(wd)
	if err != nil {
		return nil
	}
	return os.WriteFile(path, bytes, 0644)
}

func loadFile(path string) (*whitelistData, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	whitelist := &whitelistData{}
	err = json.Unmarshal(bytes, whitelist)
	if err != nil {
		return nil, err
	}

	return whitelist, nil
}

func (wd whitelistData) hasUser(userId int64) bool {
	for _, n := range wd.UserIds {
		if n == userId {
			return true
		}
	}
	return false
}

func (wd whitelistData) hasGroup(groupId int64) bool {
	for _, n := range wd.GroupIds {
		if n == groupId {
			return true
		}
	}
	return false
}

func (wd *whitelistData) addUser(userId int64) {
	wd.UserIds = append(wd.UserIds, userId)
}

func (wd *whitelistData) addGroup(groupId int64) {
	wd.GroupIds = append(wd.GroupIds, groupId)
}

func (wd whitelistData) isAdmin(userId int64) bool {
	return wd.Admin != nil && *wd.Admin == userId
}

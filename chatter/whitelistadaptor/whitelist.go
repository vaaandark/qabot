package whitelistadaptor

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type WhitelistAdaptor struct {
	FilePath string
	ModTime  time.Time
	whitelist
}

func (wa WhitelistAdaptor) IsModified() bool {
	if fileInfo, err := os.Stat(wa.FilePath); err != nil {
		return false
	} else {
		return fileInfo.ModTime().After(wa.ModTime)
	}
}

func NewWhitelistAdaptor(filePath string) (*WhitelistAdaptor, error) {
	_, err := os.Stat(filePath)
	wa := &WhitelistAdaptor{}
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

func (wa *WhitelistAdaptor) AddUser(userId int64) error {
	wa.whitelist.addUser(userId)
	return wa.DumpFile()
}

func (wa *WhitelistAdaptor) AddGroup(groupId int64) error {
	wa.whitelist.addGroup(groupId)
	return wa.DumpFile()
}

func (wa WhitelistAdaptor) Show() (*string, error) {
	b, err := json.Marshal(wa.whitelist)
	if err != nil {
		return nil, err
	}
	s := string(b)
	return &s, nil
}

func (wa *WhitelistAdaptor) DumpFile() error {
	if err := wa.whitelist.DumpFile(wa.FilePath); err != nil {
		return err
	}

	if fileInfo, err := os.Stat(wa.FilePath); err != nil {
		return err
	} else {
		wa.ModTime = fileInfo.ModTime()
	}

	return nil
}

func (wa *WhitelistAdaptor) LoadFile() error {
	whitelist, err := loadFile(wa.FilePath)
	if err != nil {
		return err
	}

	fileInfo, err := os.Stat(wa.FilePath)
	if err != nil {
		return err
	}

	wa.ModTime = fileInfo.ModTime()
	wa.whitelist = *whitelist

	return nil
}

func (wa *WhitelistAdaptor) HasGroup(groupId int64) bool {
	if wa.IsModified() {
		log.Print("Whitelist file has been modified")
		if err := wa.LoadFile(); err != nil {
			log.Printf("Failed to load file: %v", err)
			return false
		}
	}
	return wa.whitelist.hasGroup(groupId)
}

func (wa *WhitelistAdaptor) HasUser(userId int64) bool {
	if wa.IsModified() {
		log.Print("Whitelist file has been modified")
		if err := wa.LoadFile(); err != nil {
			log.Printf("Failed to load file: %v", err)
			return false
		}
	}
	return wa.whitelist.hasUser(userId)
}

func (wa WhitelistAdaptor) IsAdmin(userId int64) bool {
	return wa.whitelist.isAdmin(userId)
}

type whitelist struct {
	UserIds  []int64 `json:"user_ids"`
	GroupIds []int64 `json:"group_ids"`
	Admin    *int64  `json:"admin,omitempty"`
}

func (w whitelist) DumpFile(path string) error {
	bytes, err := json.Marshal(w)
	if err != nil {
		return nil
	}
	return os.WriteFile(path, bytes, 0644)
}

func loadFile(path string) (*whitelist, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	whitelist := &whitelist{}
	err = json.Unmarshal(bytes, whitelist)
	if err != nil {
		return nil, err
	}

	return whitelist, nil
}

func (w whitelist) hasUser(userId int64) bool {
	for _, n := range w.UserIds {
		if n == userId {
			return true
		}
	}
	return false
}

func (w whitelist) hasGroup(groupId int64) bool {
	for _, n := range w.GroupIds {
		if n == groupId {
			return true
		}
	}
	return false
}

func (w *whitelist) addUser(userId int64) {
	w.UserIds = append(w.UserIds, userId)
}

func (w *whitelist) addGroup(groupId int64) {
	w.GroupIds = append(w.GroupIds, groupId)
}

func (w whitelist) isAdmin(userId int64) bool {
	return w.Admin != nil && *w.Admin == userId
}

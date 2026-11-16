package repo

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm/clause"
)

// Platform-wide interface language. Empty lets every user switch between
// Chinese and English from the top bar; "en" or "zh" locks the interface to
// that language for everyone and hides the switch.
const (
	InterfaceLanguageUserChoice = ""
	InterfaceLanguageEnglish    = "en"
	InterfaceLanguageChinese    = "zh"
)

var ErrInvalidInterfaceLanguage = errors.New("interface language must be empty, en or zh")

// NormalizeInterfaceLanguage trims and validates a platform language value.
func NormalizeInterfaceLanguage(raw string) (string, bool) {
	switch lang := strings.ToLower(strings.TrimSpace(raw)); lang {
	case InterfaceLanguageUserChoice, InterfaceLanguageEnglish, InterfaceLanguageChinese:
		return lang, true
	default:
		return "", false
	}
}

type platformSettingsRow struct {
	ID                int64      `gorm:"column:id;primaryKey;autoIncrement:false"`
	InterfaceLanguage string     `gorm:"column:interface_language;size:16;not null;default:''"`
	UpdatedBy         int64      `gorm:"column:updated_by;not null;default:0"`
	Updated           *time.Time `gorm:"column:updated_at"`
}

func (platformSettingsRow) TableName() string { return "platform_settings" }

// The table holds at most this one row.
const platformSettingsID = 1

// PlatformInterfaceLanguageStatus is what administrators see.
type PlatformInterfaceLanguageStatus struct {
	Language  string     `json:"language"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
	UpdatedBy string     `json:"updatedBy,omitempty"`
}

func (r *UsersRepo) platformSettings() (platformSettingsRow, bool) {
	var row platformSettingsRow
	if r == nil || r.deps == nil || r.DB == nil {
		return row, false
	}
	result := r.DB.Where("id = ?", platformSettingsID).Limit(1).Find(&row)
	return row, result.Error == nil && result.RowsAffected > 0
}

// PlatformInterfaceLanguage returns the locked language, or "" when users
// choose their own. Unknown stored values fall back to the user's choice.
func (r *UsersRepo) PlatformInterfaceLanguage() string {
	row, ok := r.platformSettings()
	if !ok {
		return InterfaceLanguageUserChoice
	}
	lang, valid := NormalizeInterfaceLanguage(row.InterfaceLanguage)
	if !valid {
		return InterfaceLanguageUserChoice
	}
	return lang
}

// PlatformInterfaceLanguageStatus describes the setting for administrators.
func (r *UsersRepo) PlatformInterfaceLanguageStatus() PlatformInterfaceLanguageStatus {
	status := PlatformInterfaceLanguageStatus{Language: r.PlatformInterfaceLanguage()}
	if row, ok := r.platformSettings(); ok {
		status.UpdatedAt = row.Updated
		if user, ok := r.User(row.UpdatedBy); ok {
			status.UpdatedBy = user.LoginName
		}
	}
	return status
}

// SetPlatformInterfaceLanguage saves the platform-wide interface language.
func (r *UsersRepo) SetPlatformInterfaceLanguage(raw string, actorID int64) error {
	lang, ok := NormalizeInterfaceLanguage(raw)
	if !ok {
		return ErrInvalidInterfaceLanguage
	}
	now := time.Now().UTC()
	row := platformSettingsRow{ID: platformSettingsID, InterfaceLanguage: lang, UpdatedBy: actorID, Updated: &now}
	return r.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"interface_language", "updated_by", "updated_at"}),
	}).Create(&row).Error
}

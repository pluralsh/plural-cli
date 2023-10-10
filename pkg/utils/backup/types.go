package backup

type BackupOptions any

const (
	backupsDir = "backups"
)

type Backup[T BackupOptions] interface {
	Exists() bool
	Save(opts T) error
	Restore(opts T) error
	Remove() error
}

package test_package

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/tern/v2/migrate"
	"os"
)

const versionTable = "db_version"

type Migrator struct {
	migrator *migrate.Migrator
}

// NewMigrator создаёт новый экземпляр Migrator, используя указанный путь к файлам миграций.
func NewMigrator(dbDNS string, migrationPath string) (Migrator, error) {
	conn, err := pgx.Connect(context.Background(), dbDNS)
	if err != nil {
		return Migrator{}, err
	}
	migrator, err := migrate.NewMigratorEx(
		context.Background(), conn, versionTable,
		&migrate.MigratorOptions{
			DisableTx: false,
		})
	if err != nil {
		return Migrator{}, err
	}

	// Проверяем, что директория с миграциями существует
	if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
		return Migrator{}, fmt.Errorf("migration directory does not exist: %s", migrationPath)
	}

	// Загружаем миграции из директории
	migrationRoot := os.DirFS(migrationPath)
	err = migrator.LoadMigrations(migrationRoot)
	if err != nil {
		return Migrator{}, err
	}

	return Migrator{
		migrator: migrator,
	}, nil
}

// Info возвращает текущую версию миграции, максимальную миграцию и текстовое представление состояния миграций.
func (m Migrator) Info() (int32, int32, string, error) {
	version, err := m.migrator.GetCurrentVersion(context.Background())
	if err != nil {
		return 0, 0, "", err
	}
	info := ""

	var last int32
	for _, thisMigration := range m.migrator.Migrations {
		last = thisMigration.Sequence

		cur := version == thisMigration.Sequence
		indicator := "  "
		if cur {
			indicator = "->"
		}
		info = info + fmt.Sprintf(
			"%2s %3d %s\n",
			indicator,
			thisMigration.Sequence, thisMigration.Name)
	}

	return version, last, info, nil
}

// Migrate выполняет миграцию базы данных до самой последней версии схемы.
func (m Migrator) Migrate() error {
	err := m.migrator.Migrate(context.Background())
	return err
}

// MigrateTo выполняет миграцию до указанной версии схемы. Используйте '0', чтобы отменить все миграции.
func (m Migrator) MigrateTo(ver int32) error {
	err := m.migrator.MigrateTo(context.Background(), ver)
	return err
}

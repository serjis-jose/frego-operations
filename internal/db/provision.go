package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	operationsProvisionMu      sync.Mutex
	operationsProvisionApplied bool

	operationsProvisionLoadOnce sync.Once
	operationsProvisionSQL      string
	operationsProvisionLoadErr  error
)

func loadOperationsProvisionScript() (string, error) {
	operationsProvisionLoadOnce.Do(func() {
		candidates := []string{
			filepath.Join("db", "provision_tenant.sql"),
			filepath.Join("..", "db", "provision_tenant.sql"),
			filepath.Join("/", "app", "db", "provision_tenant.sql"),
		}
		for _, candidate := range candidates {
			data, err := os.ReadFile(candidate)
			if err == nil {
				operationsProvisionSQL = string(data)
				operationsProvisionLoadErr = nil
				return
			}
			operationsProvisionLoadErr = err
		}
	})

	if operationsProvisionLoadErr != nil {
		return "", fmt.Errorf("load operations provisioning script: %w", operationsProvisionLoadErr)
	}
	if operationsProvisionSQL == "" {
		return "", fmt.Errorf("operations provisioning script is empty")
	}
	return operationsProvisionSQL, nil
}

// EnsureOperationsTenantProvisioning installs or refreshes the operations tenant provisioning procedure on startup.
func EnsureOperationsTenantProvisioning(ctx context.Context, pool *pgxpool.Pool) error {
	operationsProvisionMu.Lock()
	defer operationsProvisionMu.Unlock()

	if operationsProvisionApplied {
		return nil
	}

	script, err := loadOperationsProvisionScript()
	if err != nil {
		return err
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	// Execute the script unconditionally so the latest definition is always in place.
	if _, err := conn.Exec(ctx, script); err != nil {
		return fmt.Errorf("execute provisioning script: %w", err)
	}

	operationsProvisionApplied = true
	return nil
}

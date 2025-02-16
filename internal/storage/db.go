package storage

import (
	"avito/internal/config"
	"fmt"
	"strings"
)

func GetDatabaseDSN(config config.Config) string {
	if config.PostgresConn != "" {
		return config.PostgresConn
	} else if config.PostgresJDBCUrl != "" {
		result, _ := jdbcToGoConnectionString(config.PostgresJDBCUrl, config.PostgresUser, config.PostgresPass)
		return result
	} else {
		var sb strings.Builder
		sb.WriteString("postgres://")
		sb.WriteString(config.PostgresUser)
		sb.WriteString(":")
		sb.WriteString(config.PostgresPass)
		sb.WriteString("@")
		sb.WriteString(config.PostgresHost)
		sb.WriteString(":")
		sb.WriteString(config.PostgresPort)
		sb.WriteString("/")
		sb.WriteString(config.PostgresDB)
		sb.WriteString("?sslmode=disable")
		result := sb.String()
		return result
	}
}

func jdbcToGoConnectionString(jdbc string, username, password string) (string, error) {
	if !strings.HasPrefix(jdbc, "jdbc:postgresql://") {
		return "", fmt.Errorf("invalid JDBC string: must start with jdbc:postgresql://")
	}

	jdbc = strings.TrimPrefix(jdbc, "jdbc:postgresql://")

	goConnStr := fmt.Sprintf("postgres://%s:%s@%s", username, password, jdbc)

	if !strings.Contains(goConnStr, "?") {
		goConnStr += "?sslmode=disable"
	} else if !strings.Contains(goConnStr, "sslmode=") {
		goConnStr += "&sslmode=disable"
	}

	return goConnStr, nil
}

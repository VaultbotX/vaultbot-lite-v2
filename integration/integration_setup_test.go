package integration

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"testing"
)

var (
	c *testcontainers.Container
)

func setup() {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:17.4",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForLog("database system is ready to accept connections"),
	}
	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		panic(err)
	}

	c = &postgresContainer
}

func teardown() {
	//testcontainers.CleanupContainer(t, *c)
	if c != nil {
		ctx := context.Background()
		if err := (*c).Terminate(ctx); err != nil {
			panic(err) // Handle the error appropriately
		}
	}
}

func TestMain(m *testing.M) {
	setup()
	exitCode := m.Run()
	teardown()
	os.Exit(exitCode)
}

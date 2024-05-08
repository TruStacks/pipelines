package main

import (
	"context"
	"fmt"
	"main/internal/dagger"
)

type TrustacksGolangApi struct {
	Source     *Directory
	SonarToken *Secret
	Packages   string
}

const golangVersion = "1.20"

func New(
	// application source path
	source *Directory,

	// sonarqube token
	sonarToken *Secret,
) *TrustacksGolangApi {
	return &TrustacksGolangApi{
		Source:     source,
		SonarToken: sonarToken,
	}
}

func (m *TrustacksGolangApi) Build() error {
	ctx := context.Background()
	if _, err := m.GolangCilint(ctx).Sync(ctx); err != nil {
		return err
	}
	testCtr, err := m.GoTest(ctx).Sync(ctx)
	if err != nil {
		return err
	}
	m.Source = m.Source.WithFile("coverage.out", testCtr.File("/src/coverage.out"))
	if _, err := m.SonarAnalyze(ctx); err != nil {
		return err
	}
	build, err := m.GoBuild(ctx)
	if err != nil {
		return err
	}
	if _, err := build.Sync(ctx); err != nil {
		return err
	}
	return nil
}

func (m *TrustacksGolangApi) GolangCilint(ctx context.Context) *Container {
	return dag.
		GolangciLint().
		Run(m.Source)
}

func (m *TrustacksGolangApi) GoTest(ctx context.Context) *Container {
	return dag.
		Go().
		FromVersion(golangVersion).
		Test(m.Source, GoTestOpts{
			Verbose: true,
			TestFlags: []string{
				"-short",
				"-coverprofile=coverage.out",
			},
		})
}

func (m *TrustacksGolangApi) GoBuild(ctx context.Context) (*Directory, error) {
	packages := []string{}
	commands, err := m.Source.Entries(context.TODO(), dagger.DirectoryEntriesOpts{
		Path: "cmd",
	})
	if err != nil {
		return nil, err
	}
	for _, cmd := range commands {
		packages = append(packages, fmt.Sprintf("./cmd/%s", cmd))
	}
	return dag.
		Go().
		FromVersion(golangVersion).
		Build(m.Source, GoBuildOpts{
			Packages: packages,
		}), nil
}

func (m *TrustacksGolangApi) SonarAnalyze(ctx context.Context) (string, error) {
	return dag.
		Sonar().
		Analyze(ctx, m.Source, m.SonarToken)
}

package main

import (
	"context"
	"strings"
)

type TrustacksGolangApi struct {
	Source   *Directory
	Packages string
}

const golangVersion = "1.20"

func New(
	// application source path
	source *Directory,

	// go packages to build.
	//+default="./cmd"
	//+optional
	packages string,
) *TrustacksGolangApi {
	return &TrustacksGolangApi{
		Source:   source,
		Packages: packages,
	}
}

func (m *TrustacksGolangApi) Build() error {
	ctx := context.Background()
	if _, err := m.GolangCilint().Sync(ctx); err != nil {
		return err
	}
	if _, err := m.GoTest().Sync(ctx); err != nil {
		return err
	}
	if _, err := m.GoBuild().Sync(ctx); err != nil {
		return err
	}
	return nil
}

func (m *TrustacksGolangApi) GolangCilint() *Container {
	return dag.
		GolangciLint().
		Run(m.Source)
}

func (m *TrustacksGolangApi) GoTest() *Container {
	return dag.
		Go().
		FromVersion(golangVersion).
		Test(m.Source, GoTestOpts{Verbose: true, TestFlags: []string{"-short"}})
}

func (m *TrustacksGolangApi) GoBuild() *Directory {
	return dag.
		Go().
		FromVersion(golangVersion).
		Build(m.Source, GoBuildOpts{
			Packages: strings.Split(m.Packages, ","),
		})
}

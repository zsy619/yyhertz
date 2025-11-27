package main

import (
	"fmt"
	"runtime"
	"testing"
)

func TestGetVersionInfo(t *testing.T) {
	info := GetVersionInfo()

	if info.Framework != FrameworkName {

		t.Errorf("Expected framework name '%s', got '%s'", FrameworkName, info.Framework)
	}

	if info.Version != FrameworkVersion {
		t.Errorf("Expected version '%s', got '%s'", FrameworkVersion, info.Version)
	}

	if info.BuildDate != BuildDate {
		t.Errorf("Expected build date '%s', got '%s'", BuildDate, info.BuildDate)
	}

	if info.GoVersion != runtime.Version() {
		t.Errorf("Expected Go version '%s', got '%s'", runtime.Version(), info.GoVersion)
	}

	if info.Platform != runtime.GOOS {
		t.Errorf("Expected platform '%s', got '%s'", runtime.GOOS, info.Platform)
	}

	if info.Arch != runtime.GOARCH {
		t.Errorf("Expected architecture '%s', got '%s'", runtime.GOARCH, info.Arch)
	}

	if info.Author != Author {
		t.Errorf("Expected author '%s', got '%s'", Author, info.Author)
	}

	if info.License != License {
		t.Errorf("Expected license '%s', got '%s'", License, info.License)
	}

	if info.Repository != Repository {
		t.Errorf("Expected repository '%s', got '%s'", Repository, info.Repository)
	}

	if info.Homepage != Homepage {
		t.Errorf("Expected homepage '%s', got '%s'", Homepage, info.Homepage)
	}
}

func TestGetVersionString(t *testing.T) {
	versionString := GetVersionString()
	expected := fmt.Sprintf("%s %s", FrameworkName, FrameworkVersion)

	if versionString != expected {
		t.Errorf("Expected version string '%s', got '%s'", expected, versionString)
	}
}

func TestGetBuildInfo(t *testing.T) {
	buildInfo := GetBuildInfo()
	expected := fmt.Sprintf("%s %s (built with %s on %s/%s at %s)",
		FrameworkName,
		FrameworkVersion,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
		BuildDate,
	)

	if buildInfo != expected {
		t.Errorf("Expected build info '%s', got '%s'", expected, buildInfo)
	}
}

func TestGetFeatures(t *testing.T) {
	features := GetFeatures()
	if len(features) == 0 {
		t.Error("Expected non-empty features list, got empty")
	}
}

func TestGetSystemInfo(t *testing.T) {
	sysInfo := GetSystemInfo()

	if _, ok := sysInfo["go_version"]; !ok {
		t.Error("Expected 'go_version' in system info, not found")
	}

	if _, ok := sysInfo["go_os"]; !ok {
		t.Error("Expected 'go_os' in system info, not found")
	}

	if _, ok := sysInfo["go_arch"]; !ok {
		t.Error("Expected 'go_arch' in system info, not found")
	}

	if _, ok := sysInfo["cpu_count"]; !ok {
		t.Error("Expected 'cpu_count' in system info, not found")
	}

	if _, ok := sysInfo["goroutine_count"]; !ok {
		t.Error("Expected 'goroutine_count' in system info, not found")
	}

	if _, ok := sysInfo["memory_usage"]; !ok {
		t.Error("Expected 'memory_usage' in system info, not found")
	}

	if _, ok := sysInfo["framework"]; !ok {
		t.Error("Expected 'framework' in system info, not found")
	}
}

func TestIsDebugMode(t *testing.T) {
	isDebug := IsDebugMode()
	if isDebug && BuildMode != "debug" {
		t.Error("Expected debug mode to be false, got true")
	}
}

func TestCheckDependencies(t *testing.T) {
	result := CheckDependencies()
	if !result {
		t.Error("Expected dependencies to be compatible, got false")
	}
}
